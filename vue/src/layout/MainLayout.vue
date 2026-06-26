<template>
  <el-container class="layout-container">
    <el-aside
      v-if="userStore.showMenu && menuVisible"
      width="200px"
      class="sidebar"
    >
      <div class="logo">
        <h2>{{ appTitle }}</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        class="sidebar-menu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
      >
        <el-menu-item v-if="hasMenuDashboard" index="/dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
        </el-menu-item>
        <el-menu-item v-if="hasMenuAiChat" index="/ai-chat">
          <el-icon><ChatDotRound /></el-icon>
          <span>知识库问答</span>
        </el-menu-item>
        <!-- 原始文档、标签、知识库类型、检索 -->
        <el-sub-menu v-if="hasCorpDocSection" index="corp-doc-manage">
          <template #title>
            <el-icon><FolderOpened /></el-icon>
            <span>知识文档</span>
          </template>
          <el-menu-item v-if="hasMenuRawDocuments" index="/raw-documents">
            <el-icon><FolderOpened /></el-icon>
            <span>文档管理</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuTagManagement" index="/tag-management">
            <el-icon><PriceTag /></el-icon>
            <span>文档标签</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuKnowledgeTypes" index="/knowdata-documents-type">
            <el-icon><Document /></el-icon>
            <span>知识库类型</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuFullTextSearch" index="/raw-documents-search">
            <el-icon><Search /></el-icon>
            <span>库内全文检索</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 员工、部门、角色（RBAC） -->
        <el-sub-menu v-if="hasOrgManageMenu" index="org-emp-dept-role">
          <template #title>
            <el-icon><UserFilled /></el-icon>
            <span>组织与角色</span>
          </template>
          <el-menu-item v-if="hasMenuEmployee" index="/emp">
            <el-icon><User /></el-icon>
            <span>员工管理</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuDepartment" index="/dept">
            <el-icon><OfficeBuilding /></el-icon>
            <span>部门管理</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuRole" index="/role-management">
            <el-icon><Key /></el-icon>
            <span>角色管理</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 谁能访问哪些知识库（与「知识库类型」配置区分） -->
        <el-sub-menu
          v-if="hasDocPermissionSection"
          index="doc-permission-manage"
        >
          <template #title>
            <el-icon><Document /></el-icon>
            <span>知识库授权</span>
          </template>
          <el-menu-item v-if="hasMenuAuthByEmployee" index="/emp-document-type">
            <el-icon><Document /></el-icon>
            <span>按员工授权</span>
          </el-menu-item>
          <!-- <el-menu-item index="/dept-document-type">
            <el-icon><Files /></el-icon>
            <span>部门知识库</span>
          </el-menu-item> -->
          <el-menu-item v-if="hasMenuAuthByDepartment" index="/document-type-dept">
            <el-icon><Document /></el-icon>
            <span>按部门授权</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 系统管理 -->
        <el-sub-menu v-if="hasAiOpsSection" index="system-manage">
          <template #title>
            <el-icon><Setting /></el-icon>
            <span>AI 与运维</span>
          </template>
          <el-menu-item v-if="hasMenuAiPrompt" index="/knowdata-ai-config">
            <el-icon><Setting /></el-icon>
            <span>AI 提示词配置</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuLlmSetting" index="/llm-setting">
            <el-icon><Setting /></el-icon>
            <span>LLM 配置</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuAiCallStats" index="/ai-call-stats">
            <el-icon><DataAnalysis /></el-icon>
            <span>AI 调用统计</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuAiLog" index="/ai-log-management">
            <el-icon><Document /></el-icon>
            <span>对话日志</span>
          </el-menu-item>
          <el-menu-item v-if="hasMenuAsyncTaskQueue" index="/async-task-queue">
            <el-icon><List /></el-icon>
            <span>异步任务队列</span>
          </el-menu-item>
        </el-sub-menu>

        <!-- 平台运维：clientId=admin 且 empCode=superadmin，不走 fr_permissions -->
        <el-sub-menu
          v-if="userStore.isPlatformSuperUser"
          index="platform-ops"
        >
          <template #title>
            <el-icon><Grid /></el-icon>
            <span>平台运维</span>
          </template>
          <el-menu-item index="/client-management">
            <el-icon><Setting /></el-icon>
            <span>租户与企业管理</span>
          </el-menu-item>
          <el-menu-item index="/sys-check">
            <el-icon><Setting /></el-icon>
            <span>运行依赖检查</span>
          </el-menu-item>
          <el-menu-item index="/qdrant-collection-list">
            <el-icon><Setting /></el-icon>
            <span>Qdrant 集合管理</span>
          </el-menu-item>
          <el-menu-item index="/doc-path-tree">
            <el-icon><FolderOpened /></el-icon>
            <span>服务端文档路径</span>
          </el-menu-item>
          <el-menu-item index="/permission-management">
            <el-icon><Key /></el-icon>
            <span>全局权限点配置</span>
          </el-menu-item>
          <el-menu-item index="/check-raw-documents-file-exists">
            <el-icon><Document /></el-icon>
            <span>原始文档文件校验</span>
          </el-menu-item>
          <el-menu-item index="/regenerate-summaries">
            <el-icon><Document /></el-icon>
            <span>已审核文档概要重建</span>
          </el-menu-item>
          <el-menu-item index="/blog-management">
            <el-icon><Document /></el-icon>
            <span>博客管理</span>
          </el-menu-item>
        </el-sub-menu>


      </el-menu>
      <div class="version-info">
        <span>v{{ serverVersion || appVersion }}</span>
      </div>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-tooltip
            v-if="userStore.showMenu"
            :content="menuVisible ? '收起菜单' : '展开菜单'"
            placement="bottom"
          >
            <el-button
              class="menu-toggle-btn"
              :icon="menuVisible ? DArrowLeft : Menu"
              text
              :title="menuVisible ? '收起菜单' : '展开菜单'"
              @click="toggleMenu"
            />
          </el-tooltip>
          <span class="title">{{ pageTitle }}</span>
        </div>
        <div class="header-right">
          <span v-if="userStore.role" class="user-role">
            {{ getRoleLabel(userStore.role) }}
          </span>
          <span
            v-if="userStore.showMenu"
            class="header-enterprise"
            :title="headerEnterpriseLabel"
          >{{ headerEnterpriseLabel }}</span>
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <el-icon><User /></el-icon>
              <span class="header-emp-name">{{
                userStore.userName || userStore.empCode
              }}</span>
              <el-icon class="el-icon--right"><arrow-down /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="changePassword"
                  >修改密码</el-dropdown-item
                >
                <el-dropdown-item command="logout" divided
                  >退出登录</el-dropdown-item
                >
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <el-main class="main-content">
        <router-view />
        <!-- 菜单分割条 -->
        <el-tooltip
          v-if="userStore.showMenu"
          :content="menuVisible ? '收起菜单' : '展开菜单'"
          placement="right"
        >
          <div
            :class="['menu-divider', { 'menu-visible': menuVisible }]"
            @click="toggleMenu"
          ></div>
        </el-tooltip>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed, ref, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useUserStore } from "@/stores/user";
