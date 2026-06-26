package knowdata

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除原始文档
func NewDeleteRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteRawDocumentsLogic {
	return &DeleteRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteRawDocumentsLogic) DeleteRawDocuments(req *types.DeleteRawDocumentsRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	if len(req.Ids) == 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "请选择要删除的文档",
		}, nil
	}

	// 先检查是否有已审核的文档
	for _, id := range req.Ids {
		doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, id)
		if err != nil {
			if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
				continue // 文档不存在，跳过
			}
			l.Logger.Errorf("获取文档信息失败，ID: %d, 错误: %v", id, err)
			continue
		}

		// 检查 IsAudit=1 的不能删除
		if doc.IsAudit == 1 {
			return &types.Response{
				Code:    response.ConflictCode,
				Message: "已审核的文档不能删除",
			}, nil
		}
	}

	err = l.svcCtx.Mysql.TransactCtx(l.ctx, func(ctx context.Context, s sqlx.Session) error {
		rawDocumentsModel := l.svcCtx.RawDocumentsModel.WithSession(s)

		for _, id := range req.Ids {
			// Get document info before deletion (to get file path)
			doc, err := rawDocumentsModel.FindOneByClientId(ctx, clientId, id)
			if err != nil {
				l.Logger.Errorf("获取文档信息失败，ID: %d, 错误: %v", id, err)
				continue
			}

			// Delete from database first
			err = rawDocumentsModel.DeleteByClientId(ctx, clientId, id)
			if err != nil {
				l.Logger.Errorf("删除文档记录失败，ID: %d, 错误: %v", id, err)
				continue
			}

			// Delete physical file and related files
			if doc.FilePath != "" {
				fileDir := filepath.Dir(doc.FilePath)
				fileName := filepath.Base(doc.FileName)
				fileExt := strings.ToLower(filepath.Ext(fileName))

				// 删除 FilePath 指向的文件（可能是 md 文件或其他文件）
				if _, err := os.Stat(doc.FilePath); err == nil {
					if err := os.Remove(doc.FilePath); err != nil {
						l.Logger.Errorf("删除本地文件失败: %v, 文件路径: %s, 文档ID: %d", err, doc.FilePath, id)
					} else {
						l.Logger.Infof("成功删除本地文件: %s, 文档ID: %d", doc.FilePath, id)
					}
				} else if !os.IsNotExist(err) {
					l.Logger.Errorf("检查文件状态失败: %v, 文件路径: %s, 文档ID: %d", err, doc.FilePath, id)
				}

				// 如果是 xlsx 转换的 md 文件（IsToMd=1 且文件扩展名是 .md）
				if doc.IsToMd == 1 && fileExt == ".md" {
					// FilePath 指向的是 .file 目录中的 md 文件
					// fileDir 是 .file 目录，需要获取父目录来定位原 xlsx 文件
					parentDir := filepath.Dir(fileDir)

					// 推断原 xlsx 文件名（去掉 .md，加上 .xlsx）
					baseNameWithoutExt := strings.TrimSuffix(fileName, ".md")
					originalXlsxName := baseNameWithoutExt + ".xlsx"
					originalXlsxPath := filepath.Join(parentDir, originalXlsxName)

					// 删除原 xlsx 文件（如果存在）
					if _, err := os.Stat(originalXlsxPath); err == nil {
						if err := os.Remove(originalXlsxPath); err != nil {
							l.Logger.Errorf("删除原 xlsx 文件失败: %v, 文件路径: %s, 文档ID: %d", err, originalXlsxPath, id)
						} else {
							l.Logger.Infof("成功删除原 xlsx 文件: %s, 文档ID: %d", originalXlsxPath, id)
						}
					}

					// 删除 xxx.xlsx.meta.yaml 文件（在原文件所在目录）
					metaFileName := originalXlsxName + ".meta.yaml"
					metaFilePath := filepath.Join(parentDir, metaFileName)
					if _, err := os.Stat(metaFilePath); err == nil {
						if err := os.Remove(metaFilePath); err != nil {
							l.Logger.Errorf("删除 meta 文件失败: %v, 文件路径: %s, 文档ID: %d", err, metaFilePath, id)
						} else {
							l.Logger.Infof("成功删除 meta 文件: %s, 文档ID: %d", metaFilePath, id)
						}
					}

					// 删除 .file 目录（xxx.xlsx.file），fileDir 就是 .file 目录
					if _, err := os.Stat(fileDir); err == nil {
						if err := os.RemoveAll(fileDir); err != nil {
							l.Logger.Errorf("删除 .file 目录失败: %v, 目录路径: %s, 文档ID: %d", err, fileDir, id)
						} else {
							l.Logger.Infof("成功删除 .file 目录: %s, 文档ID: %d", fileDir, id)
						}
					}
				} else {
					// 其他文件（如 PDF）：删除对应的 .meta.yaml 和 .file 目录
					metaFilePath := doc.FilePath + ".meta.yaml"
					if _, err := os.Stat(metaFilePath); err == nil {
						if err := os.Remove(metaFilePath); err != nil {
							l.Logger.Errorf("删除 meta 文件失败: %v, 文件路径: %s, 文档ID: %d", err, metaFilePath, id)
						} else {
							l.Logger.Infof("成功删除 meta 文件: %s, 文档ID: %d", metaFilePath, id)
						}
					}
					// 删除对应的 .file 目录（如 xxx.pdf.file）
					fileDirPath := doc.FilePath + ".file"
					if _, err := os.Stat(fileDirPath); err == nil {
						if err := os.RemoveAll(fileDirPath); err != nil {
							l.Logger.Errorf("删除 .file 目录失败: %v, 目录路径: %s, 文档ID: %d", err, fileDirPath, id)
						} else {
							l.Logger.Infof("成功删除 .file 目录: %s, 文档ID: %d", fileDirPath, id)
						}
					}
				}

				// 如果有 zip 文件，删除 zip 文件和对应的 .file 目录
				if doc.ZipFileName != "" {
					// 确定 zip 文件所在目录
					// 如果是 xlsx 转换的 md，zip 文件在 parentDir；否则在 fileDir
					var zipFileDir string
					if doc.IsToMd == 1 && fileExt == ".md" {
						zipFileDir = filepath.Dir(fileDir) // parentDir
					} else {
						zipFileDir = fileDir
					}

					// 删除 zip 文件
					zipFilePath := filepath.Join(zipFileDir, doc.ZipFileName)
					if _, err := os.Stat(zipFilePath); err == nil {
						if err := os.Remove(zipFilePath); err != nil {
							l.Logger.Errorf("删除 zip 文件失败: %v, 文件路径: %s, 文档ID: %d", err, zipFilePath, id)
						} else {
							l.Logger.Infof("成功删除 zip 文件: %s, 文档ID: %d", zipFilePath, id)
						}
					}

					// 删除对应的 .file 目录（zip 文件名去掉 .zip，加上 .file）
					zipBaseName := strings.TrimSuffix(doc.ZipFileName, ".zip")
					zipFileDirName := zipBaseName + ".file"
					zipFileDirPath := filepath.Join(zipFileDir, zipFileDirName)
					if _, err := os.Stat(zipFileDirPath); err == nil {
						if err := os.RemoveAll(zipFileDirPath); err != nil {
							l.Logger.Errorf("删除 zip .file 目录失败: %v, 目录路径: %s, 文档ID: %d", err, zipFileDirPath, id)
						} else {
							l.Logger.Infof("成功删除 zip .file 目录: %s, 文档ID: %d", zipFileDirPath, id)
						}
					}
				}
			} else {
				l.Logger.Infof("文档ID: %d 的文件路径为空，跳过文件删除", id)
			}
		}

		return nil
	})

	if err != nil {
		l.Logger.Errorf("删除原始文档失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "删除原始文档失败",
		}, nil
	}

	resp = &types.Response{
		Code:    response.SuccessCode,
		Message: "删除成功",
	}
	return
}
