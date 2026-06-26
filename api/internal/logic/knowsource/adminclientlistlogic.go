// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"encoding/json"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminClientListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// admin 列出 client
func NewAdminClientListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminClientListLogic {
	return &AdminClientListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminClientListLogic) AdminClientList(req *types.AdminClientListRequest) (resp *types.AdminClientListResponse, err error) {
	rolesStr, _ := l.ctx.Value("roles").(string)
	if !strings.Contains(rolesStr, consts.SUPER_ADMIN) {
		return &types.AdminClientListResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "没有权限",
			},
		}, nil
	}

	page := uint64(req.Page)
	pageSize := uint64(req.PageSize)
	list, total, qErr := l.svcCtx.ClientModel.List(l.ctx, req.ClientId, page, pageSize)
	if qErr != nil {
		return &types.AdminClientListResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    qErr.Error(),
			},
		}, nil
	}

	out := make([]types.AdminClientInfo, 0, len(list))
	for _, item := range list {
		name := ""
		desp := ""
		if strings.TrimSpace(item.ClientJsonInfo) != "" {
			var meta struct {
				Name string `json:"name"`
				Desp string `json:"desp"`
			}
			_ = json.Unmarshal([]byte(item.ClientJsonInfo), &meta)
			name = strings.TrimSpace(meta.Name)
			desp = strings.TrimSpace(meta.Desp)
		}
		out = append(out, types.AdminClientInfo{
			ClientId:       item.ClientId,
			Name:           name,
			Desp:           desp,
			ClientJsonInfo: item.ClientJsonInfo,
			CreatedAt:      item.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &types.AdminClientListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "ok",
		},
		Data: &types.AdminClientListData{
			List:  out,
			Total: total,
		},
	}, nil
}
