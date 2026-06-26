package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"knowsource/api/internal/config"
	"knowsource/api/internal/utils"
	"knowsource/common/jwtx"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	// hdModel "knowsource/model/knowdata"
	"net/http"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func init() {
	assertAPIPermissionMapUsesOnlyFrCatalog()
}

// frPermissionsCatalog 与 api/fr_permissions_insert.sql 中 permission 列完全一致（修改 SQL 时请同步此处）。
var frPermissionsCatalog = map[string]struct{}{
	"功能-上传文档":    {},
	"功能-修改文档标签":  {},
	"功能-删除文档":    {},
	"功能-变更文档类型":  {},
	"功能-审核文档":    {},
	"功能-更新文档内容":  {},
	"菜单-首页":      {},
	"菜单-知识库问答":   {},
	"菜单-文档管理":    {},
	"菜单-文档标签":    {},
	"菜单-知识库类型":   {},
	"菜单-库内全文检索":  {},
	"菜单-员工管理":    {},
	"菜单-部门管理":    {},
	"菜单-角色管理":    {},
	"菜单-按员工授权":   {},
	"菜单-按部门授权":   {},
	"菜单-AI提示词配置": {},
	"菜单-LLM配置":   {},
	"菜单-AI调用统计":  {},
	"菜单-对话日志":    {},
	"菜单-异步任务队列":  {},
}

// frPermissionRequirementValid 校验 apiUrlPermissionMap 的值（可含 | 多选一）是否均为 fr_permissions 已定义编码。
func frPermissionRequirementValid(required string) bool {
	for _, p := range strings.Split(required, "|") {
		p = strings.TrimSpace(p)
		if p == "" {
			return false
		}
		if _, ok := frPermissionsCatalog[p]; !ok {
			return false
		}
	}
	return true
}

func assertAPIPermissionMapUsesOnlyFrCatalog() {
	for pathKey, req := range apiUrlPermissionMap {
		if !frPermissionRequirementValid(req) {
			panic(fmt.Sprintf("authmiddleware apiUrlPermissionMap: %q -> %q 含未收录于 fr_permissions_insert 的权限编码", pathKey, req))
		}
	}
}

type AuthMiddleware struct {
	AppConfig       config.Config
	RedisClient     *redis.Redis
	PermissionModel model.PermissionModel
	UserRoleModel   model.UserRoleModel
	UserModel       model.UserModel
}

