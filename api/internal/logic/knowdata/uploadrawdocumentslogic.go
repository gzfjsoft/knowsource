package knowdata

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/otiai10/copy"
	"gopkg.in/yaml.v3"

	"knowsource/api/internal/config"
	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/api/internal/utils/random"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func MoveFile(src, dst string) error {
	// 第一步：复制源文件到目标路径
	if err := copy.Copy(src, dst); err != nil {
		return fmt.Errorf("复制失败：%w", err)
	}

	// 第二步：复制成功后，删除源文件
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("删除源文件失败：%w", err)
	}

	return nil
}

// validateFileType validates if the file extension is allowed for the given file type
func validateFileType(ext, fileType string) error {
	allowedExts, exists := constants.AllowedFileTypes[fileType]
	if !exists {
		return errors.New("不支持的文件类型")
	}

	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			return nil
		}
	}

	return errors.New("不支持的文件扩展名: " + ext)
}

type UploadRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上传原始文档，如果zip,自动解压
func NewUploadRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadRawDocumentsLogic {
	return &UploadRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadRawDocumentsLogic) UploadRawDocuments(req *types.UploadRawDocumentsRequest, file *multipart.File) (resp *types.ListRawDocumentsResp, err error) {
	ext := strings.ToLower(filepath.Ext(req.FileName))

	err = validateFileType(ext, constants.FileTypeRawDocuments)
	if err != nil {
		return nil, err
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return nil, errors.New("clientId不能为空，请重新登录")
	}

	// Validate document type exists
	_, err = l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.DocumentCode)
	if err != nil {
		return nil, errors.New("获取文档类型失败,类型不存在: " + req.DocumentCode)
	}

	uuidkey, zipInfo, err := saveUploadRawFile(*file, req.FileName, ext, l.svcCtx.Config)
	if err != nil {
		return nil, err
	}

	// 创建目标目录
	rawDocumentsDir := l.svcCtx.Config.Knowdata.DocumentPath
	//  "../files/raw_documents/"
	tempDir := filepath.Join(rawDocumentsDir, clientId, req.DocumentCode)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录%s失败: %v", tempDir, err)
	}

	// 扫描解压目录下的文件
	tempUploadDir := filepath.Join(l.svcCtx.Config.Knowdata.TempFilePath, uuidkey)
	files, err := scanExtractedFiles(tempUploadDir, tempDir)
	if err != nil {
		return nil, err
	}

	// 根据是否是zip文件设置ZipFileName和ZipFileSize
	var zipFileName string
	var zipFileSize int64
	if ext == ".zip" {
		// 如果是zip文件，使用zip文件信息
		zipFileName = zipInfo.FileName
		zipFileSize = zipInfo.FileSize
	} else {
		// 如果不是zip文件，设置为空和0
		zipFileName = ""
		zipFileSize = 0
	}

	var insertedFiles []types.RawDocuments
	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, s sqlx.Session) error {
		rawDocumentsModel := l.svcCtx.RawDocumentsModel.WithSession(s)

		for _, file := range files {
			fileExt := strings.ToLower(filepath.Ext(file.Name))
			var finalFile FileInfo = file
			var mdFile *FileInfo = nil

			fileMD5, err := calculateFileMD5(file.Path)
			if err != nil {
				l.Logger.Errorf("计算文件 MD5 失败: %v", err)
				return fmt.Errorf("计算文件 MD5 失败: %v", err)
			}
			file.MD5 = fileMD5

			// 如果是 xlsx 文件，转换为 md
			var mdContent string
			if fileExt == ".xlsx" {
				// 获取原文件所在目录和基础名称
				originalFileDir := filepath.Dir(file.Path)
				baseNameWithoutExt := strings.TrimSuffix(file.Name, ".xlsx")

				// 创建 .file 目录（例如：测试xls转MD.xlsx.file）
				mdDirName := file.Name + ".file"
				mdDir := filepath.Join(originalFileDir, mdDirName)

				// 确保目录存在
				if err := os.MkdirAll(mdDir, 0755); err != nil {
					l.Logger.Errorf("创建 MD 文件目录失败: %v, 目录: %s", err, mdDir)
					return fmt.Errorf("创建 MD 文件目录失败: %v", err)
				}

				// 生成 md 文件路径（在 .file 目录中，文件名为 xxx.md）
				mdFileName := baseNameWithoutExt + ".md"
				mdPath := filepath.Join(mdDir, mdFileName)

				// 转换 xlsx 到 md
				if err := utils.ConvertXlsxToMd(file.Path, mdPath); err != nil {
					l.Logger.Errorf("转换 xlsx 到 md 失败: %v, 文件: %s", err, file.Path)
					return fmt.Errorf("转换 xlsx 到 md 失败: %v", err)
				}

				// 读取 md 文件内容
				contentBytes, err := os.ReadFile(mdPath)
				if err != nil {
					l.Logger.Errorf("读取 md 文件内容失败: %v, 文件: %s", err, mdPath)
					return fmt.Errorf("读取 md 文件内容失败: %v", err)
				}
				mdContent = string(contentBytes)

				// 计算 md 文件的 MD5 和大小
				mdMD5, err := calculateFileMD5(mdPath)
				if err != nil {
					l.Logger.Errorf("计算 md 文件 MD5 失败: %v", err)
					return fmt.Errorf("计算 md 文件 MD5 失败: %v", err)
				}

				mdSize := getFileSize(mdPath)

				// 创建 md 文件信息
				mdFile = &FileInfo{
					Name: mdFileName,
					Path: mdPath,
					Size: mdSize,
					MD5:  mdMD5,
				}

			} else if fileExt == ".md" || fileExt == ".txt" {
				// .md 和 .txt 直接读取内容
				contentBytes, err := os.ReadFile(file.Path)
				if err != nil {
					l.Logger.Errorf("读取文件内容失败: %v, 文件: %s", err, file.Path)
					return fmt.Errorf("读取文件内容失败: %v", err)
				}
				mdContent = string(contentBytes)
			}

			// Check if MD5 already exists for this document type
			existingByMD5, err := rawDocumentsModel.FindByMD5AndDocumentCode(ctx, clientId, finalFile.MD5, req.DocumentCode)
			if err == nil && existingByMD5 != nil {
				return fmt.Errorf("文件MD5已存在: 文件 %s 的MD5值 %s 在文档类型 %s 中已存在", finalFile.Name, finalFile.MD5, req.DocumentCode)
			}
			if err != nil && err != model.ErrNotFound {
				l.Logger.Errorf("检查MD5失败: %v", err)
				return fmt.Errorf("检查MD5失败: %v", err)
			}

			// Check if filename already exists for this document type
			existingByFileName, err := rawDocumentsModel.FindByFileNameAndDocumentCode(ctx, clientId, finalFile.Name, req.DocumentCode)
			if err == nil && existingByFileName != nil {
				return fmt.Errorf("文件名已存在: 文件 %s 在文档类型 %s 中已存在。请取消审核并删除原有文件，然后再重新上传", finalFile.Name, req.DocumentCode)
			}
			if err != nil && err != model.ErrNotFound {
				l.Logger.Errorf("检查文件名失败: %v", err)
				return fmt.Errorf("检查文件名失败: %v", err)
			}

			// xlsx 转换的 md、以及直接上传的 .md / .txt 均视为已是 MD 文本，isToMd 设为 1
			isToMd := int64(0)
			if mdFile != nil || fileExt == ".md" || fileExt == ".txt" {
				isToMd = 1
			}

			// xlsx / .md / .txt 设置 file_list，便于“更新内容”时写回对应文件
			var fileList []string
			if mdFile != nil {
				fileList = []string{mdFile.Name}
			} else if fileExt == ".md" || fileExt == ".txt" {
				fileList = []string{finalFile.Name}
			}
			fileListStr := ""
			if len(fileList) > 0 {
				if b, _ := json.Marshal(fileList); len(b) > 0 {
					fileListStr = string(b)
				}
			}

			// 同时设置 content 和 content_org（首次上传时两者相同）
			// 状态约定：
			// - 直接得到 Markdown 内容的（.md/.txt 或 xlsx->md）：已提取文字未审核入库
			// - 需要异步提取文字的（如 pdf/doc/docx）：已上传，正在提取文字...
			status := constants.RawDocumentsStatusUploadedExtracting
			if mdFile != nil || fileExt == ".md" || fileExt == ".txt" || isToMd == 1 {
				status = constants.RawDocumentsStatusExtractedNotInDB
			}
			data := &model.RawDocuments{
				ClientId:      clientId,
				DocumentCode:  req.DocumentCode,
				FileMd5:       finalFile.MD5,
				FileName:      finalFile.Name, // 对于 xlsx，保持原始文件名
				FilePath:      finalFile.Path, // 路径指向 md 文件（对于 xlsx）
				FileSize:      finalFile.Size,
				Content:       mdContent,
				ContentOrg:    mdContent, // 首次上传时，原始内容与当前内容相同
				FileList:      fileListStr,
				Tag:           req.Tag,
				ZipFileName:   zipFileName,
				ZipFileSize:   zipFileSize,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				IsAudit:       0,
				IsToMd:        isToMd,
				IsToAi:        0,
				UploadUser:    l.ctx.Value("userName").(string), //TODOTODO
				UploadEmpcode: l.ctx.Value("empCode").(string),
				AuditUser:     "",
				AuditedAt:     sql.NullTime{Valid: false},
				Status:        status,
			}

			intRes, er := rawDocumentsModel.Insert(ctx, data)
			if er != nil {
				l.Logger.Errorf("保存原始文档文件失败: %v", er)
				return errors.New("保存原始文档文件失败")
			}

			// 创建 meta.yaml 文件
			var tagFilePath string
			if mdFile != nil && fileExt == ".xlsx" {
				// 如果是 xlsx 转换的 md，meta 文件名为 xxx.xlsx.meta.yaml（在原文件所在目录）
				originalFileDir := filepath.Dir(file.Path)
				metaFileName := file.Name + ".meta.yaml"
				tagFilePath = filepath.Join(originalFileDir, metaFileName)
			} else {
				// 其他文件使用原来的格式
				tagFilePath = finalFile.Path + ".meta.yaml"
			}

			// 创建 YAML 结构
			metaData := map[string]interface{}{
				"tag": req.Tag,
			}

			// 将数据转换为 YAML 格式
			yamlData, err := yaml.Marshal(metaData)
			if err != nil {
				l.Logger.Errorf("转换YAML格式失败: %v, 文件路径: %s", err, tagFilePath)
				// 不中断流程，只记录错误
			} else {
				// 写入 YAML 文件
				err = os.WriteFile(tagFilePath, yamlData, 0644)
				if err != nil {
					l.Logger.Errorf("写入tag文件失败: %v, 文件路径: %s", err, tagFilePath)
					// 不中断流程，只记录错误
				}
			}

			fileId, _ := intRes.LastInsertId()
			var auditedAt int64 = 0
			if data.AuditedAt.Valid {
				auditedAt = data.AuditedAt.Time.Unix()
			}
			insertedFiles = append(insertedFiles, types.RawDocuments{
				Id:            fileId,
				DocumentCode:  req.DocumentCode,
				FileMd5:       finalFile.MD5,
				FileName:      finalFile.Name, // 使用保存的文件名（对于 xlsx 保持原始文件名）
				FilePath:      finalFile.Path,
				FileSize:      finalFile.Size,
				Content:       mdContent,
				Tag:           req.Tag,
				ZipFileName:   zipFileName,
				ZipFileSize:   zipFileSize,
				CreatedAt:     data.CreatedAt.Unix(),
				UpdatedAt:     data.UpdatedAt.Unix(),
				IsAudit:       0,
				IsToMd:        isToMd,
				IsToAi:        0,
				UploadUser:    data.UploadUser,
				UploadEmpcode: data.UploadEmpcode,
				AuditUser:     "",
				AuditedAt:     auditedAt,
				Status:        data.Status,
			})

			// 需要异步识别文字的文档：创建 async_task，由统一 worker 处理（避免 goroutine 与 worker 双轨）
			if isToMd == 0 {
				if enqErr := knowsourceLogic.EnqueueRawDocumentConvertZIP(ctx, l.svcCtx, clientId, fileId, finalFile.Name); enqErr != nil {
					l.Logger.Errorf("上传后创建识别任务失败, id=%d, err=%v", fileId, enqErr)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 设置 Redis 标志，表示有新PDF
	if l.svcCtx.RedisClient != nil {
		if err := l.svcCtx.RedisClient.Set("have_new_pdf", "1"); err != nil {
			l.Logger.Errorf("设置Redis have_new_pdf标志失败: %v", err)
			// 不中断流程，只记录错误
		} else {
			l.Logger.Infof("成功设置Redis have_new_pdf=1")
		}
	}

	resp = &types.ListRawDocumentsResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "上传成功",
		},
		Data: types.ListRawDocumentsData{
			List:  insertedFiles,
			Total: int64(len(insertedFiles)),
		},
	}
	return
}

func saveUploadRawFile(file io.Reader, fileName, ext string, config config.Config) (string, ZipInfo, error) {
	key := random.GetUUID()
	zipInfo := ZipInfo{}

	// 创建临时目录
	tempDir := filepath.Join(config.Knowdata.TempFilePath, key)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", zipInfo, fmt.Errorf("创建临时目录失败: %v", err)
	}

	// 保存上传的文件到临时目录
	savePath := filepath.Join(tempDir, fileName)
	dst, err := os.Create(savePath)
	if err != nil {
		os.RemoveAll(tempDir) // 清理临时目录
		return "", zipInfo, fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.RemoveAll(tempDir) // 清理临时目录
		return "", zipInfo, fmt.Errorf("保存文件失败: %v", err)
	}

	// 处理ZIP文件
	if ext == ".zip" {
		// 在解压和删除之前，先获取zip文件信息
		fileInfo, err := os.Stat(savePath)
		if err != nil {
			os.RemoveAll(tempDir)
			return "", zipInfo, fmt.Errorf("获取ZIP文件信息失败: %v", err)
		}

		zipInfo.FileName = fileName
		zipInfo.FileSize = fileInfo.Size()

		// 解压ZIP文件到当前目录（tempDir）
		if err := utils.Unzip(savePath, tempDir); err != nil {
			os.RemoveAll(tempDir)
			return "", zipInfo, fmt.Errorf("解压ZIP文件失败: %v", err)
		}

		// 删除原始ZIP文件
		if err := os.Remove(savePath); err != nil {
			logx.Infof("警告: 删除临时ZIP文件失败: %v", err)
		}
	}

	return key, zipInfo, nil
}

