package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/constants"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkDocumentConvertedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标注文档已转化
func NewMarkDocumentConvertedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkDocumentConvertedLogic {
	return &MarkDocumentConvertedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkDocumentConvertedLogic) MarkDocumentConverted(req *types.MarkDocumentConvertedRequest) (resp *types.Response, err error) {

	if req.Authorization == "" {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Authorization 不能为空",
		}, nil
	}

	// 检查 Authorization 是否为 RAG_SERVER_TOKEN
	if req.Authorization != "RAG_SERVER_TOKEN" {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Authorization 不正确",
		}, nil
	}

	// 检查文件名列表是否为空
	if len(req.RawDocumentFileNames) == 0 {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "文件名列表不能为空",
		}, nil
	}

	// 逐个更新文档，因为每个文档的 content 可能不同
	successCount := 0
	for _, docInfo := range req.RawDocumentFileNames {
		clientId := strings.TrimSpace(docInfo.ClientId)
		if clientId == "" {
			l.Errorf("clientId 为空，跳过更新: 文件名=%s, 文档类型=%s", docInfo.FileName, docInfo.DocumentCode)
			continue
		}
		// 构建更新 SQL，同时更新 is_to_md 和 content
		var query string
		var args []interface{}

		if docInfo.Content != "" {
			// 如果提供了 content，同时更新 is_to_md 和 content
			query = "UPDATE `raw_documents` SET `is_to_md` = 1, `content` = ?, `content_org` = ?, `status` = ? WHERE `client_id` = ? AND `file_name` = ? AND `document_code` = ? AND `is_audit` <> 1"
			args = []interface{}{docInfo.Content, docInfo.Content, constants.RawDocumentsStatusExtractedNotInDB, clientId, docInfo.FileName, docInfo.DocumentCode}
		} else {
			// 如果没有提供 content，只更新 is_to_md
			query = "UPDATE `raw_documents` SET `is_to_md` = 1, `status` = ? WHERE `client_id` = ? AND `file_name` = ? AND `document_code` = ? AND `is_audit` <> 1"
			args = []interface{}{constants.RawDocumentsStatusExtractedNotInDB, clientId, docInfo.FileName, docInfo.DocumentCode}
		}

		result, err := l.svcCtx.Mysql.ExecCtx(l.ctx, query, args...)
		if err != nil {
			l.Errorf("更新文档失败: %v, 文件名: %s, 文档类型: %s", err, docInfo.FileName, docInfo.DocumentCode)
			continue
		}

		// 检查是否有行被更新
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			l.Errorf("获取更新行数失败: %v, 文件名: %s, 文档类型: %s", err, docInfo.FileName, docInfo.DocumentCode)
			continue
		}

		if rowsAffected > 0 {
			successCount++
			l.Infof("成功更新文档: 文件名=%s, 文档类型=%s, 是否更新content=%v",
				docInfo.FileName, docInfo.DocumentCode, docInfo.Content != "")
		} else {
			l.Infof("未找到匹配的文档: 文件名=%s, 文档类型=%s", docInfo.FileName, docInfo.DocumentCode)
		}
	}

	if successCount == 0 {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "没有文档被更新",
		}, nil
	}

	l.Infof("批量更新完成: 成功更新 %d/%d 个文档", successCount, len(req.RawDocumentFileNames))

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}