func NewAuthMiddleware(c config.Config, conn sqlx.SqlConn, r *redis.Redis) *AuthMiddleware {
	return &AuthMiddleware{
		AppConfig:       c,
		RedisClient:     r,
		PermissionModel: model.NewPermissionModel(conn),
		UserRoleModel:   model.NewUserRoleModel(conn),
		UserModel:       model.NewUserModel(conn),
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeResponseUnauthorized(w, "未提供认证信息")
			return
		}

		// 检查Bearer前缀
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeResponseUnauthorized(w, "认证格式错误")
			return
		}

		// 提取token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			writeResponseUnauthorized(w, "token不能为空")
			return
		}

		// 获取appPlatform头
		appPlatformHeader := r.Header.Get("App-Platform")

		var userId int64
		var empCode string
		var clientId string
		var role string
		var isAdmin int64
		// var companyId int64
		// var orgId int64
		var userName string
		// var roleIds string

		//&& m.AppConfig.IsDebug == 1

		if m.AppConfig.AdminJWT != "" && token == m.AppConfig.AdminJWT {
			userId = 1
			empCode = consts.SUPER_ADMIN
			userName = "超级管理员"
			role = consts.SUPER_ADMIN
			isAdmin = 1
			clientId = strings.TrimSpace(r.Header.Get("Client-Id"))
			if clientId == "" {
				clientId = strings.TrimSpace(r.Header.Get("ClientId"))
			}

		} else {
			// 验证token
			claims, err := jwtx.ParseTokenKnowdataWithContext(r.Context(), token)
			if err != nil {
				writeResponseUnauthorized(w, "token无效: "+err.Error())
				return
			}
			userId = claims.UserId
			empCode = claims.EmpCode
			role = claims.Roles
			clientId = claims.ClientId
			// companyId = claims.CompanyId
			// orgId = claims.OrgId
			isAdmin = claims.IsAdmin
			userName = claims.UserName
			if strings.TrimSpace(clientId) == "" {
				writeResponseUnauthorized(w, "clientId不能为空，请重新登录")
				return
			}
			// 判断用户是否下线(因为被删除)
			key := fmt.Sprintf("Offline:User_%s_%d", clientId, userId)
			backList, err := m.RedisClient.Get(key)
			if err == nil && backList != "" {
				if backList == "2" {
					m.RedisClient.Setex(key, "1", 240)
				}
				writeResponseUnauthorized(w, "用户会话已中断，请重新登录")
				return
			}

			// 组织是否禁用
			// key = fmt.Sprintf("Offline:Organization_%d", orgId)
			// backList, err = m.RedisClient.Get(key)
			// if err == nil && backList != "" {
			// 	writeResponseUnauthorized(w, "组织是否禁用,用户会话已中断，请重新登录")
			// 	return
			// }
		}
		ctx := r.Context()
		// if roleLevel == 0 { // check Permissin
		// res := m.checkUserPermission(ctx, userId, r.Method+"|"+r.URL.Path, appPlatformHeader)
		// 	if !res {
		// 		httpx.WriteJson(w, http.StatusForbidden, response.Fail(http.StatusUnauthorized, "权限不足，无法访问"))
		// 		return
		// 	}
		// }
		if strings.TrimSpace(clientId) == "" {
			writeResponseUnauthorized(w, "clientId不能为空，请重新登录")
			return
		}

		err := m.checkUserPermissionByEmpCode(ctx, clientId, empCode, role, r.Method+"|"+r.URL.Path, appPlatformHeader)
		if err != nil {
			httpx.WriteJson(w, http.StatusOK, response.FailWithInfo(response.UnauthorizedCode, "权限不足，无法访问", err.Error()))
			return
		}
		// 将用户信息存入context
		isApp := 0
		if appPlatformHeader != "web" {
			isApp = 1
		}
		ctx = context.WithValue(ctx, "userId", userId)
		ctx = context.WithValue(ctx, "clientId", clientId)
		ctx = context.WithValue(ctx, "empCode", empCode)
		// ctx = context.WithValue(ctx, "companyId", companyId)
		ctx = context.WithValue(ctx, "role", role)
		ctx = context.WithValue(ctx, "roles", role)
		// ctx = context.WithValue(ctx, "orgId", orgId)
		ctx = context.WithValue(ctx, "userName", userName)
		ctx = context.WithValue(ctx, "isApp", isApp)
		ctx = context.WithValue(ctx, "isAdmin", isAdmin)
		// ctx = context.WithValue(ctx, "roleIds", roleIds)

		next(w, r.WithContext(ctx))
	}
}

func writeResponseUnauthorized(w http.ResponseWriter, message string) {
	httpx.WriteJson(w, http.StatusUnauthorized, response.Fail(http.StatusUnauthorized, message))
}

// isFrPlatformRootOperator 平台根账号：租户 clientId=admin 且登录 empCode=superadmin。
// 此类请求不校验 fr_permissions（与前端 isPlatformSuperUser 一致）。
func isFrPlatformRootOperator(clientId, empCode string) bool {
	return strings.EqualFold(strings.TrimSpace(clientId), consts.ONLY_ADMIN) &&
		strings.EqualFold(strings.TrimSpace(empCode), consts.SUPER_ADMIN)
}

// isAdminTenantSuperadminRestrictedPath MySQL 备份/下载：仅 admin 租户且 JWT 角色含 superadmin（与 logic 层一致）
func isAdminTenantSuperadminRestrictedPath(reqPath string) bool {
	n := normalizeURL(reqPath)
	if n == "POST|/api/v1/mysql/backup" {
		return true
	}
	return strings.HasPrefix(n, "GET|/api/v1/mysql/backup/download/")
}

func allowAdminTenantSuperadmin(clientId, role string) bool {
	return strings.EqualFold(strings.TrimSpace(clientId), consts.ONLY_ADMIN) && utils.IsSuperAdminRole(role)
}

