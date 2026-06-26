package knowdata

import (
	"context"
	"strings"

	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取原始文档列表
func NewListRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListRawDocumentsLogic {
	return &ListRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListRawDocumentsLogic) ListRawDocuments(req *types.ListRawDocumentsRequest) (resp *types.ListRawDocumentsResp, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.ListRawDocumentsResp{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	// Calculate pagination
	offset := int64((req.Page - 1) * req.PageSize)
	limit := int64(req.PageSize)

	resp = &types.ListRawDocumentsResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "获取成功",
		},
	}

	// Get documents
	documents, err := l.svcCtx.RawDocumentsModel.FindByDocumentCode(l.ctx, clientId, req.DocumentCode, req.FileName, req.Tag, req.IsAudit, offset, limit)
	if err != nil {
		l.Logger.Errorf("获取原始文档列表失败: %v", err)
		resp.Response.Code = response.ServerErrorCode
		resp.Response.Info = err.Error()
		return resp, nil
	}

	// Get total count
	total, err := l.svcCtx.RawDocumentsModel.CountByDocumentCode(l.ctx, clientId, req.DocumentCode, req.FileName, req.Tag, req.IsAudit)
	if err != nil {
		l.Logger.Errorf("获取原始文档总数失败: %v", err)
		resp.Response.Code = response.ServerErrorCode
		resp.Response.Info = err.Error()
		return resp, nil
	}

	var list []types.RawDocuments
	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	for _, doc := range documents {
		var auditedAt int64 = 0
		if doc.AuditedAt.Valid {
			auditedAt = doc.AuditedAt.Time.Unix()
		}

		status := constants.ResolveRawDocumentListStatus(doc.Status, doc.IsAudit, doc.IsToMd)
		if constants.IsRawDocumentInsertingStatus(status) && doc.IsAudit != 1 {
			activeTask, aErr := taskModel.FindActiveByTaskTypeAndSourceId(
				l.ctx,
				clientId,
				constants.AsyncTaskTypeRawDocumentsAuditIn,
				doc.Id,
			)
			// 没有活动审核任务但仍显示「正在入库...」时，自动纠正为可继续审核的状态
			if aErr == nil && activeTask == nil {
				status = constants.RawDocumentsStatusExtractedNotInDB
				if _, upErr := knowsourceLogic.UpdateRawDocumentStatus(
					l.ctx,
					l.svcCtx,
					clientId,
					doc.Id,
					constants.RawDocumentsStatusExtractedNotInDB,
					"",
				); upErr != nil {
					l.Logger.Errorf("纠正文档遗留入库状态失败: id=%d err=%v", doc.Id, upErr)
				}
			}
		}
		list = append(list, types.RawDocuments{
			Id:            doc.Id,
			DocumentCode:  doc.DocumentCode,
			FileMd5:       doc.FileMd5,
			FileName:      doc.FileName,
			FilePath:      doc.FilePath,
			FileSize:      doc.FileSize,
			Content:       "", // 列表不返回内容，减少数据量
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
		})
	}

	resp = &types.ListRawDocumentsResp{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "获取成功",
		},
		Data: types.ListRawDocumentsData{
			List:  list,
			Total: total,
		},
	}
	return
}
