package knowsource

import (
	"context"
	"fmt"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/constants"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkDocumentRagedLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标注文档已转化
func NewMarkDocumentRagedLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkDocumentRagedLogic {
	return &MarkDocumentRagedLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkDocumentRagedLogic) MarkDocumentRaged(req *types.MarkDocumentRagedRequest) (resp *types.Response, err error) {

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

	// 构建 SQL 条件：处理 .md 文件的情况
	// 如果文件名是 .md，去掉 .md 后尝试用 .doc, .docx, .pdf 来匹配
	var conditions []string
	var args []interface{}

	for _, docInfo := range req.RawDocumentFileNames {
		clientId := strings.TrimSpace(docInfo.ClientId)
		if clientId == "" {
			l.Errorf("clientId 为空，跳过更新: 文件名=%s, 文档类型=%s", docInfo.FileName, docInfo.DocumentCode)
			continue
		}
		fileName := docInfo.FileName
		documentCode := docInfo.DocumentCode

		// 检查是否是 .md 文件（不区分大小写）
		lowerFileName := strings.ToLower(fileName)
		if strings.HasSuffix(lowerFileName, ".md") {
			// 去掉 .md 扩展名（不区分大小写）
			baseName := fileName[:len(fileName)-3]

			// 生成可能的文件名：.doc, .docx, .pdf
			extensions := []string{".doc", ".docx", ".pdf"}
			var fileConditions []string

			for _, ext := range extensions {
				fileConditions = append(fileConditions, "(`client_id` = ? AND `file_name` = ? AND `document_code` = ?)")
				args = append(args, clientId, baseName+ext, documentCode)
			}

			// 使用 OR 连接多个可能的文件名
			conditions = append(conditions, "("+strings.Join(fileConditions, " OR ")+")")
		} else {
			// 非 .md 文件，直接使用原文件名
			conditions = append(conditions, "(`client_id` = ? AND `file_name` = ? AND `document_code` = ?)")
			args = append(args, clientId, fileName, documentCode)
		}
	}
	if len(conditions) == 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "文件名列表缺少有效 clientId",
		}, nil
	}

	// 使用 OR 连接所有条件
	whereClause := strings.Join(conditions, " OR ")

	// 批量更新 is_to_ai 为 1；已审核文档不再允许被回调改写状态
	query := fmt.Sprintf(
		"UPDATE `raw_documents` SET `is_to_ai` = 1, `status` = ? WHERE (%s) AND `is_audit` <> 1",
		whereClause,
	)

	args = append([]interface{}{constants.RawDocumentsStatusInserted}, args...)
	_, err = l.svcCtx.Mysql.ExecCtx(l.ctx, query, args...)
	if err != nil {
		l.Errorf("批量更新文档 IsToAI 失败: %v", err)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新失败",
			Info:    err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}