// platformRootExclusiveAPI 仅平台根账号可调用的接口；其他用户一律拒绝且不查 fr_permissions。
var platformRootExclusiveAPI = map[string]struct{}{
	"POST|/api/raw-documents/regenerate-summaries":     {},
	"POST|/api/raw-documents/check-file-exists":        {},
	"POST|/api/knowsource/doc-path/tree":               {},
	"POST|/api/knowsource/qdrant/collection/list":      {},
	"POST|/api/knowsource/qdrant/collection/file/list": {},
	"POST|/api/knowsource/sys/check":                   {},
	"POST|/api/knowsource/sync/mysql":                  {},
	"POST|/api/knowsource/admin/client/create":         {},
	"POST|/api/knowsource/admin/client/delete":         {},
	"POST|/api/knowsource/admin/client/list":           {},
}

// API URL 与权限 Key 映射（仅使用 api/fr_permissions_insert.sql 中的 permission 编码）。
// 值中 | 表示多个子菜单共用同一接口时，拥有其一即可。
var apiUrlPermissionMap = map[string]string{
	"POST|/api/raw-documents/audit":                              "功能-审核文档",
	"POST|/api/raw-documents/audit/cancel":                       "功能-审核文档",
	"POST|/api/raw-documents/index-to-qdrant":                    "功能-审核文档",
	"POST|/api/raw-documents/qdrant/chunks":                      "功能-审核文档",
	"POST|/api/raw-documents/qa/list":                            "功能-审核文档",
	"POST|/api/raw-documents/upload":                             "功能-上传文档",
	"POST|/api/raw-documents/delete":                             "功能-删除文档",
	"POST|/api/raw-documents/tag/change":                         "功能-修改文档标签",
	"POST|/api/raw-documents/content/update":                     "功能-更新文档内容",
	"POST|/api/raw-documents/content/markdown-normalize/preview": "功能-更新文档内容",
	"POST|/api/knowsource/admin/change/file/document-type":       "功能-变更文档类型",

	// ----- 知识文档（子菜单） -----
	"POST|/api/raw-documents/list":            "菜单-文档管理",
	"POST|/api/raw-documents/get":             "菜单-文档管理",
	"POST|/api/raw-documents/get-by-filename": "菜单-文档管理",
	"GET|/api/raw-documents/download/:id":     "菜单-文档管理",
	"POST|/api/raw-documents/content/diff":    "菜单-文档管理",
	"POST|/api/raw-documents/search":          "菜单-库内全文检索",
	"POST|/api/raw-documents/vectors/search":  "菜单-库内全文检索",
	"POST|/api/raw-documents/tags/distinct":   "菜单-文档管理|菜单-文档标签",
	"POST|/api/documents/type/create":         "菜单-知识库类型",
	"POST|/api/documents/type/update":         "菜单-知识库类型",
	"POST|/api/documents/type/delete":         "菜单-知识库类型",
	"POST|/api/documents/type/list":           "菜单-知识库类型",
	"POST|/api/documents/type/get":            "菜单-知识库类型",

	"POST|/api/knowsource/document/convert/cancel": "菜单-文档管理",
	"POST|/api/knowsource/document/convert/to/md":  "菜单-文档管理",
	"POST|/api/knowsource/document/convert/to/zip": "菜单-文档管理",

	// ----- 组织与角色 -----
	"POST|/api/knowsource/emp/list":             "菜单-员工管理",
	"POST|/api/knowsource/emp/create":           "菜单-员工管理",
	"POST|/api/knowsource/emp/update":           "菜单-员工管理",
	"POST|/api/knowsource/emp/delete":           "菜单-员工管理",
	"POST|/api/knowsource/admin/reset/password": "菜单-员工管理",
	"POST|/api/knowsource/admin/set/role":       "菜单-员工管理",
	"POST|/api/knowsource/hr/import/user/dept":  "菜单-员工管理",

	// "POST|/api/knowsource/dept/tree": "菜单-部门管理|菜单-按部门授权|菜单-按员工授权",
	// "POST|/api/knowsource/dept/list": "菜单-部门管理|菜单-按部门授权|菜单-按员工授权",
	"POST|/api/knowsource/dept/create": "菜单-部门管理",
	"POST|/api/knowsource/dept/update": "菜单-部门管理",
	"POST|/api/knowsource/dept/delete": "菜单-部门管理",
	"POST|/api/knowsource/dept/move":   "菜单-部门管理",

	// ----- 知识库授权 -----
	"POST|/api/knowsource/emp-document-type/list":       "菜单-按员工授权",
	"POST|/api/knowsource/emp-document-type/list/group": "菜单-按员工授权",
	"POST|/api/knowsource/emp-document-type/create":     "菜单-按员工授权",
	"POST|/api/knowsource/emp-document-type/delete":     "菜单-按员工授权",
	// "POST|/api/knowsource/dept-document-type/list":      "菜单-按部门授权",
	"POST|/api/knowsource/dept-document-type/create": "菜单-按部门授权",
	"POST|/api/knowsource/dept-document-type/delete": "菜单-按部门授权",

	// ----- 角色 / 权限绑定（角色管理页） -----
	"POST|/api/knowsource/role/create":            "菜单-角色管理",
	"POST|/api/knowsource/role/delete":            "菜单-角色管理",
	"POST|/api/knowsource/role/update":            "菜单-角色管理",
	"POST|/api/knowsource/role/list":              "菜单-角色管理",
	"POST|/api/knowsource/role/get":               "菜单-角色管理",
	"POST|/api/knowsource/permission/list":        "菜单-角色管理",
	"POST|/api/knowsource/permission/create":      "菜单-角色管理",
	"POST|/api/knowsource/permission/update":      "菜单-角色管理",
	"POST|/api/knowsource/permission/delete":      "菜单-角色管理",
	"POST|/api/knowsource/permission/get":         "菜单-角色管理",
	"POST|/api/knowsource/role-permission/list":   "菜单-角色管理",
	"POST|/api/knowsource/role-permission/create": "菜单-角色管理",
	"POST|/api/knowsource/role-permission/update": "菜单-角色管理",
	"POST|/api/knowsource/role-permission/delete": "菜单-角色管理",
	"POST|/api/knowsource/role-permission/get":    "菜单-角色管理",
	"POST|/api/knowsource/user-role/list":         "菜单-角色管理",
	"POST|/api/knowsource/user-role/create":       "菜单-角色管理",
	"POST|/api/knowsource/user-role/update":       "菜单-角色管理",
	"POST|/api/knowsource/user-role/delete":       "菜单-角色管理",
	"POST|/api/knowsource/user-role/get":          "菜单-角色管理",

	"POST|/api/knowsource/tag/create": "菜单-文档标签",
	"POST|/api/knowsource/tag/delete": "菜单-文档标签",
	"POST|/api/knowsource/tag/update": "菜单-文档标签",
	"POST|/api/knowsource/tag/list":   "菜单-文档标签",
	"POST|/api/knowsource/tag/get":    "菜单-文档标签",

	"POST|/api/knowsource/dify/option/create":   "菜单-AI提示词配置",
	"POST|/api/knowsource/dify/option/delete":   "菜单-AI提示词配置",
	"POST|/api/knowsource/dify/option/get":      "菜单-AI提示词配置",
	"POST|/api/knowsource/dify/option/list":     "菜单-AI提示词配置",
	"POST|/api/knowsource/dify/option/update":   "菜单-AI提示词配置",
	"GET|/api/conf/ai/list":                     "菜单-AI提示词配置",
	"GET|/api/conf/ai/name/:name":               "菜单-知识库问答|菜单-AI提示词配置",
	"POST|/api/conf/ai/code/name/get":           "菜单-知识库问答|菜单-AI提示词配置",
	"POST|/api/conf/ai":                         "菜单-AI提示词配置",
	"POST|/api/conf/ai/:id":                     "菜单-AI提示词配置",
	"DELETE|/api/conf/ai/:id":                   "菜单-AI提示词配置",
	"POST|/api/knowsource/llm/setting/load":     "菜单-LLM配置",
	"POST|/api/knowsource/llm/setting/save":     "菜单-LLM配置",
	"POST|/api/knowsource/llm/setting/defaults": "菜单-LLM配置",
	"POST|/api/knowsource/llm/chat/models":      "菜单-LLM配置|菜单-知识库问答",
	"POST|/api/knowsource/llm/embedding/models": "菜单-LLM配置|菜单-知识库问答",
	"POST|/api/knowsource/llm/service/test":     "菜单-LLM配置|菜单-知识库问答",
	"POST|/api/knowsource/ai/call/stats/query":  "菜单-AI调用统计",
	"POST|/api/knowsource/async-task/list":      "菜单-异步任务队列",
	"POST|/api/knowsource/async-task/cancel":    "菜单-异步任务队列",
	"POST|/api/knowsource/async-task/delete":    "菜单-异步任务队列",
	"POST|/api/knowsource/async-task/watermark": "菜单-异步任务队列",
	"POST|/api/knowsource/system/stats":         "菜单-首页",
}

