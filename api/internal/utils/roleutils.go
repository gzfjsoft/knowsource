package utils

import (
	"context"
	"strings"

	"knowsource/consts"
)

// IsAdminRole 检查角色字符串中是否包含 admin 或 superadmin
// 支持单个角色字符串或逗号分隔的角色字符串
func IsAdminRole(role string) bool {
	if role == "" {
		return false
	}
	roles := strings.Split(role, ",")
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == consts.ONLY_ADMIN || r == consts.SUPER_ADMIN {
			return true
		}
	}
	return false
}

// IsSuperAdminRole 检查角色字符串中是否包含 superadmin
// 支持单个角色字符串或逗号分隔的角色字符串
func IsSuperAdminRole(role string) bool {
	if role == "" {
		return false
	}
	roles := strings.Split(role, ",")
	for _, r := range roles {
		r = strings.TrimSpace(r)
		if r == consts.SUPER_ADMIN {
			return true
		}
	}
	return false
}

// GetRoleFromContext 从 context 中获取角色字符串
func GetRoleFromContext(ctx context.Context) (string, bool) {
	roleValue := ctx.Value("role")
	if roleValue == nil {
		return "", false
	}
	role, ok := roleValue.(string)
	return role, ok
}

// IsAdminRoleFromContext 从 context 中获取角色并判断是否是管理员
func IsAdminRoleFromContext(ctx context.Context) bool {
	role, ok := GetRoleFromContext(ctx)
	if !ok {
		return false
	}
	return IsAdminRole(role)
}

// IsSuperAdminRoleFromContext 从 context 中获取角色并判断是否是超级管理员
func IsSuperAdminRoleFromContext(ctx context.Context) bool {
	role, ok := GetRoleFromContext(ctx)
	if !ok {
		return false
	}
	return IsSuperAdminRole(role)
}
