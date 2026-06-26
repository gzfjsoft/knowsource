<template>
  <div class="ai-log-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <span>{{ activeLogType === 'chat' ? '对话日志' : '问答提取日志' }}</span>
            <el-radio-group v-model="activeLogType" size="small" @change="handleLogTypeChange">
              <el-radio-button label="chat">对话日志</el-radio-button>
              <el-radio-button label="qa">问答提取日志</el-radio-button>
            </el-radio-group>
          </div>
          <div class="card-actions">
            <el-button
              type="danger"
              :disabled="selectedRows.length === 0"
              :loading="batchDeleting"
              @click="handleBatchDelete"
            >
              <el-icon><Delete /></el-icon>
              批量删除 ({{ selectedRows.length }})
            </el-button>
            <el-button type="primary" @click="loadData" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- 文件列表 -->
      <el-table
        v-loading="loading"
        :data="fileList"
        border
        stripe
        style="width: 100%"
        row-key="name"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column prop="name" label="文件名" width="400" show-overflow-tooltip>
          <template #default="{ row }">
            <span class="file-name" @click="handleView(row.name)">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="size" label="文件大小" width="120">
          <template #default="{ row }">
            {{ formatFileSize(row.size) }}
          </template>
        </el-table-column>
        <el-table-column prop="modified" label="修改时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.modified) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="handleView(row.name)">
              <el-icon><View /></el-icon>
              查看
            </el-button>
            <el-button type="danger" size="small" @click="handleDelete(row.name)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 查看文件内容对话框 -->
      <el-dialog
        v-model="viewDialogVisible"
        :title="`查看日志: ${currentFileName}`"
        width="80%"
        :before-close="handleDialogClose"
      >
        <div class="log-content">
          <el-scrollbar height="500px">
            <pre class="log-text">{{ logContent }}</pre>
          </el-scrollbar>
        </div>
        <template #footer>
          <el-button @click="viewDialogVisible = false">关闭</el-button>
          <el-button type="primary" @click="handleCopyContent">复制内容</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, View, Delete } from '@element-plus/icons-vue'
import {
  getAiLogList,
  getAiLog,
  deleteAiLog,
  getRawDocQaLlmLogList,
  getRawDocQaLlmLog,
  deleteRawDocQaLlmLog
} from '@/api/knowsource'

const loading = ref(false)
const batchDeleting = ref(false)
const activeLogType = ref('chat')
const fileList = ref([])
const selectedRows = ref([])
const viewDialogVisible = ref(false)
const currentFileName = ref('')
const logContent = ref('')

const handleSelectionChange = (rows) => {
  selectedRows.value = rows
}

// 加载文件列表
const loadData = async () => {
  loading.value = true
  try {
    const res =
      activeLogType.value === 'qa'
        ? await getRawDocQaLlmLogList()
        : await getAiLogList()
    if (res.code === 200 && res.data && res.data.list) {
      // list 每项为 { name, datetime }（datetime 为 Unix 时间戳）
      fileList.value = res.data.list.map((item) => {
        const fileName = item.name || ''
        const parts = fileName.replace('.log.txt', '').split('_')
        const timestamp = parts[0]
        const keys = parts.slice(1).join('_')
        const modified = item.datetime != null ? new Date(item.datetime * 1000) : new Date()
        return {
          name: fileName,
          keys: keys || '',
          timestamp: timestamp,
          modified,
          size: 0
        }
      })
      // 后端已按修改时间降序，前端无需再排序
    } else {
      ElMessage.error(res.message || '获取文件列表失败')
    }
  } catch (error) {
    console.error('加载文件列表失败:', error)
    ElMessage.error('加载文件列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 查看文件内容
const handleView = async (fileName) => {
  currentFileName.value = fileName
  viewDialogVisible.value = true
  loading.value = true
  
  try {
    // 移除 .log.txt 后缀作为参数
    const nameWithoutExt = fileName.replace('.log.txt', '')
    const res =
      activeLogType.value === 'qa'
        ? await getRawDocQaLlmLog(nameWithoutExt)
        : await getAiLog(nameWithoutExt)
    if (res.code === 200 && res.data) {
      logContent.value = res.data.content || ''
    } else {
      ElMessage.error(res.message || '获取文件内容失败')
      logContent.value = ''
    }
  } catch (error) {
    console.error('获取文件内容失败:', error)
    ElMessage.error('获取文件内容失败: ' + (error.message || '未知错误'))
    logContent.value = ''
  } finally {
    loading.value = false
  }
}

// 删除文件
const handleDelete = (fileName) => {
  ElMessageBox.confirm(`确定要删除文件 "${fileName}" 吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      // 移除 .log.txt 后缀作为参数
      const nameWithoutExt = fileName.replace('.log.txt', '')
      const res =
        activeLogType.value === 'qa'
          ? await deleteRawDocQaLlmLog(nameWithoutExt)
          : await deleteAiLog(nameWithoutExt)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      console.error('删除文件失败:', error)
      ElMessage.error('删除文件失败: ' + (error.message || '未知错误'))
    }
  }).catch(() => {})
}

// 批量删除
const handleBatchDelete = () => {
  if (selectedRows.value.length === 0) return
  const names = selectedRows.value.map(r => r.name)
  ElMessageBox.confirm(`确定要删除选中的 ${names.length} 个文件吗？`, '批量删除', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    batchDeleting.value = true
    let success = 0
    let fail = 0
    for (const row of selectedRows.value) {
      try {
        const nameWithoutExt = row.name.replace('.log.txt', '')
        const res =
          activeLogType.value === 'qa'
            ? await deleteRawDocQaLlmLog(nameWithoutExt)
            : await deleteAiLog(nameWithoutExt)
        if (res.code === 200) success++
        else fail++
      } catch {
        fail++
      }
    }
    batchDeleting.value = false
    if (fail === 0) {
      ElMessage.success(`已成功删除 ${success} 个文件`)
    } else {
      ElMessage.warning(`删除完成：成功 ${success} 个，失败 ${fail} 个`)
    }
    loadData()
  }).catch(() => {})
}

const handleLogTypeChange = () => {
  selectedRows.value = []
  fileList.value = []
  loadData()
}

// 复制内容
const handleCopyContent = async () => {
  try {
    await navigator.clipboard.writeText(logContent.value)
    ElMessage.success('内容已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败')
  }
}

// 关闭对话框
const handleDialogClose = () => {
  viewDialogVisible.value = false
  currentFileName.value = ''
  logContent.value = ''
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '-'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// 格式化时间
const formatTime = (date) => {
  if (!date) return '-'
  const d = new Date(date)
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 18px;
  font-weight: 500;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.file-name {
  color: #409eff;
  cursor: pointer;
  text-decoration: underline;
}

.file-name:hover {
  color: #66b1ff;
}

.log-content {
  margin: 20px 0;
}

.log-text {
  margin: 0;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-wrap: break-word;
  color: #303133;
}
</style>