func (m *AuthMiddleware) checkUserPermission(ctx context.Context, userId int64, reqPath, platform string) bool {
	permission := m.getPermissionKeyByUrl(reqPath, platform)
	if permission == "" {
		return true
	}
	isApp := platform != "web"
	if isApp {
		return true
	}
	if permission == "super-admin" {
		logx.Error("验证API权限失败: 管理员才拥有的权限")
		return false
	}
	if permission == "" {
		return true
	}
	rolePermissionMap, err := m.PermissionModel.GetRolePermissionMapsCache(ctx, m.RedisClient)
	if err != nil {
		logx.Errorf("验证API权限失败【GetRolePermissionMapsCache】: %v", err)
		return false
	}

	userRoles, err := m.UserRoleModel.FindByUserIdWithCache(ctx, userId, m.RedisClient)
	if err != nil {
		logx.Errorf("验证API权限失败【FindByUserIdWithCache】: %v", err)
		return false
	}

	// arr := strings.Split(roleIds, ",")
	// roleIdArray := make([]int64, 0)
	// for _, v := range arr {
	// 	roleId := utils.ConvertInt64(v, 0)
	// 	if roleId == 0 {
	// 		continue
	// 	}
	// 	roleIdArray = append(roleIdArray, roleId)
	// }
	perms := strings.Split(permission, "|")
	for _, role := range userRoles {
		roleId := role.RoleId
		if permissions, ok := rolePermissionMap[roleId]; ok {
			for _, p := range permissions {
				for _, perm := range perms {
					if perm == p {
						return true
					}
				}
			}
		}
	}
	return false
}