import { ElMessageBox } from "element-plus";
import { listMyDocumentType } from "@/api/knowsource";
import { getSysVersion } from "@/api/knowdata";
import {
  HomeFilled,
  User,
  UserFilled,
  OfficeBuilding,
  Document,
  Files,
  ChatDotRound,
  ArrowDown,
  Menu,
  DArrowLeft,
  FolderOpened,
  Folder,
  Setting,
  Key,
  Connection,
  PriceTag,
  List,
  Search,
  DataAnalysis,
  Grid,
} from "@element-plus/icons-vue";

const route = useRoute();
const router = useRouter();
const userStore = useUserStore();

const activeMenu = computed(() => route.path);
const pageTitle = computed(() => route.meta.title || "首页");
/** 与角色徽章同排、紧挨其后；无则回退 clientId（仅 showMenu 时展示，与员工名下拉分开） */
const headerEnterpriseLabel = computed(() => {
  const name = (userStore.userInfo?.clientName || "").trim();
  const id = (userStore.clientId || "").trim();
  return name || id || "—";
});
const appTitle = computed(() => {
  const name = userStore.userInfo?.clientName || "";
  if (name) return `${name} 知源智库 AI`;
  return "知源智库 AI";
});
const menuVisible = ref(true);
const isAIChatPage = computed(() => route.path === "/ai-chat");
const documentTypeList = ref([]);
const selectedDocumentCode = ref("");
const hasActiveSession = ref(false);
const isDocumentTypeLocked = ref(false);

// 权限判断（仅子菜单 / 功能点，与 fr_permissions 一致，无打包菜单）
const hasPermission = (permission) => {
  const permissions = userStore.empPermissions || [];
  return permissions.includes(permission);
};

