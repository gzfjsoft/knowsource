// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package knowsource

import (
	"context"
	"database/sql"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"knowsource/api/internal/superadmin"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminClientCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// admin 创建 client
func NewAdminClientCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminClientCreateLogic {
	return &AdminClientCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminClientCreateLogic) AdminClientCreate(req *types.AdminClientCreateRequest) (resp *types.Response, err error) {
	rolesStr, _ := l.ctx.Value("roles").(string)
	if !strings.Contains(rolesStr, consts.SUPER_ADMIN) {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "没有权限",
		}, nil
	}

	if strings.TrimSpace(req.ClientId) == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "clientId不能为空",
		}, nil
	}

	// 校验 clientId 格式：仅支持字母、数字、下划线和横线
	reg := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !reg.MatchString(req.ClientId) {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "clientId格式错误：仅支持字母、数字、下划线和横线",
		}, nil
	}

	if strings.TrimSpace(req.Name) == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "name不能为空",
		}, nil
	}

	// name 可以使用中文和其他字符，不需要格式验证
	metaBytes, _ := json.Marshal(map[string]string{
		"name": strings.TrimSpace(req.Name),
		"desp": strings.TrimSpace(req.Desp),
	})
	meta := string(metaBytes)

	// 若存在则更新，否则创建
	var opResult string
	existing, findErr := l.svcCtx.ClientModel.FindOneByClientId(l.ctx, req.ClientId)
	if findErr == nil && existing != nil {
		existing.ClientJsonInfo = meta
		if oe := strings.TrimSpace(req.OwnerEmail); oe != "" {
			existing.OwnerEmail = strings.ToLower(oe)
		}
		if upErr := l.svcCtx.ClientModel.Update(l.ctx, existing); upErr != nil {
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    upErr.Error(),
			}, nil
		}
		opResult = "updated"
	} else if findErr == nil || findErr == model.ErrNotFound {
		owner := strings.TrimSpace(req.OwnerEmail)
		_, insErr := l.svcCtx.ClientModel.Insert(l.ctx, &model.Client{
			ClientId:       strings.TrimSpace(req.ClientId),
			ClientJsonInfo: meta,
			RegistrationIp: "",
			Status:         1,
			VerifiedAt:     sql.NullTime{Time: time.Now(), Valid: true},
			OwnerEmail:     strings.ToLower(owner),
		})
		if insErr != nil {
			return &types.Response{
				Code:    response.ServerErrorCode,
				Message: "Database error",
				Info:    insErr.Error(),
			}, nil
		}
		opResult = "created"
	} else {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "Database error",
			Info:    findErr.Error(),
		}, nil
	}

	// 无论新建还是更新，都确保该 clientId 存在 superadmin 及其权限同步
	if _, se := superadmin.EnsureSuperadmin(l.ctx, l.svcCtx, strings.TrimSpace(req.ClientId)); se != nil {
		logx.Errorf("EnsureSuperadmin failed clientId=%s err=%v", req.ClientId, se)
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: opResult,
	}, nil
}
