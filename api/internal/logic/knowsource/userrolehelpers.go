package knowsource

import (
	"strings"

	"knowsource/model"
)

// frUserRolesToCodes 从 fr_user_roles 行生成 JWT 用的逗号串与返回前端的 roles 切片（去空、trim）。
// 若行存在但角色码全空（异常数据），兜底为 user，避免 strings.Split("", ",") 得到 [""] 导致前端无法识别角色。
func frUserRolesToCodes(rows []*model.FrUserRoles) (codes []string, rolesCSV string) {
	for _, r := range rows {
		if r == nil {
			continue
		}
		c := strings.TrimSpace(r.Role)
		if c == "" {
			continue
		}
		codes = append(codes, c)
	}
	if len(codes) == 0 {
		return []string{"user"}, "user"
	}
	return codes, strings.Join(codes, ",")
}
