package knowsource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/otiai10/copy"
	"github.com/zeromicro/go-zero/core/logx"
)

type AdminChangeFileDocumentTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// admin 移动文件到其他文档类型
func NewAdminChangeFileDocumentTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminChangeFileDocumentTypeLogic {
	return &AdminChangeFileDocumentTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminChangeFileDocumentTypeLogic) AdminChangeFileDocumentType(req *types.AdminChangeFileDocumentTypeRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// Validate request
	if req.FileName == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "文件名不能为空",
		}, nil
	}

	if req.OldDocumentCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "原文档类型编码不能为空",
		}, nil
	}

	if req.NewDocumentCode == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "新文档类型编码不能为空",
		}, nil
	}

	// Check if old and new document codes are the same
	if req.OldDocumentCode == req.NewDocumentCode {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "原文档类型和新文档类型不能相同",
		}, nil
	}

	// Validate old document type exists
	_, err = l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.OldDocumentCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: fmt.Sprintf("原文档类型 %s 不存在", req.OldDocumentCode),
			}, nil
		}
		l.Logger.Errorf("查询原文档类型失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询原文档类型失败",
		}, nil
	}

	// Validate new document type exists
	_, err = l.svcCtx.DocumentTypeModel.FindOneByClientIdCode(l.ctx, clientId, req.NewDocumentCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: fmt.Sprintf("新文档类型 %s 不存在", req.NewDocumentCode),
			}, nil
		}
		l.Logger.Errorf("查询新文档类型失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询新文档类型失败",
		}, nil
	}

	// Find the file by fileName and old document code
	file, err := l.svcCtx.RawDocumentsModel.FindByFileNameAndDocumentCode(l.ctx, clientId, req.FileName, req.OldDocumentCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.ParameterErrorCode,
				Message: fmt.Sprintf("文件 %s 在文档类型 %s 中不存在", req.FileName, req.OldDocumentCode),
			}, nil
		}
		l.Logger.Errorf("查询文件失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询文件失败",
		}, nil
	}

	// 检查 IsAudit=1 的不能更改类型
	if file.IsAudit == 1 {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "已审核的文档不能更改类型",
		}, nil
	}

	// Check if new document type already has a file with the same filename
	existingByFileName, err := l.svcCtx.RawDocumentsModel.FindByFileNameAndDocumentCode(l.ctx, clientId, req.FileName, req.NewDocumentCode)
	if err == nil && existingByFileName != nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: fmt.Sprintf("新文档类型 %s 中已存在同名文件 %s", req.NewDocumentCode, req.FileName),
		}, nil
	}
	if err != nil && err != model.ErrNotFound {
		l.Logger.Errorf("检查文件名失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "检查文件名失败",
		}, nil
	}

	// Check if new document type already has a file with the same MD5
	existingByMD5, err := l.svcCtx.RawDocumentsModel.FindByMD5AndDocumentCode(l.ctx, clientId, file.FileMd5, req.NewDocumentCode)
	if err == nil && existingByMD5 != nil {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: fmt.Sprintf("新文档类型 %s 中已存在相同MD5的文件 (MD5: %s)", req.NewDocumentCode, file.FileMd5),
		}, nil
	}
	if err != nil && err != model.ErrNotFound {
		l.Logger.Errorf("检查MD5失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "检查MD5失败",
		}, nil
	}

	// Move file to new document type directory
	oldFilePath := file.FilePath
	rawDocumentsDir := l.svcCtx.Config.Knowdata.DocumentPath
	newDocumentDir := filepath.Join(rawDocumentsDir, req.NewDocumentCode)

	// Create new document type directory if it doesn't exist
	if err := os.MkdirAll(newDocumentDir, 0755); err != nil {
		l.Logger.Errorf("创建新文档类型目录失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "创建新文档类型目录失败",
		}, nil
	}

	// Get the filename from the old path
	fileName := filepath.Base(oldFilePath)
	newFilePath := filepath.Join(newDocumentDir, fileName)

	// Check if source file exists
	if _, err := os.Stat(oldFilePath); os.IsNotExist(err) {
		l.Logger.Errorf("源文件不存在: %s", oldFilePath)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: fmt.Sprintf("源文件不存在: %s", oldFilePath),
		}, nil
	}

	// Move file to new location
	if err := moveFile(oldFilePath, newFilePath); err != nil {
		l.Logger.Errorf("移动文件失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: fmt.Sprintf("移动文件失败: %v", err),
		}, nil
	}

	// Move meta.yaml file if it exists
	oldTagFilePath := oldFilePath + ".meta.yaml"
	newTagFilePath := newFilePath + ".meta.yaml"
	if _, err := os.Stat(oldTagFilePath); err == nil {
		// meta.yaml 文件存在，移动它
		if err := moveFile(oldTagFilePath, newTagFilePath); err != nil {
			l.Logger.Errorf("移动meta.yaml文件失败: %v, 从 %s 到 %s", err, oldTagFilePath, newTagFilePath)
			// 不中断流程，只记录错误
		} else {
			l.Infof("成功移动meta.yaml文件: %s -> %s", oldTagFilePath, newTagFilePath)
		}
	} else if os.IsNotExist(err) {
		// meta.yaml 文件不存在，这是正常的（可能旧文件没有meta.yaml），只记录日志
		l.Infof("meta.yaml文件不存在，跳过移动: %s", oldTagFilePath)
	} else {
		// 其他错误
		l.Logger.Errorf("检查meta.yaml文件失败: %v, 文件路径: %s", err, oldTagFilePath)
	}

	// Update document code, file path, and set IsToAi to false
	file.DocumentCode = req.NewDocumentCode
	file.FilePath = newFilePath
	file.IsToAi = 0 // 改变知识库时，将 isAI 设置为 false
	err = l.svcCtx.RawDocumentsModel.Update(l.ctx, file)
	if err != nil {
		// If database update fails, try to move file back
		if moveErr := moveFile(newFilePath, oldFilePath); moveErr != nil {
			l.Logger.Errorf("数据库更新失败，且文件回滚失败: %v, %v", err, moveErr)
		}
		// 如果数据库更新失败，也尝试回滚 meta.yaml 文件
		if _, statErr := os.Stat(newTagFilePath); statErr == nil {
			if moveErr := moveFile(newTagFilePath, oldTagFilePath); moveErr != nil {
				l.Logger.Errorf("数据库更新失败，meta.yaml文件回滚失败: %v", moveErr)
			}
		}
		l.Logger.Errorf("更新文档类型失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新文档类型失败",
		}, nil
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

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "文档类型更新成功",
	}, nil
}

// moveFile moves a file from source to destination
func moveFile(src, dst string) error {
	// First, copy the source file to the destination
	if err := copy.Copy(src, dst); err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// After successful copy, remove the source file
	if err := os.Remove(src); err != nil {
		// If removal fails, try to remove the copied file to maintain consistency
		os.Remove(dst)
		return fmt.Errorf("删除源文件失败: %w", err)
	}

	return nil
}
