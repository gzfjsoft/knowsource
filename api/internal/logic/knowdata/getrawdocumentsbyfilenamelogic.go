package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetRawDocumentsByFilenameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 根据文件名获取原始文档（返回 Markdown 内容）
func NewGetRawDocumentsByFilenameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRawDocumentsByFilenameLogic {
	return &GetRawDocumentsByFilenameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRawDocumentsByFilenameLogic) GetRawDocumentsByFilename(req *types.GetRawDocumentsByFilenameRequest) (resp *types.GetRawDocumentsResp, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.GetRawDocumentsResp{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	if req.FileName == "" {
		return &types.GetRawDocumentsResp{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "文件名不能为空",
			},
		}, nil
	}

	// 先按文件名精确查找（同名文件假定唯一；若后续需要可扩展为按 documentCode+fileName）
	doc, err := l.svcCtx.RawDocumentsModel.FindByFileName(l.ctx, clientId, req.FileName)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return &types.GetRawDocumentsResp{
				Response: types.Response{
					Code:    response.RecordNotExistCode,
					Message: "文档不存在",
				},
			}, nil
		}
		l.Logger.Errorf("根据文件名查询原始文档失败: %v, fileName: %s", err, req.FileName)
		return &types.GetRawDocumentsResp{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询失败",
				Info:    err.Error(),
			},
		}, nil
	}

	var auditedAt int64
	if doc.AuditedAt.Valid {
		auditedAt = doc.AuditedAt.Time.Unix()
	}

	rawDoc := &types.RawDocuments{
		Id:            doc.Id,
		DocumentCode:  doc.DocumentCode,
		FileMd5:       doc.FileMd5,
		FileName:      doc.FileName,
		FilePath:      doc.FilePath,
		FileSize:      doc.FileSize,
		Content:       doc.Content,
		ContentOrg:    doc.ContentOrg,
		FileList:      doc.FileList,
		Tag:           doc.Tag,
		ZipFileName:   doc.ZipFileName,
		ZipFileSize:   doc.ZipFileSize,
		CreatedAt:     doc.CreatedAt.Unix(),
		UpdatedAt:     doc.UpdatedAt.Unix(),
		IsAudit:       doc.IsAudit,
		IsToMd:        doc.IsToMd,
		IsToAi:        doc.IsToAi,
		UploadUser:    doc.UploadUser,
		UploadEmpcode: doc.UploadEmpcode,
		AuditUser:     doc.AuditUser,
		AuditedAt:     auditedAt,
		Status:        doc.Status,
	}

	return &types.GetRawDocumentsResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: rawDoc,
	}, nil
}