const hasCorpDocSection = computed(
  () =>
    hasPermission("菜单-文档管理") ||
    hasPermission("菜单-文档标签") ||
    hasPermission("菜单-知识库类型") ||
    hasPermission("菜单-库内全文检索"),
);

const hasMenuRawDocuments = computed(() => hasPermission("菜单-文档管理"));
const hasMenuTagManagement = computed(() => hasPermission("菜单-文档标签"));
const hasMenuKnowledgeTypes = computed(() => hasPermission("菜单-知识库类型"));
const hasMenuFullTextSearch = computed(() =>
  hasPermission("菜单-库内全文检索"),
);

const hasMenuEmployee = computed(() => hasPermission("菜单-员工管理"));
const hasMenuDepartment = computed(() => hasPermission("菜单-部门管理"));
const hasMenuRole = computed(() => hasPermission("菜单-角色管理"));

/** 至少能看到员工/部门或角色其一则显示「组织与角色」分组 */
const hasOrgManageMenu = computed(
  () =>
    hasMenuEmployee.value ||
    hasMenuDepartment.value ||
    hasMenuRole.value,
);

const hasDocPermissionSection = computed(
  () =>
    hasPermission("菜单-按员工授权") ||
    hasPermission("菜单-按部门授权"),
);

const hasMenuAuthByEmployee = computed(() =>
  hasPermission("菜单-按员工授权"),
);
const hasMenuAuthByDepartment = computed(() =>
  hasPermission("菜单-按部门授权"),
);

const hasAiOpsSection = computed(
  () =>
    hasPermission("菜单-AI提示词配置") ||
    hasPermission("菜单-LLM配置") ||
    hasPermission("菜单-AI调用统计") ||
    hasPermission("菜单-对话日志") ||
    hasPermission("菜单-异步任务队列"),
);

const hasMenuAiPrompt = computed(() =>
  hasPermission("菜单-AI提示词配置"),
);
const hasMenuLlmSetting = computed(() => hasPermission("菜单-LLM配置"));
const hasMenuAiCallStats = computed(() =>
  hasPermission("菜单-AI调用统计"),
);
const hasMenuAiLog = computed(() => hasPermission("菜单-对话日志"));
const hasMenuAsyncTaskQueue = computed(() =>
  hasPermission("菜单-异步任务队列"),
);

/** 拥有任一租户侧栏相关子菜单权限时，首页/问答可与细项联动展示 */
const hasTenantMenuAccess = computed(
  () =>
    hasCorpDocSection.value ||
    hasDocPermissionSection.value ||
    hasAiOpsSection.value ||
    hasOrgManageMenu.value,
);

const hasMenuDashboard = computed(
  () => hasPermission("菜单-首页") || hasTenantMenuAccess.value,
);

const hasMenuAiChat = computed(
  () => hasPermission("菜单-知识库问答") || hasTenantMenuAccess.value,
);

// 监听菜单切换事件
const handleToggleMenu = (event) => {
  menuVisible.value = event.detail.visible;
};

// 切换菜单显示/隐藏
const toggleMenu = () => {
  menuVisible.value = !menuVisible.value;
  // 通知其他组件菜单状态变化
  window.dispatchEvent(
    new CustomEvent("toggle-menu", { detail: { visible: menuVisible.value } }),
  );
};

onMounted(() => {
  window.addEventListener("toggle-menu", handleToggleMenu);
  loadServerVersion();
});

onUnmounted(() => {
  window.removeEventListener("toggle-menu", handleToggleMenu);
});

// 前端版本号（后端未返回时兜底）
const appVersion = "1.0.0";
// 后端 /sys/version 返回的 info 版本号
const serverVersion = ref("");
const loadServerVersion = async () => {
  try {
    const res = await getSysVersion();
    if (res && (res.code === 200 || res.code === 0) && res.info) {
      serverVersion.value = res.info;
    }
  } catch {
    // 忽略错误，使用 appVersion
  }
};

// 获取角色标签
const getRoleLabel = (role) => {
  const roleMap = {
    admin: "管理员",
    superadmin: "超级管理员",
    manager: "经理",
    user: "普通用户",
    demo: "演示用户",
  };
  return roleMap[role] || role;
};

const handleCommand = (command) => {
  if (command === "logout") {
    ElMessageBox.confirm("确定要退出登录吗？", "提示", {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    })
      .then(() => {
        userStore.logout();
        router.push("/login");
      })
      .catch(() => {});
  } else if (command === "changePassword") {
    router.push("/change-password");
  }
};