// normalizeURL converts a URL with actual parameters to its pattern equivalent
// e.g., /api/user/123 -> /api/user/:id, /api/video/456/chapter -> /api/video/:id/chapter
func normalizeURL(url string) string {
	// Common patterns to match numeric IDs and other parameters
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		// Match numeric IDs at the end of path segments
		{regexp.MustCompile(`/\d+(/|$)`), "/:id$1"},
		// Match UUID-like patterns
		{regexp.MustCompile(`/[0-9a-fA-F-]{32,}/`), "/:id/"},
		// Match other alphanumeric parameters
		// {regexp.MustCompile(`/[a-zA-Z0-9_-]{8,}/`), "/:id/"},
	}

	normalized := url
	for _, pattern := range patterns {
		normalized = pattern.regex.ReplaceAllString(normalized, pattern.replacement)
	}

	return normalized
}

func (m *AuthMiddleware) getPermissionKeyByUrl(reqPath, platform string) string {
	_ = strings.ToLower(platform)
	// 动态配置名（如「问候词」）在路径中，归一成 :name 再查表
	if strings.HasPrefix(reqPath, "GET|/api/conf/ai/name/") {
		if perm := apiUrlPermissionMap["GET|/api/conf/ai/name/:name"]; perm != "" {
			return perm
		}
	}
	normalizedPath := normalizeURL(reqPath)
	return apiUrlPermissionMap[normalizedPath]
}

// empHasAnyMappedPermission 当映射值为 "A|B|C" 时，员工拥有其中任意一个即满足。
func empHasAnyMappedPermission(permissionMap map[string]bool, required string) bool {
	if strings.TrimSpace(required) == "" {
		return true
	}
	for _, p := range strings.Split(required, "|") {
		p = strings.TrimSpace(p)
		if p != "" && permissionMap[p] {
			return true
		}
	}
	return false
}

