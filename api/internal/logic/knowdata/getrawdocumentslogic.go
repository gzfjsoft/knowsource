package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type GetRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取原始文档
func NewGetRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRawDocumentsLogic {
	return &GetRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRawDocumentsLogic) GetRawDocuments(req *types.GetRawDocumentsRequest) (resp *types.GetRawDocumentsResp, err error) {
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

	// 检查 ID 是否有效
	if req.Id <= 0 {
		return &types.GetRawDocumentsResp{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "ID 不能为空或无效",
			},
		}, nil
	}

	// 根据 ID 查询文档
	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return &types.GetRawDocumentsResp{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "文档不存在",
				},
			}, nil
		}
		l.Logger.Errorf("查询原始文档失败: %v, ID: %d", err, req.Id)
		return &types.GetRawDocumentsResp{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 转换为响应类型
	var auditedAt int64 = 0
	if doc.AuditedAt.Valid {
		auditedAt = doc.AuditedAt.Time.Unix()
	}
	status := constants.ResolveRawDocumentListStatus(doc.Status, doc.IsAudit, doc.IsToMd)
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
		Status:        status,
		StatusMsg:     strings.TrimSpace(doc.StatusMsg),
	}

	return &types.GetRawDocumentsResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: rawDoc,
	}, nil
}
