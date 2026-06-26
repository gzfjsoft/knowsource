/**
 * 与 MainLayout.vue 侧栏权限一致：非 admin/superadmin 用户可访问的路由由 fr_permissions 菜单项决定。
 */

function hasTenantMenuAccess(permissions) {
  const has = (k) => permissions.includes(k);
  const corp =
    has('菜单-文档管理') ||
    has('菜单-文档标签') ||
    has('菜单-知识库类型') ||
    has('菜单-库内全文检索');
  const docPerm = has('菜单-按员工授权') || has('菜单-按部门授权');
  const aiOps =
    has('菜单-AI提示词配置') ||
    has('菜单-LLM配置') ||
    has('菜单-AI调用统计') ||
    has('菜单-对话日志') ||
    has('菜单-异步任务队列');
  const org =
    has('菜单-员工管理') ||
    has('菜单-部门管理') ||
    has('菜单-角色管理');
  return corp || docPerm || aiOps || org;
}

/**
 * @param {string} path - router to.path（无 query）
 * @param {string[]} permissions - userInfo.empPermissions
 * @returns {boolean}
 */
export function canNonAdminAccessPath(path, permissions) {
  const list = Array.isArray(permissions) ? permissions : [];
  const has = (k) => list.includes(k);
  const tenant = hasTenantMenuAccess(list);

  if (path === '/change-password') return true;

  if (path === '/dashboard') {
    return has('菜单-首页') || tenant;
  }

  if (path === '/ai-chat' || path.startsWith('/ai-chat/')) {
    return has('菜单-知识库问答') || tenant;
  }

  if (path === '/raw-documents' || path.startsWith('/raw-documents/content/')) {
    return has('菜单-文档管理');
  }

  if (path.startsWith('/md-preview/')) {
    return has('菜单-文档管理') || has('菜单-库内全文检索');
  }

  if (
    path === '/raw-documents-search' ||
    path.startsWith('/raw-documents-search/')
  ) {
    return has('菜单-库内全文检索');
  }

  if (
    path === '/knowdata-documents-type' ||
    path.startsWith('/knowdata-documents-type/')
  ) {
    return has('菜单-知识库类型');
  }

  if (path === '/tag-management' || path.startsWith('/tag-management/')) {
    return has('菜单-文档标签');
  }

  if (path === '/emp' || path.startsWith('/emp/')) {
    return has('菜单-员工管理');
  }

  if (
    path === '/dept' ||
    path.startsWith('/dept/') ||
    path === '/dept-tree' ||
    path.startsWith('/dept-tree/')
  ) {
    return has('菜单-部门管理');
  }

  if (
    path === '/role-management' ||
    path.startsWith('/role-management/')
  ) {
    return has('菜单-角色管理');
  }

  if (
    path === '/emp-document-type' ||
    path.startsWith('/emp-document-type/')
  ) {
    return has('菜单-按员工授权');
  }

  if (
    path === '/document-type-dept' ||
    path.startsWith('/document-type-dept/')
  ) {
    return has('菜单-按部门授权');
  }

  if (path === '/dept-document-type' || path.startsWith('/dept-document-type/')) {
    return has('菜单-按部门授权');
  }

  if (
    path === '/knowdata-ai-config' ||
    path.startsWith('/knowdata-ai-config/')
  ) {
    return has('菜单-AI提示词配置');
  }

  if (path === '/llm-setting' || path.startsWith('/llm-setting/')) {
    return has('菜单-LLM配置');
  }

  if (path === '/ai-call-stats' || path.startsWith('/ai-call-stats/')) {
    return has('菜单-AI调用统计');
  }

  if (
    path === '/ai-log-management' ||
    path.startsWith('/ai-log-management/')
  ) {
    return has('菜单-对话日志');
  }

  if (
    path === '/async-task-queue' ||
    path.startsWith('/async-task-queue/')
  ) {
    return has('菜单-异步任务队列');
  }

  return false;
}