// FileInfo represents information about a file
type FileInfo struct {
	Name string
	Path string
	Size int64
	MD5  string
}

// ZipInfo represents information about a zip file
type ZipInfo struct {
	FileName string
	FileSize int64
}

// scanExtractedFiles scans the extracted files and moves them to the target directory
func scanExtractedFiles(sourceDir, targetDir string) ([]FileInfo, error) {
	var files []FileInfo

	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files and system files
		if strings.HasPrefix(filepath.Base(path), ".") || strings.Contains(path, "__MACOSX") {
			return nil
		}

		// Get relative path from source directory
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Create target path
		targetPath := filepath.Join(targetDir, relPath)

		// Ensure target directory exists
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("创建目标目录失败: %v", err)
		}

		// Calculate MD5 for the file before moving (calculate on source path)
		fileMD5, err := calculateFileMD5(path)
		if err != nil {
			return fmt.Errorf("计算文件MD5失败: %v", err)
		}

		// Move file to target directory ,avoid "invalid cross-device link"
		if err := MoveFile(path, targetPath); err != nil {
			return fmt.Errorf("移动文件失败: %v, %s - %s", err, path, targetPath)
		}
		// if err := os.Rename(path, targetPath); err != nil {
		// 	return fmt.Errorf("移动文件失败: %v, %s -> %s", err, path, targetPath)
		// }
		//

		files = append(files, FileInfo{
			Name: filepath.Base(targetPath),
			Path: targetPath,
			Size: info.Size(),
			MD5:  fileMD5,
		})

		return nil
	})

	return files, err
}

// getFileSize returns the size of a file
func getFileSize(filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}

// calculateFileMD5 calculates the MD5 hash of a file
func calculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
