package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/cryptx"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowsourceChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改密码
func NewKnowsourceChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowsourceChangePasswordLogic {
	return &KnowsourceChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowsourceChangePasswordLogic) KnowsourceChangePassword(req *types.KnowsourceChangePasswordRequest) (resp *types.Response, err error) {
	// 验证参数
	if req.EmpCode == "" || req.OldPassword == "" || req.NewPassword == "" {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "员工编码、旧密码和新密码不能为空",
		}, nil
	}

	clientId, _ := l.ctx.Value("clientId").(string)
	if strings.TrimSpace(clientId) == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 查找员工密码（按 tenant）
	empPassword, err := l.svcCtx.EmpPasswordModel.FindOneByClientIdEmpCode(l.ctx, clientId, req.EmpCode)
	if err != nil {
		if err == model.ErrNotFound {
			return &types.Response{
				Code:    response.UserNotExistCode,
				Message: "员工密码未设置",
			}, nil
		}
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	// 验证旧密码
	oldPasswordHash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.OldPassword)
	if empPassword.Password != oldPasswordHash {
		return &types.Response{
			Code:    response.ForbiddenCode,
			Message: "旧密码错误",
		}, nil
	}

	// 加密新密码
	newPasswordHash := cryptx.PasswordEncrypt(l.svcCtx.Config.Salt, req.NewPassword)

	// 更新密码
	empPassword.Password = newPasswordHash
	err = l.svcCtx.EmpPasswordModel.Update(l.ctx, empPassword)
	if err != nil {
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "密码修改成功",
	}, nil
}
