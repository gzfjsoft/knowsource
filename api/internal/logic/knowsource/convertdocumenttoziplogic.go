package knowsource

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConvertDocumentToZIPLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// convert document to zip
func NewConvertDocumentToZIPLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConvertDocumentToZIPLogic {
	return &ConvertDocumentToZIPLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConvertDocumentToZIPLogic) ConvertDocumentToZIP(req *types.ConvertDocumentToMDRequest) (resp *types.ConvertDocumentToZIPResponse, err error) {
	// 异步提交转换任务（识别文字）；实际转换由 async_task worker 完成。
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}
	if req.Id <= 0 {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "id不能为空",
			},
		}, nil
	}

	rawDoc, findErr := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if findErr != nil {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询文档失败",
				Info:    findErr.Error(),
			},
		}, nil
	}
	currentStatus := strings.TrimSpace(rawDoc.Status)
	allowReExtract := currentStatus == constants.RawDocumentsStatusExtractedNotInDB || currentStatus == ""
	if rawDoc.IsToMd == 1 && !allowReExtract {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "文档已识别",
			},
			Data: &types.ConvertDocumentToZIPData{
				Id:       rawDoc.Id,
				FileName: rawDoc.FileName,
				FileList: []string{},
			},
		}, nil
	}

	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	active, aErr := taskModel.FindActiveByTaskTypeAndSourceId(l.ctx, clientId, constants.AsyncTaskTypeRawDocumentsConvertZIP, rawDoc.Id)
	if aErr != nil {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询识别任务失败",
				Info:    aErr.Error(),
			},
		}, nil
	}
	if active != nil {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "正在提取文字...",
			},
			Data: &types.ConvertDocumentToZIPData{
				Id:       rawDoc.Id,
				FileName: rawDoc.FileName,
				FileList: []string{},
			},
		}, nil
	}

	if cErr := EnqueueRawDocumentConvertZIP(l.ctx, l.svcCtx, clientId, rawDoc.Id, rawDoc.FileName); cErr != nil {
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "创建识别任务失败",
				Info:    cErr.Error(),
			},
		}, nil
	}

	return &types.ConvertDocumentToZIPResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "已提交识别任务，后台处理中",
		},
		Data: &types.ConvertDocumentToZIPData{
			Id:       rawDoc.Id,
			FileName: rawDoc.FileName,
			FileList: []string{},
		},
	}, nil
}

