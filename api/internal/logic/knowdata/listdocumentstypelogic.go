package knowdata

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDocumentsTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文档类型列表
func NewListDocumentsTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDocumentsTypeLogic {
	return &ListDocumentsTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDocumentsTypeLogic) ListDocumentsType() (resp *types.ListDocumentsTypeResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.ListDocumentsTypeResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}

	documentTypes, err := l.svcCtx.DocumentTypeModel.FindAllByClientId(l.ctx, clientId)
	if err != nil {
		return nil, err
	}

	var list []types.DocumentsType
	for _, docType := range documentTypes {
		// 查询该知识库的所有不重复 tag
		tags, tagErr := l.svcCtx.RawDocumentsModel.FindDistinctTags(l.ctx, clientId, docType.Code)
		if tagErr != nil {
			// 如果查询 tag 失败，记录日志但不影响主流程，使用空数组
			l.Logger.Errorf("查询知识库 %s 的 tag 失败: %v", docType.Code, tagErr)
			tags = []string{}
		}

		list = append(list, types.DocumentsType{
			Code:        docType.Code,
			Name:        docType.Name,
			IsDisabled:  docType.IsDisabled,
			Description: docType.Description,
			Tags:        tags,
			CreatedAt:   docType.CreatedAt.Unix(),
			UpdatedAt:   docType.UpdatedAt.Unix(),
		})
	}

	resp = &types.ListDocumentsTypeResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "获取成功",
		},
		Data: &types.DocumentsTypeData{
			List:  list,
			Total: int64(len(list)),
		},
	}
	return
}
