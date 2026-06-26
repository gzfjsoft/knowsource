package knowsource

import (
	"context"
	"strings"

	"knowsource/api/internal/superadmin"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImportHrUserDeptLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// import hr user dept
func NewImportHrUserDeptLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImportHrUserDeptLogic {
	return &ImportHrUserDeptLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ImportHrUserDeptLogic) ImportHrUserDept() (resp *types.Response, err error) {
	// todo: add your logic here and delete this line

	err = ImportHrUserDept()

	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		clientId = "demo"
	}
	if empRow, fErr := l.svcCtx.FrEmpModel.FindOneByClientIdFempCode(l.ctx, clientId, consts.SUPER_ADMIN); fErr == nil && empRow != nil {
		_ = l.svcCtx.FrEmpModel.Delete(l.ctx, empRow.Id)
	}

	_, err1 := l.svcCtx.FrEmpModel.Insert(l.ctx, &model.FrEmp{
		ClientId: clientId,
		Frylb:    "在职人员",
		Status:   0,

		FempCode:  consts.SUPER_ADMIN,
		FempName:  "超级管理员",
		DeptCode:  "001",
		Fposition: "超级管理员",
		Fbranch:   "总部",
	})
	if err1 != nil {
		logx.Errorw("导入HR系统superadmin用户失败", logx.Field("error", err1))
	}

	if _, se := superadmin.EnsureSuperadmin(l.ctx, l.svcCtx, clientId); se != nil {
		logx.Errorw("确保 superadmin 角色绑定失败", logx.Field("error", se))
	}

	if err != nil {
		return &types.Response{
			Code:    500,
			Info:    err.Error(),
			Message: "导入HR系统用户失败",
		}, nil
	}

	return &types.Response{
		Code:    200,
		Message: "导入HR系统用户成功",
		Info:    "导入HR系统用户成功",
	}, nil
}