// ConvertDocumentToZIPSync 同步执行转换（供异步任务执行器调用）
func (l *ConvertDocumentToZIPLogic) ConvertDocumentToZIPSync(req *types.ConvertDocumentToMDRequest, asyncTaskId int64) (resp *types.ConvertDocumentToZIPResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)

	// 获取文件根目录
	filesRoot := l.svcCtx.Config.Knowdata.DocumentPath
	if filesRoot == "" {
		filesRoot = l.svcCtx.Config.FilesRoot
	}

	// 查询文档
	var rawDoc *model.RawDocuments
	if clientId != "" {
		rawDoc, err = l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	} else {
		rawDoc, err = l.svcCtx.RawDocumentsModel.FindOne(l.ctx, req.Id)
	}
	if err != nil {
		l.Errorf("查询文档失败: %v", err)
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询文档失败",
				Info:    err.Error(),
			},
		}, nil
	}

	if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
		return nil, cancelErr
	}

	// 提取文字开始（无论 doc/pdf/xlsx 等，统一进入“正在提取文字...”）
	if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
		return nil, cancelErr
	}
	_ = MarkRawDocumentExtracting(l.ctx, l.svcCtx, clientId, rawDoc.Id)

	// 确定原文件的完整路径
	// 先尝试直接使用 FilePath（可能是完整路径）
	var originalFilePath string
	originalFilePath = filepath.Clean(rawDoc.FilePath)
	if _, err = os.Stat(originalFilePath); err != nil {
		// 文件不存在，尝试拼接 FilesRoot
		originalFilePath = filepath.Clean(filepath.Join(filesRoot, rawDoc.FilePath))
		if _, err = os.Stat(originalFilePath); err != nil {
			// 两种方式都失败，返回错误
			l.Errorf("原文件不存在: 尝试路径1=%s, 尝试路径2=%s", filepath.Clean(rawDoc.FilePath), originalFilePath)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "原文件不存在" + rawDoc.FilePath,
					Info:    err.Error() + filepath.Clean(rawDoc.FilePath) + "," + originalFilePath,
				},
			}, nil
		}
	}

	// 获取原文件所在目录和文件扩展名
	originalFileDir := filepath.Dir(originalFilePath)
	baseName := filepath.Base(rawDoc.FileName)
	ext := strings.ToLower(filepath.Ext(baseName))
	baseNameWithoutExt := baseName
	if ext != "" {
		baseNameWithoutExt = baseName[:len(baseName)-len(ext)]
	}

	// 根据文件类型进行不同的处理
	var zipPath string
	var extractedDir string
	var fileList []string

	// 如果是 md 或 txt 文件，不转换但写入 file_list（单文件），便于 UpdateRawDocumentsContent 按 file_list 找文件
	if ext == ".md" || ext == ".txt" {
		fileList = []string{rawDoc.FileName}
		fileListStr, _ := json.Marshal(fileList)
		rawDoc.FileList = string(fileListStr)
		rawDoc.Status = constants.RawDocumentsStatusExtractedNotInDB
		if updateErr := l.svcCtx.RawDocumentsModel.Update(l.ctx, rawDoc); updateErr != nil {
			l.Errorf("更新 rawdoc file_list 失败: %v, ID: %d", updateErr, rawDoc.Id)
		} else {
			l.Infof("已写入 file_list (单文件), ID: %d", rawDoc.Id)
		}
		return &types.ConvertDocumentToZIPResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "MD 和 TXT 文件不需要转换",
			},
		}, nil
	}

	// 生成输出 ZIP 文件路径（保存到原文件同一目录）
	var zipFileName string
	if ext != "" && len(ext) > 1 {
		zipFileName = baseNameWithoutExt + ext + ".zip"
	} else {
		zipFileName = baseNameWithoutExt + ".zip"
	}
	outputZipPath := filepath.Join(originalFileDir, zipFileName)

	// 解压目录
	extractedDirName := strings.TrimSuffix(zipFileName, ".zip") + ".file"
	extractedDir = filepath.Join(originalFileDir, extractedDirName)

	// 如果是 xlsx 文件，使用 ConvertXlsxToMd 转换
	if ext == ".xlsx" {
		// 确保解压目录存在
		if err := os.MkdirAll(extractedDir, 0755); err != nil {
			l.Errorf("创建解压目录失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建解压目录失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 生成 md 文件路径
		mdFileName := baseNameWithoutExt + ".md"
		mdPath := filepath.Join(extractedDir, mdFileName)

		// 转换 xlsx 到 md
		if err := utils.ConvertXlsxToMd(originalFilePath, mdPath); err != nil {
			l.Errorf("转换 xlsx 到 md 失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "转换 xlsx 到 md 失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 创建 ZIP 文件，包含 md 文件
		zipFile, err := os.Create(outputZipPath)
		if err != nil {
			l.Errorf("创建 ZIP 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建 ZIP 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		// 添加 md 文件到 ZIP
		mdFile, err := os.Open(mdPath)
		if err != nil {
			l.Errorf("打开 MD 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "打开 MD 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}
		defer mdFile.Close()

		mdFileInfo, err := mdFile.Stat()
		if err != nil {
			l.Errorf("获取 MD 文件信息失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "获取 MD 文件信息失败",
					Info:    err.Error(),
				},
			}, nil
		}

		header, err := zip.FileInfoHeader(mdFileInfo)
		if err != nil {
			l.Errorf("创建 ZIP 文件头失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建 ZIP 文件头失败",
					Info:    err.Error(),
				},
			}, nil
		}
		header.Name = mdFileName
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			l.Errorf("创建 ZIP 写入器失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建 ZIP 写入器失败",
					Info:    err.Error(),
				},
			}, nil
		}

		if _, err := io.Copy(writer, mdFile); err != nil {
			l.Errorf("写入 ZIP 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "写入 ZIP 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}

		zipWriter.Close()
		zipFile.Close()

		zipPath = outputZipPath
		fileList = []string{mdFileName}

		// 对于 xlsx，解压成功后删除 ZIP 文件
		if err := os.Remove(zipPath); err != nil {
			l.Errorf("删除 ZIP 文件失败: %v, 文件路径: %s", err, zipPath)
		} else {
			l.Infof("成功删除 ZIP 文件: %s", zipPath)
		}

		// 写入 file_list，并读取 MD 内容更新 rawdoc（无论 md 是否读成功都至少写入 file_list）
		fileListStr, _ := json.Marshal(fileList)
		rawDoc.FileList = string(fileListStr)
		mdContent, readErr := os.ReadFile(mdPath)
		if readErr != nil {
			l.Errorf("读取 MD 文件失败: %v, 文件路径: %s", readErr, mdPath)
		} else {
			rawDoc.Content = string(mdContent)
			rawDoc.ContentOrg = string(mdContent)
			rawDoc.IsToMd = 1
			rawDoc.Status = constants.RawDocumentsStatusExtractedNotInDB
		}
		if updateErr := l.svcCtx.RawDocumentsModel.Update(l.ctx, rawDoc); updateErr != nil {
			l.Errorf("更新 rawdoc 内容/file_list 失败: %v, ID: %d", updateErr, rawDoc.Id)
		} else {
			l.Infof("成功更新 rawdoc 及 file_list, ID: %d", rawDoc.Id)
		}

	} else if ext == ".doc" || ext == ".docx" {
		// 对于 doc 和 docx，使用 pandoc 转换
		// 确保解压目录存在
		if err := os.MkdirAll(extractedDir, 0755); err != nil {
			l.Errorf("创建解压目录失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建解压目录失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 使用 pandoc 转换 doc/docx 到 md
		mdPath, err := utils.ConvertDocxToMd(originalFilePath, extractedDir)
		if err != nil {
			l.Errorf("转换 doc/docx 到 md 失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "转换 doc/docx 到 md 失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 创建 ZIP 文件，包含 md 文件和媒体文件
		zipFile, err := os.Create(outputZipPath)
		if err != nil {
			l.Errorf("创建 ZIP 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建 ZIP 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}
		defer zipFile.Close()

		zipWriter := zip.NewWriter(zipFile)
		defer zipWriter.Close()

		// 添加 md 文件到 ZIP
		mdFileName := filepath.Base(mdPath)
		mdFile, err := os.Open(mdPath)
		if err != nil {
			l.Errorf("打开 MD 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "打开 MD 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}
		defer mdFile.Close()

		mdFileInfo, err := mdFile.Stat()
		if err != nil {
			l.Errorf("获取 MD 文件信息失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "获取 MD 文件信息失败",
					Info:    err.Error(),
				},
			}, nil
		}

		header, err := zip.FileInfoHeader(mdFileInfo)
		if err != nil {
			l.Errorf("创建 ZIP 文件头失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建 ZIP 文件头失败",
					Info:    err.Error(),
				},
			}, nil
		}
		header.Name = mdFileName
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			l.Errorf("创建 ZIP 写入器失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建 ZIP 写入器失败",
					Info:    err.Error(),
				},
			}, nil
		}

		if _, err := io.Copy(writer, mdFile); err != nil {
			l.Errorf("写入 ZIP 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "写入 ZIP 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 添加媒体文件目录到 ZIP（如果存在）
		mediaDir := filepath.Join(extractedDir, "media")
		if _, err := os.Stat(mediaDir); err == nil {
			err = filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				relPath, err := filepath.Rel(extractedDir, path)
				if err != nil {
					return err
				}
				header, err := zip.FileInfoHeader(info)
				if err != nil {
					return err
				}
				header.Name = filepath.ToSlash(relPath)
				header.Method = zip.Deflate
				writer, err := zipWriter.CreateHeader(header)
				if err != nil {
					return err
				}
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(writer, file)
				return err
			})
			if err != nil {
				l.Errorf("添加媒体文件到 ZIP 失败: %v", err)
				// 不返回错误，继续执行
			}
		}

		zipWriter.Close()
		zipFile.Close()

		zipPath = outputZipPath

		// 获取文件列表
		fileList, err = utils.ListZipFiles(zipPath)
		if err != nil {
			l.Errorf("获取 ZIP 文件列表失败: %v", err)
			fileList = []string{}
		}

		// 解压成功后，删除 ZIP 文件
		if err := os.Remove(zipPath); err != nil {
			l.Errorf("删除 ZIP 文件失败: %v, 文件路径: %s", err, zipPath)
		} else {
			l.Infof("成功删除 ZIP 文件: %s", zipPath)
		}

		// 写入 file_list，并读取 MD 内容更新 rawdoc（无论 md 是否读成功都至少写入 file_list）
		fileListStr, _ := json.Marshal(fileList)
		rawDoc.FileList = string(fileListStr)
		mdContent, readErr := os.ReadFile(mdPath)
		if readErr != nil {
			l.Errorf("读取 MD 文件失败: %v, 文件路径: %s", readErr, mdPath)
		} else {
			rawDoc.Content = string(mdContent)
			rawDoc.ContentOrg = string(mdContent)
			rawDoc.IsToMd = 1
			rawDoc.Status = constants.RawDocumentsStatusExtractedNotInDB
		}
		if updateErr := l.svcCtx.RawDocumentsModel.Update(l.ctx, rawDoc); updateErr != nil {
			l.Errorf("更新 rawdoc 内容/file_list 失败: %v, ID: %d", updateErr, rawDoc.Id)
		} else {
			l.Infof("成功更新 rawdoc 及 file_list, ID: %d", rawDoc.Id)
		}

	} else {
		// 对于 pdf，使用 MinerU 转换
		if l.svcCtx.Config.MinerU.URL == "" {
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "MinerU 服务未配置",
				},
			}, nil
		}

		// 创建文档转换服务
		converter := utils.NewDocumentConverter(l.svcCtx.Config.MinerU.URL, filesRoot)

		// 转换文档为 ZIP
		zipPath, err = converter.ConvertDocumentToZIP(l.ctx, rawDoc, outputZipPath)
		if err != nil {
			l.Errorf("转换文档为 ZIP 失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "转换文档为 ZIP 失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 确保解压目录存在
		if err := os.MkdirAll(extractedDir, 0755); err != nil {
			l.Errorf("创建解压目录失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "创建解压目录失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 解压 ZIP 文件
		if err := utils.Unzip(zipPath, extractedDir); err != nil {
			l.Errorf("解压 ZIP 文件失败: %v", err)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "解压 ZIP 文件失败",
					Info:    err.Error(),
				},
			}, nil
		}

		// 获取解压后的文件列表（在删除 ZIP 文件之前）
		fileList, err = utils.ListZipFiles(zipPath)
		if err != nil {
			l.Errorf("获取 ZIP 文件列表失败: %v", err)
			fileList = []string{}
		}

		// 解压成功后，删除 ZIP 文件
		if err := os.Remove(zipPath); err != nil {
			l.Errorf("删除 ZIP 文件失败: %v, 文件路径: %s", err, zipPath)
		} else {
			l.Infof("成功删除 ZIP 文件: %s", zipPath)
		}

		// 对于 pdf，读取 MD 文件内容并更新到 rawdoc
		mdFileName := baseNameWithoutExt + ".md"
		mdFilePath := filepath.Join(extractedDir, mdFileName)
		var mdContent string
		mdFound := false

		// 先查 .file 根目录同名 md；找不到则递归查找同名 md；最后兜底全目录唯一 md。
		if _, err := os.Stat(mdFilePath); err == nil {
			mdFound = true
		} else {
			sameNameCandidates := make([]string, 0, 2)
			allMdCandidates := make([]string, 0, 4)
			walkErr := filepath.Walk(extractedDir, func(path string, info os.FileInfo, walkErr error) error {
				if walkErr != nil {
					return nil
				}
				if info == nil || info.IsDir() {
					return nil
				}
				if !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
					return nil
				}
				allMdCandidates = append(allMdCandidates, path)
				if strings.EqualFold(info.Name(), mdFileName) {
					sameNameCandidates = append(sameNameCandidates, path)
				}
				return nil
			})
			if walkErr != nil {
				l.Errorf("递归扫描 MD 文件失败: %v, dir=%s", walkErr, extractedDir)
			}

			if len(sameNameCandidates) > 0 {
				mdFilePath = sameNameCandidates[0]
				mdFound = true
				l.Infof("在子目录找到同名 MD 文件: %s", mdFilePath)
			} else if len(allMdCandidates) == 1 {
				mdFilePath = allMdCandidates[0]
				mdFound = true
				l.Infof("使用唯一 MD 文件作为识别结果: %s", mdFilePath)
			}
		}

		if mdFound {
			content, readErr := os.ReadFile(mdFilePath)
			if readErr != nil {
				l.Errorf("读取 MD 文件失败: %v, 文件路径: %s", readErr, mdFilePath)
			} else {
				mdContent = string(content)
				l.Infof("成功读取 MD 文件: %s, 内容长度: %d", mdFilePath, len(mdContent))
			}
		} else {
			l.Infof("未找到识别结果 MD 文件: expected=%s, dir=%s, ID=%d", mdFileName, extractedDir, rawDoc.Id)
		}
		// PDF 识别后若没有任何产出文件或找不到识别结果文件（md），按失败处理。
		if len(fileList) == 0 || !mdFound {
			info := fmt.Sprintf("识别结果缺失: fileList=%d, mdFound=%v, extractedDir=%s", len(fileList), mdFound, extractedDir)
			l.Errorf("PDF 识别失败: %s, ID: %d", info, rawDoc.Id)
			return &types.ConvertDocumentToZIPResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "PDF识别失败：未找到识别产出文件",
					Info:    info,
				},
			}, nil
		}

		// 写入 file_list，若有 MD 内容则同时更新 content（无论是否有 md 内容都至少写入 file_list）
		fileListStr, _ := json.Marshal(fileList)
		rawDoc.FileList = string(fileListStr)
		if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
			return nil, cancelErr
		}
		if mdContent != "" {
			rawDoc.Content = mdContent
			rawDoc.ContentOrg = mdContent
			rawDoc.IsToMd = 1
			rawDoc.Status = constants.RawDocumentsStatusExtractedNotInDB
		}
		if updateErr := l.svcCtx.RawDocumentsModel.Update(l.ctx, rawDoc); updateErr != nil {
			l.Errorf("更新 rawdoc 内容/file_list 失败: %v, ID: %d", updateErr, rawDoc.Id)
		} else {
			l.Infof("成功更新 rawdoc 及 file_list, ID: %d", rawDoc.Id)
		}
	}

	return &types.ConvertDocumentToZIPResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.ConvertDocumentToZIPData{
			Id:           rawDoc.Id,
			FileName:     rawDoc.FileName,
			ZipFilePath:  zipPath,
			ExtractedDir: extractedDir,
			FileList:     fileList,
		},
	}, nil
}
