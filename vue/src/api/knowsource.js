import request from "@/utils/request";

// ==================== 登录相关 ====================
export function getCaptcha() {
  return request({
    url: "/v1/user/captcha",
    method: "get",
  });
}

export function login(data) {
  return request({
    url: "/knowsource/login",
    method: "post",
    data,
  });
}

export function oaLogin(data) {
  return request({
    url: "/knowsource/oa/login",
    method: "post",
    data,
  });
}

export function refreshToken() {
  return request({
    url: "/knowsource/refresh/token",
    method: "post",
  });
}

// ==================== 企业账户（租户）自助注册/验证 ====================
export function tenantRegister(data) {
  return request({
    url: "/knowsource/tenant/register",
    method: "post",
    data,
  });
}

export function tenantVerifyEmail(data) {
  return request({
    url: "/knowsource/tenant/verify-email",
    method: "post",
    data,
  });
}

export function sendVerificationCode(data) {
  return request({
    url: "/knowsource/tenant/send-verification-code",
    method: "post",
    data,
  });
}

// ==================== Client 管理（系统管理）====================
export function adminClientCreate(data) {
  return request({
    url: "/knowsource/admin/client/create",
    method: "post",
    data,
  });
}

export function adminClientDelete(data) {
  return request({
    url: "/knowsource/admin/client/delete",
    method: "post",
    data,
  });
}

export function adminClientList(data) {
  return request({
    url: "/knowsource/admin/client/list",
    method: "post",
    data,
  });
}

export function changePassword(data) {
  return request({
    url: "/knowsource/change-password",
    method: "post",
    data,
  });
}

// ==================== 员工知识库权限 ====================
export function createEmpDocumentType(data) {
  return request({
    url: "/knowsource/emp-document-type/create",
    method: "post",
    data,
  });
}

export function deleteEmpDocumentType(data) {
  return request({
    url: "/knowsource/emp-document-type/delete",
    method: "post",
    data,
  });
}

export function listEmpDocumentType(data) {
  return request({
    url: "/knowsource/emp-document-type/list",
    method: "post",
    data,
  });
}

export function listEmpDocumentTypeGroup(data) {
  return request({
    url: "/knowsource/emp-document-type/list/group",
    method: "post",
    data,
  });
}

// ==================== 部门文档类型绑定 ====================
export function createDeptDocumentType(data) {
  return request({
    url: "/knowsource/dept-document-type/create",
    method: "post",
    data,
  });
}

export function deleteDeptDocumentType(data) {
  return request({
    url: "/knowsource/dept-document-type/delete",
    method: "post",
    data,
  });
}

export function listDeptDocumentType(data) {
  return request({
    url: "/knowsource/dept-document-type/list",
    method: "post",
    data,
  });
}

// ==================== 员工列表 ====================
export function listEmp(data) {
  return request({
    url: "/knowsource/emp/list",
    method: "post",
    data,
  });
}

// ==================== 员工增删改 ====================
export function createEmp(data) {
  return request({
    url: "/knowsource/emp/create",
    method: "post",
    data,
  });
}

export function updateEmp(data) {
  return request({
    url: "/knowsource/emp/update",
    method: "post",
    data,
  });
}

export function deleteEmp(data) {
  return request({
    url: "/knowsource/emp/delete",
    method: "post",
    data,
  });
}

// ==================== 部门列表 ====================
export function listDept(data) {
  return request({
    url: "/knowsource/dept/list",
    method: "post",
    data,
  });
}

// ==================== 部门增删改 ====================
export function createDept(data) {
  return request({
    url: "/knowsource/dept/create",
    method: "post",
    data,
  });
}

export function updateDept(data) {
  return request({
    url: "/knowsource/dept/update",
    method: "post",
    data,
  });
}

export function deleteDept(data) {
  return request({
    url: "/knowsource/dept/delete",
    method: "post",
    data,
  });
}

export function moveDept(data) {
  return request({
    url: "/knowsource/dept/move",
    method: "post",
    data,
  });
}

// ==================== 部门树 ====================
export function getDeptTree(data) {
  return request({
    url: "/knowsource/dept/tree",
    method: "post",
    data,
  });
}

// ==================== 我的文档类型 ====================
export function listMyDocumentType() {
  return request({
    url: "/knowsource/my-document-type/list",
    method: "post",
  });
}

