package superadmin

import (
	"context"
	"database/sql"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/consts"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	superadminEmpCode = consts.SUPER_ADMIN
	superadminRole    = consts.SUPER_ADMIN
)

// SyncPermissions 从 demo 租户同步权限列表到目标租户
func SyncPermissions(ctx context.Context, svcCtx *svc.ServiceContext, targetClientId string) error {
	_ = svcCtx
	_ = targetClientId
	return nil
}

// EnsureSuperadminRoleBindings 确保 superadmin 角色、全权限绑定、用户-角色关联（不创建员工、不改密码）
func EnsureSuperadminRoleBindings(ctx context.Context, svcCtx *svc.ServiceContext, clientId string) error {
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return sql.ErrNoRows
	}
	_ = SyncPermissions(ctx, svcCtx, clientId)

	// superadmin 角色
	_, err := svcCtx.FrRolesModel.FindOneByClientIdRole(ctx, clientId, superadminRole)
	if err != nil {
		if err == model.ErrNotFound {
			_, err = svcCtx.FrRolesModel.Insert(ctx, &model.FrRoles{
				ClientId: clientId,
				Role:     superadminRole,
				Name:     "超级管理员",
			})
			if err != nil {
				logx.Errorf("[%s] 创建 superadmin 角色失败: %v", clientId, err)
				return err
			}
			logx.Infof("[%s] 已创建 superadmin 角色", clientId)
		} else {
			logx.Errorf("[%s] 查询 superadmin 角色失败: %v", clientId, err)
			return err
		}
	}

	var permRows []struct {
		Permission string `db:"permission"`
	}
	if qErr := svcCtx.Mysql.QueryRowsCtx(ctx, &permRows, "SELECT permission FROM fr_permissions"); qErr != nil {
		logx.Errorf("[%s] 查询全部权限失败: %v", clientId, qErr)
		return qErr
	}
	bound := 0
	for _, row := range permRows {
		_, findErr := svcCtx.FrRolesPermissionsModel.FindOneByClientIdRolePermission(ctx, clientId, superadminRole, row.Permission)
		if findErr == model.ErrNotFound {
			_, insErr := svcCtx.FrRolesPermissionsModel.Insert(ctx, &model.FrRolesPermissions{
				ClientId:   clientId,
				Role:       superadminRole,
				Permission: row.Permission,
			})
			if insErr == nil {
				bound++
			}
		}
	}
	if bound > 0 {
		logx.Infof("[%s] 已为 superadmin 角色补绑 %d 个权限", clientId, bound)
	}

	_, err = svcCtx.FrUserRolesModel.FindOneByClientIdEmpCodeRole(ctx, clientId, superadminEmpCode, superadminRole)
	if err != nil {
		if err == model.ErrNotFound {
			_, err = svcCtx.FrUserRolesModel.Insert(ctx, &model.FrUserRoles{
				ClientId: clientId,
				EmpCode:  superadminEmpCode,
				Role:     superadminRole,
			})
			if err != nil {
				logx.Errorf("[%s] 绑定 superadmin 用户到角色失败: %v", clientId, err)
				return err
			}
			logx.Infof("[%s] 已绑定 superadmin 用户到角色", clientId)
		} else {
			logx.Errorf("[%s] 查询 superadmin 用户角色失败: %v", clientId, err)
			return err
		}
	}
	return nil
}

// ensureSuperadminEmp 若无 superadmin 员工则创建；不自动设置密码，需通过重置密码流程初始化
func ensureSuperadminEmp(ctx context.Context, svcCtx *svc.ServiceContext, clientId string) error {
	_, err := svcCtx.FrEmpModel.FindOneByClientIdFempCode(ctx, clientId, superadminEmpCode)
	if err != nil {
		if err == model.ErrNotFound {
			_, err = svcCtx.FrEmpModel.Insert(ctx, &model.FrEmp{
				ClientId:  clientId,
				Frylb:     "在职人员",
				Status:    0,
				FempName:  "超级管理员",
				FempCode:  superadminEmpCode,
				DeptCode:  "001",
				FdeptId:   sql.NullInt64{},
				Fposition: "超级管理员",
			})
			if err != nil {
				logx.Errorf("[%s] 创建 superadmin 用户失败: %v", clientId, err)
				return err
			}
			logx.Infof("[%s] 已创建 superadmin 用户（未设置密码，请通过重置密码流程初始化）", clientId)
		} else {
			logx.Errorf("[%s] 查询 superadmin 用户失败: %v", clientId, err)
			return err
		}
	}
	return nil
}

// EnsureSuperadmin 确保 superadmin 账号、角色与权限绑定（不设置默认密码）
func EnsureSuperadmin(ctx context.Context, svcCtx *svc.ServiceContext, clientId string) (string, error) {
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return "", sql.ErrNoRows
	}
	if err := ensureSuperadminEmp(ctx, svcCtx, clientId); err != nil {
		return "", err
	}
	if err := EnsureSuperadminRoleBindings(ctx, svcCtx, clientId); err != nil {
		return "", err
	}
	return superadminEmpCode, nil
}

// SuperadminEmpCode 供租户注册等场景使用
func SuperadminEmpCode() string { return superadminEmpCode }
