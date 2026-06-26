import { createRouter, createWebHistory } from "vue-router";
import { ElMessageBox } from "element-plus";
import { useUserStore } from "@/stores/user";
import {
  getBootstrapGate,
  isInitRedirectDeclined,
  setInitRedirectDeclined,
  clearInitRedirectDeclined,
} from "@/utils/bootstrapGate";
import { canNonAdminAccessPath } from "@/utils/routeMenuAccess";

const routes = [
  {
    path: "/index",
    name: "Index",
    component: () => import("@/views/Index.vue"),
    meta: { requiresAuth: false, title: "首页" },
  },
  {
    path: "/install",
    name: "Install",
    component: () => import("@/views/Install.vue"),
    meta: { requiresAuth: false, title: "安装指南" },
  },
  {
    path: "/init",
    name: "SetupInit",
    component: () => import("@/views/SetupInit.vue"),
    meta: { requiresAuth: false },
  },
  {
    path: "/login",
    name: "Login",
    component: () => import("@/views/Login.vue"),
    meta: { requiresAuth: false },
  },
  {
    path: "/enterprise/register",
    name: "EnterpriseRegister",
    component: () => import("@/views/EnterpriseRegister.vue"),
    meta: { requiresAuth: false },
  },
  {
    path: "/email/verify",
    name: "EmailVerification",
    component: () => import("@/views/EmailVerification.vue"),
    meta: { requiresAuth: false },
  },
  {
    path: "/oa-redirect",
    name: "OARedirect",
    component: () => import("@/views/OARedirect.vue"),
    meta: { requiresAuth: false },
  },
  {
    path: "/blogs",
    name: "BlogListPublic",
    component: () => import("@/views/blog/BlogList.vue"),
    meta: { requiresAuth: false, title: "博客列表" },
  },
  {
    path: "/blog/:slug",
    name: "BlogReadPublic",
    component: () => import("@/views/blog/BlogRead.vue"),
    meta: { requiresAuth: false, title: "博客详情" },
  },
  {
    path: "/",
    component: () => import("@/layout/MainLayout.vue"),
    redirect: "/dashboard",
    meta: { requiresAuth: true },
    children: [
      {
        path: "dashboard",
        name: "Dashboard",
        component: () => import("@/views/Dashboard.vue"),
        meta: { title: "首页" },
      },
      {
        path: "emp",
        name: "Employee",
        component: () => import("@/views/org/EmployeeList.vue"),
        meta: { title: "员工管理" },
      },
      {
        path: "dept",
        name: "Department",
        component: () => import("@/views/org/DeptTree.vue"),
        meta: { title: "部门管理" },
      },
      {
        path: "dept-tree",
        name: "DeptTree",
        component: () => import("@/views/org/DeptTree.vue"),
        meta: { title: "部门树" },
      },
      {
        path: "emp-document-type",
        name: "EmpDocumentType",
        component: () => import("@/views/document/EmpDocumentType.vue"),
        meta: { title: "按员工授权" },
      },
      {
        path: "dept-document-type",
        name: "DeptDocumentType",
        component: () => import("@/views/document/DeptDocumentType.vue"),
        meta: { title: "部门文档类型绑定" },
      },
      {
        path: "document-type-dept",
        name: "DocumentTypeDept",
        component: () => import("@/views/document/DocumentTypeDept.vue"),
        meta: { title: "按部门授权" },
      },
      {
        path: "ai-chat",
        name: "AIChat",
        component: () => import("@/views/ai/AIChat.vue"),
        meta: { title: "知识库问答" },
      },
      {
        path: "change-password",
        name: "ChangePassword",
        component: () => import("@/views/ChangePassword.vue"),
        meta: { title: "修改密码" },
      },
      {
        path: "admin",
        redirect: "/client-management",
      },
      {
        path: "knowdata-documents-type",
        name: "KnowdataDocumentsType",
        component: () => import("@/views/knowdata/DocumentsType.vue"),
        meta: { title: "知识库类型" },
      },
      {
        path: "raw-documents",
        name: "RawDocuments",
        component: () => import("@/views/knowdata/RawDocuments.vue"),
        meta: { title: "文档管理" },
      },
      {
        path: "raw-documents-search",
        name: "RawDocumentsSearch",
        component: () => import("@/views/knowdata/RawDocumentsSearch.vue"),
        meta: { title: "库内全文检索" },
      },
      {
        path: "raw-documents/content/:id",
        name: "RawDocumentContent",
        component: () => import("@/views/knowdata/RawDocumentContent.vue"),
        meta: { title: "文档内容" },
      },
      {
        path: "md-preview/:id",
        name: "MdPreview",
        component: () => import("@/views/MdPreview.vue"),
        meta: { title: "MD 预览" },
      },
      {
        path: "check-raw-documents-file-exists",
        name: "CheckRawDocumentsFileExists",
        component: () =>
          import("@/views/knowdata/CheckRawDocumentsFileExists.vue"),
        meta: { title: "原始文档文件校验", superPlatformOnly: true },
      },
      {
        path: "knowdata-ai-config",
        name: "KnowdataAIConfig",
        component: () => import("@/views/knowdata/AIConfig.vue"),
        meta: { title: "AI 提示词配置" },
      },
      {
        path: "llm-setting",
        name: "LLMSetting",
        component: () => import("@/views/knowdata/LLMSetting.vue"),
        meta: { title: "LLM 配置" },
      },
      {
        path: "doc-path-tree",
        name: "DocPathTree",
        component: () => import("@/views/document/DocPathTree.vue"),
        meta: { title: "服务端文档路径", superPlatformOnly: true },
      },
      {
        path: "role-management",
        name: "RoleManagement",
        component: () => import("@/views/org/RoleManagement.vue"),
        meta: { title: "角色管理" },
      },
      {
        path: "permission-management",
        name: "PermissionManagement",
        component: () => import("@/views/org/PermissionManagement.vue"),
        meta: { title: "全局权限点配置", superPlatformOnly: true },
      },
      {
        path: "tag-management",
        name: "TagManagement",
        component: () => import("@/views/tag/TagManagement.vue"),
        meta: { title: "文档标签" },
      },
      {
        path: "ai-log-management",
        name: "AiLogManagement",
        component: () => import("@/views/admin/AiLogManagement.vue"),
        meta: { title: "对话日志" },
      },
      {
        path: "ai-call-stats",
        name: "AiCallStats",
        component: () => import("@/views/admin/AiCallStats.vue"),
        meta: { title: "AI 调用统计" },
      },
      {
        path: "client-management",
        name: "ClientManagement",
        component: () => import("@/views/system/ClientManagement.vue"),
        meta: { title: "租户与企业管理", superPlatformOnly: true },
      },
      {
        path: "sys-check",
        name: "SysCheck",
        component: () => import("@/views/admin/SysCheck.vue"),
        meta: { title: "运行依赖检查", superPlatformOnly: true },
      },
      {
        path: "qdrant-collection-list",
        name: "QdrantCollectionList",
        component: () => import("@/views/admin/QdrantCollectionList.vue"),
        meta: { title: "Qdrant 集合管理", superPlatformOnly: true },
      },
      {
        path: "regenerate-summaries",
        name: "RegenerateSummaries",
        component: () => import("@/views/admin/RegenerateSummaries.vue"),
        meta: { title: "已审核文档概要重建", superPlatformOnly: true },
      },
      {
        path: "blog-management",
        name: "BlogManagement",
        component: () => import("@/views/admin/BlogManagement.vue"),
        meta: { title: "博客管理", superPlatformOnly: true },
      },
      {
        path: "async-task-queue",
        name: "AsyncTaskQueue",
        component: () => import("@/views/admin/AsyncTaskQueue.vue"),
        meta: { title: "异步任务队列" },
      },
    ],
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const userStore = useUserStore();

  if (to.path !== "/init") {
    const gate = await getBootstrapGate();
    if (!gate.error && gate.networkError === false && gate.appReady === true) {
      clearInitRedirectDeclined();
    }
    if (!gate.error && !gate.networkError && gate.appReady === false) {
      if (!isInitRedirectDeclined()) {
        try {
          await ElMessageBox.confirm(
            "系统尚未完成初始化配置，是否进入配置向导？选择「暂不」将留在当前页面继续浏览。",
            "系统初始化",
            {
              confirmButtonText: "进入配置",
              cancelButtonText: "暂不",
              type: "warning",
              distinguishCancelAndClose: true,
            }
          );
          clearInitRedirectDeclined();
          next("/init");
        } catch {
          setInitRedirectDeclined();
          next();
        }
        return;
      }
    }
  }
  if (to.path === "/init") {
    clearInitRedirectDeclined();
    const gate = await getBootstrapGate();
    if (!gate.error && gate.appReady === true) {
      next("/login");
      return;
    }
  }

  // 未登录访问 / 或 /dashboard：优先引导到对外首页（需要放在 requiresAuth 之前）
  if (!userStore.isLoggedIn && (to.path === "/" || to.path === "/dashboard")) {
    next("/index");
    return;
  }

  // 如果未登录且需要认证，跳转到登录页
  if (to.meta.requiresAuth && !userStore.isLoggedIn) {
    next("/login");
    return;
  }

  // 如果已登录且访问登录页，根据 role 重定向
  if (to.path === "/login" && userStore.isLoggedIn) {
    if (!userStore.isAdmin) {
      next("/ai-chat");
    } else {
      next("/dashboard");
    }
    return;
  }

  // 非 admin/superadmin：默认可进问答与改密；其余页面与侧栏一致，按 empPermissions 放行（如 demo 角色）
  if (userStore.isLoggedIn && !userStore.isAdmin) {
    const p = to.path;
    if (p === "/change-password") {
      next();
      return;
    }
    if (p === "/ai-chat" || p.startsWith("/ai-chat/")) {
      next();
      return;
    }
    if (!canNonAdminAccessPath(p, userStore.empPermissions || [])) {
      next("/ai-chat");
      return;
    }
  }

  // 访问根路径时，根据 role 重定向
  if (to.path === "/" && userStore.isLoggedIn) {
    if (!userStore.isAdmin) {
      next("/ai-chat");
    } else {
      next("/dashboard");
    }
    return;
  }

  // 平台运维：仅 clientId=admin 且 empCode=superadmin（不依赖 fr_permissions）
  if (to.meta.superPlatformOnly && userStore.isLoggedIn) {
    if (!userStore.isPlatformSuperUser) {
      next("/dashboard");
      return;
    }
  }

  next();
});

export default router;