// 加载知识库列表
const loadDocumentTypes = async () => {
  try {
    const res = await listMyDocumentType();
    if (res.code === 200 && res.data && res.data.list) {
      // 过滤掉已禁止的知识库（isDisabled === 1）
      documentTypeList.value = (res.data.list || []).filter(
        (item) => item.isDisabled !== 1,
      );

      // 从 localStorage 读取之前的选择
      const savedDocumentCode = localStorage.getItem("ai-chat-document-code");
      if (
        savedDocumentCode &&
        documentTypeList.value.some(
          (dt) => dt.documentTypeCode === savedDocumentCode,
        )
      ) {
        selectedDocumentCode.value = savedDocumentCode;
      } else if (documentTypeList.value.length > 0) {
        // 默认选择第一个
        selectedDocumentCode.value = documentTypeList.value[0].documentTypeCode;
        localStorage.setItem(
          "ai-chat-document-code",
          selectedDocumentCode.value,
        );
      }

      // 通知 AIChat 组件更新知识库
      window.dispatchEvent(
        new CustomEvent("document-type-changed", {
          detail: { documentCode: selectedDocumentCode.value },
        }),
      );
    }
  } catch (error) {
    console.error("加载知识库列表失败:", error);
  }
};

// 处理知识库变化
const handleDocumentTypeChange = (value) => {
  if (value) {
    localStorage.setItem("ai-chat-document-code", value);
  } else {
    localStorage.removeItem("ai-chat-document-code");
  }
  // 通知 AIChat 组件更新知识库
  window.dispatchEvent(
    new CustomEvent("document-type-changed", {
      detail: { documentCode: value },
    }),
  );
};

// 监听会话状态变化
const handleSessionStatusChanged = (event) => {
  hasActiveSession.value = event.detail.hasSession || false;
  isDocumentTypeLocked.value = event.detail.isLocked || false;
  // 立即更新知识库显示（包括空字符串）
  if (event.detail.documentCode !== undefined) {
    // 明确处理空字符串的情况
    selectedDocumentCode.value =
      event.detail.documentCode === null ? "" : event.detail.documentCode || "";
    // 如果知识库列表已加载，确保显示
    if (event.detail.documentCode && documentTypeList.value.length > 0) {
      hasActiveSession.value = true;
    }
  }
};

// 监听路由变化，在 AI 问答页面时加载知识库
watch(
  isAIChatPage,
  (isAIChat) => {
    if (isAIChat) {
      loadDocumentTypes();
      window.addEventListener(
        "session-status-changed",
        handleSessionStatusChanged,
      );
    } else {
      documentTypeList.value = [];
      selectedDocumentCode.value = "";
      hasActiveSession.value = false;
      isDocumentTypeLocked.value = false;
      window.removeEventListener(
        "session-status-changed",
        handleSessionStatusChanged,
      );
    }
  },
  { immediate: true },
);

onUnmounted(() => {
  window.removeEventListener(
    "session-status-changed",
    handleSessionStatusChanged,
  );
});
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #304156;
  overflow: hidden;
}

.logo {
  height: 60px;
  line-height: 60px;
  text-align: center;
  background-color: #2b3a4a;
  color: #fff;
}

.logo h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 500;
}

.sidebar-menu {
  border: none;
  height: calc(100vh - 60px - 40px);
  overflow-y: auto;
}

.version-info {
  height: 40px;
  line-height: 40px;
  text-align: center;
  background-color: #2b3a4a;
  color: #909399;
  font-size: 12px;
  border-top: 1px solid #3a4a5a;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
}

.header-left {
  display: flex;
  align-items: center;
}

.menu-toggle-btn {
  margin-right: 8px;
}

.header-left .title {
  font-size: 18px;
  font-weight: 500;
  color: #303133;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-role {
  padding: 4px 12px;
  background-color: #409eff;
  color: #fff;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  color: #606266;
}

.user-info .el-icon--right {
  margin-left: 0;
}

.header-enterprise {
  max-width: 220px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
  color: #606266;
}

.header-emp-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.main-content {
  background-color: #f0f2f5;
  padding: 20px;
  overflow-y: auto;
  position: relative;
}

.menu-divider {
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  z-index: 1000;
  background-color: #dcdfe6;
  cursor: pointer;
}

.menu-divider.menu-visible {
  left: 200px;
}

.menu-divider:hover {
  background-color: #409eff;
  width: 5px;
}
</style>