// ==================== 管理员重置密码 ====================
export function adminResetPassword(data) {
  return request({
    url: "/knowsource/admin/reset/password",
    method: "post",
    data,
  });
}

// ==================== 管理员设置角色 ====================
export function adminSetRole(data) {
  return request({
    url: "/knowsource/admin/set/role",
    method: "post",
    data,
  });
}

// ==================== 文档路径文件树 ====================
export function getDocPathTree(data) {
  return request({
    url: "/knowsource/doc-path/tree",
    method: "post",
    data,
  });
}

// ==================== 导入 HR 用户和部门 ====================
export function importHrUserDept() {
  return request({
    url: "/knowsource/hr/import/user/dept",
    method: "post",
  });
}

// ==================== 转换文档为 Markdown ====================
export function convertDocumentToMD(data) {
  return request({
    url: "/knowsource/document/convert/to/md",
    method: "post",
    data,
  });
}

// ==================== 转换文档为 ZIP ====================
export function convertDocumentToZIP(data) {
  return request({
    url: "/knowsource/document/convert/to/zip",
    method: "post",
    data,
  });
}

// ==================== 取消文档转换任务 ====================
export function cancelConvertDocument(data) {
  return request({
    url: "/knowsource/document/convert/cancel",
    method: "post",
    data,
  });
}

// ==================== AsyncTask 队列 ====================
export function listAsyncTask(data) {
  return request({
    url: "/knowsource/async-task/list",
    method: "post",
    data,
  });
}

export function cancelAsyncTask(data) {
  return request({
    url: "/knowsource/async-task/cancel",
    method: "post",
    data,
  });
}

export function deleteAsyncTask(data) {
  return request({
    url: "/knowsource/async-task/delete",
    method: "post",
    data,
  });
}

/** 当前租户 async_task 变更水印（Redis），与上次值比较后再决定是否拉列表 */
export function getAsyncTaskWatermark(data = {}) {
  return request({
    url: "/knowsource/async-task/watermark",
    method: "post",
    data,
  });
}

// ==================== 角色管理 ====================
export function createFrRole(data) {
  return request({
    url: "/knowsource/role/create",
    method: "post",
    data,
  });
}

export function getFrRole(data) {
  return request({
    url: "/knowsource/role/get",
    method: "post",
    data,
  });
}

export function updateFrRole(data) {
  return request({
    url: "/knowsource/role/update",
    method: "post",
    data,
  });
}

export function deleteFrRole(data) {
  return request({
    url: "/knowsource/role/delete",
    method: "post",
    data,
  });
}

export function listFrRole(data) {
  return request({
    url: "/knowsource/role/list",
    method: "post",
    data,
  });
}

// ==================== 权限管理 ====================
export function createFrPermission(data) {
  return request({
    url: "/knowsource/permission/create",
    method: "post",
    data,
  });
}

export function getFrPermission(data) {
  return request({
    url: "/knowsource/permission/get",
    method: "post",
    data,
  });
}

export function updateFrPermission(data) {
  return request({
    url: "/knowsource/permission/update",
    method: "post",
    data,
  });
}

export function deleteFrPermission(data) {
  return request({
    url: "/knowsource/permission/delete",
    method: "post",
    data,
  });
}

export function listFrPermission(data) {
  return request({
    url: "/knowsource/permission/list",
    method: "post",
    data,
  });
}

// ==================== 用户角色关联 ====================
export function createFrUserRole(data) {
  return request({
    url: "/knowsource/user-role/create",
    method: "post",
    data,
  });
}

export function getFrUserRole(data) {
  return request({
    url: "/knowsource/user-role/get",
    method: "post",
    data,
  });
}

export function updateFrUserRole(data) {
  return request({
    url: "/knowsource/user-role/update",
    method: "post",
    data,
  });
}

export function deleteFrUserRole(data) {
  return request({
    url: "/knowsource/user-role/delete",
    method: "post",
    data,
  });
}

export function listFrUserRole(data) {
  return request({
    url: "/knowsource/user-role/list",
    method: "post",
    data,
  });
}

// ==================== 角色权限关联 ====================
export function createFrRolePermission(data) {
  return request({
    url: "/knowsource/role-permission/create",
    method: "post",
    data,
  });
}

export function getFrRolePermission(data) {
  return request({
    url: "/knowsource/role-permission/get",
    method: "post",
    data,
  });
}

