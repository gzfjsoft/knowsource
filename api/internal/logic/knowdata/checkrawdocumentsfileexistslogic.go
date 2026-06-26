package knowdata

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckRawDocumentsFileExistsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 检查所有原始文档文件是否存在于硬盘
func NewCheckRawDocumentsFileExistsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckRawDocumentsFileExistsLogic {
	return &CheckRawDocumentsFileExistsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckRawDocumentsFileExistsLogic) CheckRawDocumentsFileExists() (resp *types.CheckRawDocumentsFileExistsResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.CheckRawDocumentsFileExistsResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// 获取文件根目录
	filesRoot := l.svcCtx.Config.Knowdata.DocumentPath
	if filesRoot == "" {
		filesRoot = l.svcCtx.Config.FilesRoot
	}

	// 获取所有rawdoc
	documents, err := l.svcCtx.RawDocumentsModel.FindAll(l.ctx, clientId)
	if err != nil {
		l.Logger.Errorf("获取原始文档列表失败: %v", err)
		return &types.CheckRawDocumentsFileExistsResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "获取原始文档列表失败",
				Info:    err.Error(),
			},
		}, nil
	}

	var list []types.RawDocumentFileStatus
	var existsCount int64 = 0
	var missingCount int64 = 0

	// 检查每个文件是否存在
	for _, doc := range documents {
		status := types.RawDocumentFileStatus{
			Id:           doc.Id,
			DocumentCode: doc.DocumentCode,
			FileName:     doc.FileName,
			FilePath:     doc.FilePath,
			FileSize:     doc.FileSize,
			Exists:       false,
		}

		// 先尝试直接使用 FilePath
		filePath := filepath.Clean(doc.FilePath)
		if _, err := os.Stat(filePath); err == nil {
			status.Exists = true
			status.ActualPath = filePath
			existsCount++
		} else {
			// 文件不存在，尝试拼接 FilesRoot
			filePath = filepath.Clean(filepath.Join(filesRoot, doc.FilePath))
			if _, err := os.Stat(filePath); err == nil {
				status.Exists = true
				status.ActualPath = filePath
				existsCount++
			} else {
				// 两种方式都失败，文件不存在
				status.Exists = false
				status.Error = err.Error()
				missingCount++
			}
		}

		list = append(list, status)
	}

	resp = &types.CheckRawDocumentsFileExistsResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "检查完成",
		},
		Data: types.CheckRawDocumentsFileExistsData{
			List:         list,
			Total:        int64(len(list)),
			ExistsCount:  existsCount,
			MissingCount: missingCount,
		},
	}

	return
}