// checkUserPermissionByEmpCode 根据员工编码检查用户权限
// 从 Redis 读取权限列表进行判断
func (m *AuthMiddleware) checkUserPermissionByEmpCode(ctx context.Context, clientId string, empCode string, role string, reqPath string, platform string) error {
	// 如果 empCode 为空，返回错误
	if empCode == "" {
		logx.Errorf("员工编码为空，无法检查权限")
		return errors.New("员工编码为空，无法检查权限")
	}
	if strings.TrimSpace(clientId) == "" {
		logx.Errorf("clientId为空，无法检查权限")
		return errors.New("clientId为空，无法检查权限")
	}

	normalizedReq := normalizeURL(reqPath)
	if isFrPlatformRootOperator(clientId, empCode) {
		return nil
	}
	if isAdminTenantSuperadminRestrictedPath(reqPath) {
		if allowAdminTenantSuperadmin(clientId, role) {
			return nil
		}
		logx.Infof("MySQL 备份接口拒绝: clientId=%s role=%s path=%s", clientId, role, reqPath)
		return errors.New("权限不足，仅 admin 租户的 superadmin 角色可访问")
	}
	if _, exclusive := platformRootExclusiveAPI[normalizedReq]; exclusive {
		logx.Infof("平台独占接口拒绝: empCode=%s clientId=%s path=%s", empCode, clientId, reqPath)
		return errors.New("权限不足，无法访问")
	}

	// 从 Redis 读取权限列表
	permissionsKey := fmt.Sprintf("user:permissions:%s:%s", clientId, empCode)
	permissionsJSON, err := m.RedisClient.Get(permissionsKey)
	if err != nil {
		logx.Errorf("从Redis读取权限列表失败: %v, empCode: %s", err, empCode)
		return fmt.Errorf("从Redis读取权限列表失败: %v", err)
	}

	if permissionsJSON == "" {
		logx.Errorf("权限列表为空，empCode: %s", empCode)
		return errors.New("权限列表不存在，请重新登录" + empCode)
	}

	// 解析权限列表
	var empPermissions []string
	err = json.Unmarshal([]byte(permissionsJSON), &empPermissions)
	if err != nil {
		logx.Errorf("解析权限列表失败: %v, empCode: %s", err, empCode)
		return fmt.Errorf("解析权限列表失败: %v", err)
	}

	// 创建权限映射以便快速查找
	permissionMap := make(map[string]bool)
	for _, perm := range empPermissions {
		permissionMap[perm] = true
	}

	// 获取该路径需要的权限
	permission := m.getPermissionKeyByUrl(reqPath, platform)

	// 如果不需要特定权限，则允许访问
	if permission == "" {
		return nil
	}

	if !frPermissionRequirementValid(permission) {
		logx.Errorf("中间件权限码未在 fr_permissions_insert 白名单: path=%s required=%s", reqPath, permission)
		return errors.New("权限配置错误")
	}

	if empHasAnyMappedPermission(permissionMap, permission) {
		return nil
	}

	// 没有匹配的权限
	logx.Infof("用户权限不足，empCode: %s, reqPath: %s, requiredPermission: %s", empCode, reqPath, permission)
	return errors.New("权限不足，无法访问")
}

// checkPermissionCodeByUserId 检查用户是否拥有指定的权限编码
func (m *AuthMiddleware) checkPermissionCodeByUserId(ctx context.Context, userId int64, permissionCode string) (bool, error) {
	// 查询用户是否拥有该权限编码
	permissions, err := m.PermissionModel.FindPermissionCodeByUserId(ctx, userId, permissionCode)
	if err != nil {
		// 如果是 ErrNotFound，说明用户没有该权限
		if err == model.ErrNotFound {
			return false, nil
		}
		logx.Errorf("检查用户权限编码失败: %v, userId: %d, permissionCode: %s", err, userId, permissionCode)
		return false, err
	}

	// 检查是否包含完全匹配的权限编码
	for _, perm := range permissions {
		if perm == permissionCode {
			return true, nil
		}
	}

	return false, nil
}

// GET /api/user/message/count
