<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #67C23A;">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.empCount || 0 }}</div>
              <div class="stat-label">员工总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #409EFF;">
              <el-icon><OfficeBuilding /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.deptCount || 0 }}</div>
              <div class="stat-label">部门总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #E6A23C;">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.documentCount || 0 }}</div>
              <div class="stat-label">知识库类型</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon" style="background-color: #F56C6C;">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.sessionCount || 0 }}</div>
              <div class="stat-label">AI 会话数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card v-loading="rawDocsLoading">
          <template #header>
            <div class="doc-overview-header">
              <span class="doc-overview-title">文档概览</span>
              <div class="doc-overview-actions">
                <el-button
                  v-if="hasUploadPermission"
                  type="primary"
                  @click="goQuickUpload"
                >
                  <el-icon><Upload /></el-icon>
                  上传文档
                </el-button>
                <el-button @click="$router.push('/raw-documents')">
                  <el-icon><FolderOpened /></el-icon>
                  文档管理
                </el-button>
              </div>
            </div>
          </template>
          <el-row :gutter="20" align="middle">
            <el-col :xs="24" :sm="6" :md="5">
              <div class="doc-total-block">
                <div class="doc-total-value">{{ rawDocTotal }}</div>
                <div class="doc-total-label">文档总数</div>
              </div>
            </el-col>
            <el-col :xs="24" :sm="18" :md="19">
              <div class="recent-docs-caption">最近上传（最多 10 条）</div>
              <el-table
                :data="recentDocs"
                size="small"
                stripe
                border
                empty-text="暂无文档"
                class="recent-docs-table"
              >
                <el-table-column prop="fileName" label="文件名" min-width="160" show-overflow-tooltip>
                  <template #default="{ row }">
                    <el-link type="primary" @click="goDocContent(row)">
                      {{ row.fileName }}
                    </el-link>
                  </template>
                </el-table-column>
                <el-table-column label="知识库" width="140" show-overflow-tooltip>
                  <template #default="{ row }">
                    {{ docTypeLabel(row.documentCode) }}
                  </template>
                </el-table-column>
                <el-table-column label="标签" width="100" show-overflow-tooltip>
                  <template #default="{ row }">
                    {{ row.tag || '—' }}
                  </template>
                </el-table-column>
                <el-table-column label="审核" width="88">
                  <template #default="{ row }">
                    <el-tag :type="row.isAudit === 1 ? 'success' : 'info'" size="small">
                      {{ row.isAudit === 1 ? '已审核' : '未审核' }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="上传时间" width="168">
                  <template #default="{ row }">
                    {{ formatDocTime(row.createdAt) }}
                  </template>
                </el-table-column>
              </el-table>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>欢迎使用</span>
          </template>
          <div class="welcome-content">
            <p>
              欢迎，{{ userStore.userName || userStore.empCode }}
              <span v-if="userStore.role" class="user-role-badge">
                （{{ getRoleLabel(userStore.role) }}）
              </span>
              ！
            </p>
            <p>当前时间：{{ currentTime }}</p>
          </div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>快速操作</span>
          </template>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/emp')">
              <el-icon><User /></el-icon>
              员工管理
            </el-button>
            <el-button type="success" @click="$router.push('/dept')">
              <el-icon><OfficeBuilding /></el-icon>
              部门管理
            </el-button>
            <el-button type="warning" @click="$router.push('/ai-chat')">
              <el-icon><ChatDotRound /></el-icon>
              知识库问答
            </el-button>
            <el-button
              v-if="hasUploadPermission"
              type="primary"
              plain
              @click="goQuickUpload"
            >
              <el-icon><Upload /></el-icon>
              上传文档
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <div class="copyright">

      <SiteIcpLine theme="light" class="icp-footer" />
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import {
  User,
  OfficeBuilding,
  Document,
  ChatDotRound,
  Upload,
  FolderOpened
} from '@element-plus/icons-vue'
import { getSystemStats, listMyDocumentType } from '@/api/knowsource'
import { listRawDocuments } from '@/api/knowdata'
import SiteIcpLine from '@/components/SiteIcpLine.vue'

const router = useRouter()
const userStore = useUserStore()
const hasUploadPermission = computed(() =>
  (userStore.empPermissions || []).includes('功能-上传文档')
)

const stats = ref({
  empCount: 0,
  deptCount: 0,
  documentCount: 0,
  sessionCount: 0
})

const rawDocsLoading = ref(false)
const rawDocTotal = ref(0)
const recentDocs = ref([])
const documentTypes = ref([])

const currentTime = ref('')
let timeInterval = null

onMounted(() => {
  updateTime()
  timeInterval = setInterval(updateTime, 1000)
  loadStats()
  loadDocumentOverview()
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
})

const updateTime = () => {
  const now = new Date()
  currentTime.value = now.toLocaleString('zh-CN')
}

const loadStats = async () => {
  try {
    const res = await getSystemStats()
    if (res.code === 200 && res.data) {
      stats.value = res.data
    }
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

const loadDocumentOverview = async () => {
  rawDocsLoading.value = true
  try {
    const [typeRes, listRes] = await Promise.all([
      listMyDocumentType(),
      listRawDocuments({
        page: 1,
        pageSize: 10,
        documentCode: '',
        fileName: '',
        tag: '',
        isAudit: ''
      })
    ])
    if (typeRes.code === 200 && typeRes.data?.list) {
      documentTypes.value = (typeRes.data.list || []).filter((item) => item.isDisabled !== 1)
    }
    if (listRes.code === 200 && listRes.data) {
      rawDocTotal.value = Number(listRes.data.total) || 0
      recentDocs.value = listRes.data.list || []
    } else {
      rawDocTotal.value = 0
      recentDocs.value = []
    }
  } catch (e) {
    console.error('加载文档概览失败:', e)
    rawDocTotal.value = 0
    recentDocs.value = []
  } finally {
    rawDocsLoading.value = false
  }
}

const docTypeLabel = (code) => {
  const c = String(code || '').trim()
  if (!c) return '—'
  const item = documentTypes.value.find((d) => d.documentTypeCode === c)
  return item ? item.documentTypeName : c
}

const formatDocTime = (ts) => {
  if (!ts) return '—'
  return new Date(ts * 1000).toLocaleString('zh-CN')
}

const goDocContent = (row) => {
  if (!row?.id) return
  router.push({ name: 'RawDocumentContent', params: { id: row.id } })
}

const goQuickUpload = () => {
  router.push({ path: '/raw-documents', query: { openUpload: '1' } })
}

const getRoleLabel = (role) => {
  const roleMap = {
    admin: '管理员',
    superadmin: '超级管理员',
    manager: '经理',
    user: '普通用户',
    demo: '演示用户'
  }
  return roleMap[role] || role
}
</script>

<style scoped>
.stat-card {
  margin-bottom: 20px;
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 24px;
  margin-right: 15px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.doc-overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.doc-overview-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.doc-overview-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.doc-total-block {
  text-align: center;
  padding: 16px 8px;
}

.doc-total-value {
  font-size: 36px;
  font-weight: bold;
  color: #409eff;
  line-height: 1.2;
}

.doc-total-label {
  margin-top: 8px;
  font-size: 14px;
  color: #909399;
}

.recent-docs-caption {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
}

.recent-docs-table {
  width: 100%;
}

.welcome-content {
  padding: 20px 0;
}

.welcome-content p {
  margin: 10px 0;
  font-size: 16px;
  color: #606266;
}

.user-role-badge {
  color: #409eff;
  font-weight: 500;
  margin: 0 4px;
}

.quick-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.copyright {
  margin-top: 30px;
  text-align: center;
  font-size: 12px;
  color: #909399;
  padding: 20px 0;
}

.copyright .icp-footer {
  margin-top: 10px;
}

.copyright :deep(.site-icp-line--light) {
  color: #909399;
}

.copyright :deep(.site-icp-line--light a) {
  color: #409eff;
}
</style>
