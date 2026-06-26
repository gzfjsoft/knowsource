package knowdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AuditRawDocumentsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 审核原始文档
func NewAuditRawDocumentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuditRawDocumentsLogic {
	return &AuditRawDocumentsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuditRawDocumentsLogic) AuditRawDocuments(req *types.AuditRawDocumentsRequest) (resp *types.Response, err error) {
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

	// 检查 IsAudit 只能为 1 或 0
	if req.IsAudit != 0 && req.IsAudit != 1 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "IsAudit 只能为 0 或 1",
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

	// 审核通过（IsAudit=1）时：必须已转 MD 成功才允许审核
	if req.IsAudit == 1 && doc.IsToMd != 1 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "转 MD 未成功，不允许审核；请先等待文档转换完成或重新上传",
		}, nil
	}

	// 获取用户名
	userName, ok := l.ctx.Value("userName").(string)
	if !ok {
		userName = ""
	}

	// 审核通过：入库流程（正在入库... -> 已经入库）
	if req.IsAudit == 1 {
		taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
		active, aErr := taskModel.FindActiveByTaskTypeAndSourceId(l.ctx, clientId, constants.AsyncTaskTypeRawDocumentsAuditIn, doc.Id)
		if aErr != nil {
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询审核任务失败",
				Info:    aErr.Error(),
			}, nil
		}
		if active != nil {
			return &types.Response{
				Code:    response.SuccessCode,
				Message: "正在入库...",
			}, nil
		}

		// 先标记「正在入库...」，并清空上次失败说明
		if _, upErr := knowsourceLogic.UpdateRawDocumentStatus(l.ctx, l.svcCtx, clientId, doc.Id, constants.RawDocumentsStatusInserting, ""); upErr != nil {
			l.Logger.Errorf("更新入库状态失败: %v, ID: %d", upErr, req.Id)
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "更新状态失败",
				Info:    upErr.Error(),
			}, nil
		}

		_, cErr := taskModel.CreateWithClientId(
			l.ctx,
			clientId,
			constants.AsyncTaskTypeRawDocumentsAuditIn,
			fmt.Sprintf("审核入库:%s", doc.FileName),
			doc.Id,
			userName,
		)
		if cErr != nil {
			// 创建异步任务失败时回滚状态，避免文档长期停留在「正在入库...」
			_, _ = knowsourceLogic.UpdateRawDocumentStatus(
				l.ctx,
				l.svcCtx,
				clientId,
				doc.Id,
				constants.RawDocumentsStatusExtractedNotInDB,
				"",
			)
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "创建审核任务失败",
				Info:    cErr.Error(),
			}, nil
		}
		_ = asynctasksignal.NotifyPending(l.ctx, l.svcCtx.RedisClient, clientId)

		return &types.Response{
			Code:    response.SuccessCode,
			Message: "已提交审核任务，后台入库中",
		}, nil
	}

	// 取消审核（IsAudit=0）：先按文件名称删除 Qdrant 中该文档的向量数据，再更新 DB
	if req.IsAudit == 0 {
		// 先标记“正在出库...”
		doc.Status = constants.RawDocumentsStatusRemoving
		doc.UpdatedAt = time.Now()
		if upErr := l.svcCtx.RawDocumentsModel.Update(l.ctx, doc); upErr != nil {
			l.Logger.Errorf("更新出库状态失败: %v, ID: %d", upErr, req.Id)
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "更新状态失败",
				Info:    upErr.Error(),
			}, nil
		}

		cfg := l.svcCtx.Config
		if cfg.Qdrant.Host != "" && cfg.Qdrant.Port > 0 && strings.TrimSpace(doc.FileName) != "" {
			prefix := cfg.Qdrant.CollectionPrefix

			// 删除主集合中的文档
			collectionName := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, false)
			qc, qErr := utils.NewQdrantToolsFromConfig(&cfg)
			if qErr != nil {
				l.Logger.Errorf("初始化 QdrantTools 失败: %v", qErr)
			} else {
				if delErr := qc.DeletePointsByFileName(l.ctx, collectionName, doc.FileName); delErr != nil {
					l.Logger.Errorf("取消审核时删除 Qdrant 向量失败: %v, collection=%s fileName=%s", delErr, collectionName, doc.FileName)
					// 继续执行 DB 更新，保证审核标志被置为 0
				} else {
					l.Infof("已从 Qdrant 删除 fileName=%s collection=%s", doc.FileName, collectionName)
				}

				// 删除概要集合中的文档
				summaryCollectionName := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, true)
				if delErr := qc.DeletePointsByFileName(l.ctx, summaryCollectionName, doc.FileName); delErr != nil {
					l.Logger.Errorf("取消审核时删除 Qdrant 概要向量失败: %v, collection=%s fileName=%s", delErr, summaryCollectionName, doc.FileName)
					// 继续执行 DB 更新，保证审核标志被置为 0
				} else {
					l.Infof("已从 Qdrant 概要集合删除 fileName=%s collection=%s", doc.FileName, summaryCollectionName)
				}

				// 删除问答集合中的文档
				qaCollectionName := utils.FormatCollectionName(prefix, clientId, doc.DocumentCode, false) + "_qa"
				if delErr := qc.DeletePointsByFileName(l.ctx, qaCollectionName, doc.FileName); delErr != nil {
					l.Logger.Errorf("取消审核时删除 Qdrant 问答向量失败: %v, collection=%s fileName=%s", delErr, qaCollectionName, doc.FileName)
				} else {
					l.Infof("已从 Qdrant 问答集合删除 fileName=%s collection=%s", doc.FileName, qaCollectionName)
				}
			}
		}

		// 删除问答队列表数据
		if l.svcCtx.RawDocumentQaPairsModel != nil {
			if delErr := l.svcCtx.RawDocumentQaPairsModel.DeleteByRawDocumentId(l.ctx, clientId, doc.Id); delErr != nil {
				l.Logger.Errorf("取消审核时删除问答队列表失败: %v, rawDocumentId=%d", delErr, doc.Id)
			}
		}
	}

	// 更新审核状态
	doc.IsAudit = req.IsAudit
	doc.UpdatedAt = time.Now()

	var message string
	if req.IsAudit == 1 {
		// IsAudit=1 时记录用户名和审核时间
		doc.AuditUser = userName
		doc.AuditedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
		message = "审核操作成功"
	} else {
		// IsAudit=0 时清除用户名和审核时间
		doc.AuditUser = ""
		doc.AuditedAt = sql.NullTime{
			Time:  time.Time{},
			Valid: false,
		}
		// 出库完成后回到“已提取文字未审核入库”
		doc.Status = constants.RawDocumentsStatusExtractedNotInDB
		message = "取消审核操作成功"
	}

	err = l.svcCtx.RawDocumentsModel.Update(l.ctx, doc)
	if err != nil {
		l.Logger.Errorf("更新文档审核状态失败: %v, ID: %d", err, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新审核状态失败",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: message,
	}, nil
}
