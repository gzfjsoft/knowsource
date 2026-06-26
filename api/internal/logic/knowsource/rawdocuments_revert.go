package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/utils"
	"knowsource/common/constants"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

// RevertRawDocumentAfterCancelExtract 将文档状态恢复为中断识别前的空闲状态
func RevertRawDocumentAfterCancelExtract(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64) (targetStatus string, rows int64, err error) {
	clientId = strings.TrimSpace(clientId)
	if svcCtx == nil || docId <= 0 {
		return "", 0, nil
	}
	var doc *model.RawDocuments
	if clientId != "" {
		doc, err = svcCtx.RawDocumentsModel.FindOneByClientId(ctx, clientId, docId)
	} else {
		doc, err = svcCtx.RawDocumentsModel.FindOne(ctx, docId)
	}
	if err != nil || doc == nil {
		return "", 0, err
	}
	targetStatus = constants.RawDocumentStatusAfterCancelExtract(doc.IsToMd)
	rows, err = UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, targetStatus, "")
	if err != nil {
		return targetStatus, rows, err
	}
	if rows == 0 && clientId != "" && strings.TrimSpace(doc.ClientId) != "" && doc.ClientId != clientId {
		rows, err = UpdateRawDocumentStatus(ctx, svcCtx, doc.ClientId, docId, targetStatus, "")
	}
	return targetStatus, rows, err
}

// RevertRawDocumentAfterCancelAudit 中断审核入库后恢复为「已提取文字未审核入库」
func RevertRawDocumentAfterCancelAudit(ctx context.Context, svcCtx *svc.ServiceContext, clientId string, docId int64) (rows int64, err error) {
	clientId = strings.TrimSpace(clientId)
	if svcCtx == nil || docId <= 0 {
		return 0, nil
	}
	var doc *model.RawDocuments
	if clientId != "" {
		doc, err = svcCtx.RawDocumentsModel.FindOneByClientId(ctx, clientId, docId)
	} else {
		doc, err = svcCtx.RawDocumentsModel.FindOne(ctx, docId)
	}
	if err != nil || doc == nil {
		return 0, err
	}
	rows, err = UpdateRawDocumentStatus(ctx, svcCtx, clientId, docId, constants.RawDocumentsStatusExtractedNotInDB, "")
	if err != nil {
		return rows, err
	}
	_ = doc
	if clrErr := svcCtx.RawDocumentsModel.ClearAuditFields(ctx, clientId, docId); clrErr != nil {
		return rows, clrErr
	}
	if rows == 0 && clientId != "" && strings.TrimSpace(doc.ClientId) != clientId {
		rows, _ = UpdateRawDocumentStatus(ctx, svcCtx, doc.ClientId, docId, constants.RawDocumentsStatusExtractedNotInDB, "")
		_ = svcCtx.RawDocumentsModel.ClearAuditFields(ctx, doc.ClientId, docId)
	}

	// 任务取消时同步清理已生成的问答数据（MySQL + Qdrant）
	effectiveClientId := strings.TrimSpace(clientId)
	if effectiveClientId == "" {
		effectiveClientId = strings.TrimSpace(doc.ClientId)
	}
	if svcCtx.RawDocumentQaPairsModel != nil && effectiveClientId != "" {
		if delErr := svcCtx.RawDocumentQaPairsModel.DeleteByRawDocumentId(ctx, effectiveClientId, docId); delErr != nil {
			logx.WithContext(ctx).Errorf("cancel audit cleanup qa mysql failed: %v", delErr)
		}
	}
	if effectiveClientId != "" && strings.TrimSpace(doc.FileName) != "" && strings.TrimSpace(doc.DocumentCode) != "" {
		qaCollectionName := utils.FormatCollectionName(svcCtx.Config.Qdrant.CollectionPrefix, effectiveClientId, doc.DocumentCode, false) + "_qa"
		qc, qErr := utils.NewQdrantToolsFromConfig(&svcCtx.Config)
		if qErr == nil {
			if delErr := qc.DeletePointsByFileName(ctx, qaCollectionName, doc.FileName); delErr != nil {
				logx.WithContext(ctx).Errorf("cancel audit cleanup qa qdrant failed: %v", delErr)
			}
		}
	}
	return rows, nil
}
