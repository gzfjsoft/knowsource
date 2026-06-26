package knowdata

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateRawDocumentsContentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新原始文档内容
func NewUpdateRawDocumentsContentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateRawDocumentsContentLogic {
	return &UpdateRawDocumentsContentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateRawDocumentsContentLogic) UpdateRawDocumentsContent(req *types.UpdateRawDocumentsContentRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查 ID 是否有效
	if req.Id <= 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "ID 不能为空或无效",
		}, nil
	}

	// 查询文档是否存在
	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.RecordNotExistCode,
				Message: "文档不存在",
			}, nil
		}
		l.Logger.Errorf("查询原始文档失败: %v, ID: %d", err, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询失败",
			Info:    err.Error(),
		}, nil
	}

	// 检查文档是否是非审核状态（IsAudit != 1）且 IsToMd = 1
	if doc.IsAudit == 1 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "已审核的文档不能修改内容",
		}, nil
	}

	if doc.IsToMd != 1 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "只有已转换为 Markdown 的文档才能修改内容",
		}, nil
	}

	// 更新文档内容（保持 content_org 不变）
	// 注意：只更新 Content 字段，ContentOrg 字段保持不变（从数据库读取时已包含原值）
	doc.Content = req.Content
	doc.UpdatedAt = time.Now()

	err = l.svcCtx.RawDocumentsModel.Update(l.ctx, doc)
	if err != nil {
		l.Logger.Errorf("更新文档内容失败: %v, ID: %d", err, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新内容失败",
			Info:    err.Error(),
		}, nil
	}

	// 根据 file_list 找到 .md 文件并更新（不再写入 fix.md）；无 file_list 时回退为更新 FilePath 指向的文件
	if doc.FilePath == "" {
		// 无路径，不写文件
	} else if strings.TrimSpace(doc.FileList) != "" {
		var fileList []string
		if jsonErr := json.Unmarshal([]byte(doc.FileList), &fileList); jsonErr != nil {
			l.Logger.Errorf("解析 file_list 失败: %v, ID: %d", jsonErr, req.Id)
		} else {
			var targetRel string
			for _, name := range fileList {
				name = strings.TrimSpace(name)
				lower := strings.ToLower(name)
				if name != "" && (strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt")) {
					targetRel = name
					break
				}
			}
			if targetRel != "" {
				filesRoot := l.svcCtx.Config.Knowdata.DocumentPath
				if filesRoot == "" {
					filesRoot = l.svcCtx.Config.FilesRoot
				}
				baseDir := filepath.Dir(doc.FilePath)
				targetPath := filepath.Join(baseDir, filepath.FromSlash(targetRel))
				if !filepath.IsAbs(baseDir) && filesRoot != "" {
					targetPath = filepath.Join(filesRoot, targetPath)
				}
				if _, err := os.Stat(targetPath); err == nil {
					if writeErr := os.WriteFile(targetPath, []byte(req.Content), 0644); writeErr != nil {
						l.Logger.Errorf("更新 .md/.txt 文件失败: %v, 文件路径: %s, ID: %d", writeErr, targetPath, req.Id)
					} else {
						l.Logger.Infof("成功更新 .md/.txt 文件: %s, ID: %d", targetPath, req.Id)
					}
				} else {
					l.Logger.Infof("file_list 中的文件不存在，跳过更新: %s, ID: %d", targetPath, req.Id)
				}
			} else {
				l.Logger.Infof("file_list 中未找到 .md/.txt 文件, ID: %d", req.Id)
			}
		}
	} else {
		// 无 file_list 时：直接更新 FilePath 指向的文件（.md / .txt 上传的文档就是原文件路径）
		filesRoot := l.svcCtx.Config.Knowdata.DocumentPath
		if filesRoot == "" {
			filesRoot = l.svcCtx.Config.FilesRoot
		}
		filePath := doc.FilePath
		if !filepath.IsAbs(filePath) && filesRoot != "" {
			filePath = filepath.Join(filesRoot, filePath)
		}
		lower := strings.ToLower(filePath)
		isMdOrTxt := strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".txt")
		if _, err := os.Stat(filePath); err == nil && isMdOrTxt {
			if writeErr := os.WriteFile(filePath, []byte(req.Content), 0644); writeErr != nil {
				l.Logger.Errorf("更新 .md/.txt 文件失败: %v, 文件路径: %s, ID: %d", writeErr, filePath, req.Id)
			} else {
				l.Logger.Infof("成功更新 .md/.txt 文件: %s, ID: %d", filePath, req.Id)
			}
		}
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "更新成功",
	}, nil
}