export function updateFrRolePermission(data) {
  return request({
    url: "/knowsource/role-permission/update",
    method: "post",
    data,
  });
}

export function deleteFrRolePermission(data) {
  return request({
    url: "/knowsource/role-permission/delete",
    method: "post",
    data,
  });
}

export function listFrRolePermission(data) {
  return request({
    url: "/knowsource/role-permission/list",
    method: "post",
    data,
  });
}

// ==================== 文档标签管理 ====================
export function createFrTag(data) {
  return request({
    url: "/knowsource/tag/create",
    method: "post",
    data,
  });
}

export function getFrTag(data) {
  return request({
    url: "/knowsource/tag/get",
    method: "post",
    data,
  });
}

export function updateFrTag(data) {
  return request({
    url: "/knowsource/tag/update",
    method: "post",
    data,
  });
}

export function deleteFrTag(data) {
  return request({
    url: "/knowsource/tag/delete",
    method: "post",
    data,
  });
}

export function listFrTag(data) {
  return request({
    url: "/knowsource/tag/list",
    method: "post",
    data,
  });
}

// ==================== Dify Option 管理 ====================
export function getDifyOption(data) {
  return request({
    url: "/knowsource/dify/option/get",
    method: "post",
    data,
  });
}

export function createDifyOption(data) {
  return request({
    url: "/knowsource/dify/option/create",
    method: "post",
    data,
  });
}

export function updateDifyOption(data) {
  return request({
    url: "/knowsource/dify/option/update",
    method: "post",
    data,
  });
}

export function listDifyOption() {
  return request({
    url: "/knowsource/dify/option/list",
    method: "post",
  });
}

export function deleteDifyOption(data) {
  return request({
    url: "/knowsource/dify/option/delete",
    method: "post",
    data,
  });
}

// ==================== 系统依赖检查 ====================
export function sysCheck() {
  return request({
    url: "/knowsource/sys/check",
    method: "post",
  });
}

// ==================== AI 调用统计 ====================
export function queryAiCallStats(data) {
  return request({
    url: "/knowsource/ai/call/stats/query",
    method: "post",
    data,
  });
}

// ==================== AI Log 管理 ====================
export function getAiLogList() {
  return request({
    url: "/admin/ailog/list",
    method: "get",
  });
}

// ==================== Qdrant 管理 ====================
export function getQdrantCollectionList() {
  return request({
    url: "/admin/qdrant/collection/list",
    method: "get",
  });
}

export function getAiLog(name) {
  return request({
    url: `/admin/ailog/get/${encodeURIComponent(name)}`,
    method: "get",
  });
}

export function deleteAiLog(name) {
  return request({
    url: `/admin/ailog/delete/${encodeURIComponent(name)}`,
    method: "get",
  });
}

export function getRawDocQaLlmLogList() {
  return request({
    url: "/admin/rawdoc/qa-llm-log/list",
    method: "get",
  });
}

export function getRawDocQaLlmLog(name) {
  return request({
    url: `/admin/rawdoc/qa-llm-log/get/${encodeURIComponent(name)}`,
    method: "get",
  });
}

export function deleteRawDocQaLlmLog(name) {
  return request({
    url: `/admin/rawdoc/qa-llm-log/delete/${encodeURIComponent(name)}`,
    method: "get",
  });
}

// ==================== Blog ====================
export function blogListPublic(data) {
  return request({
    url: "/knowsource/blog/list",
    method: "post",
    data,
  });
}

export function blogRead(slug) {
  return request({
    url: `/knowsource/blog/${encodeURIComponent(slug)}`,
    method: "get",
  });
}

export function blogAdminList(data) {
  return request({
    url: "/knowsource/blog/admin/list",
    method: "post",
    data,
  });
}

export function blogAdminGet(data) {
  return request({
    url: "/knowsource/blog/admin/get",
    method: "post",
    data,
  });
}

export function blogAdminCreate(data) {
  return request({
    url: "/knowsource/blog/admin/create",
    method: "post",
    data,
  });
}

export function blogAdminUpdate(data) {
  return request({
    url: "/knowsource/blog/admin/update",
    method: "post",
    data,
  });
}

export function blogAdminDelete(data) {
  return request({
    url: "/knowsource/blog/admin/delete",
    method: "post",
    data,
  });
}

// ==================== 系统统计 ====================
export function getSystemStats() {
  return request({
    url: "/knowsource/system/stats",
    method: "post",
  });
}
