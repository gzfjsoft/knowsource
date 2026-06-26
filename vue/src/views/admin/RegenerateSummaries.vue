<template>
  <div class="regenerate-summaries">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>重新生成文档概要</span>
        </div>
      </template>
      
      <el-form :model="form" label-width="120px">
        <el-form-item label="文档类型">
          <el-select v-model="form.documentCode" placeholder="选择文档类型（不选择则处理所有类型）" @change="loadAuditedDocuments">
            <el-option label="所有类型" value=""></el-option>
            <el-option 
              v-for="docType in documentTypes" 
              :key="docType.code" 
              :label="docType.name" 
              :value="docType.code"
            ></el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            @click="loadAuditedDocuments" 
            :loading="loading"
            :disabled="loading"
          >
            加载已审核文档
          </el-button>
        </el-form-item>
      </el-form>
      
      <div v-if="documents.length > 0" class="documents-list">
        <el-table :data="documents" style="width: 100%">
          <el-table-column prop="id" label="ID" width="80"></el-table-column>
          <el-table-column prop="fileName" label="文件名" min-width="200"></el-table-column>
          <el-table-column prop="documentCode" label="文档类型" width="120"></el-table-column>
          <el-table-column prop="auditUser" label="审核人" width="100"></el-table-column>
          <el-table-column prop="auditedAt" label="审核时间" width="180">
            <template #default="scope">
              {{ formatTime(scope.row.auditedAt) }}
            </template>
          </el-table-column>
        </el-table>
        
        <div class="batch-actions">
          <el-button 
            type="primary" 
            @click="handleBatchRegenerate" 
            :loading="loading"
            :disabled="loading || documents.length === 0"
          >
            批量重新生成（每5个一组）
          </el-button>
        </div>
      </div>
      
      <div v-if="progressVisible" class="progress-container">
        <el-progress 
          :percentage="progress" 
          :format="progressFormat"
          :status="progressStatus"
        ></el-progress>
        <div class="progress-info">
          <p>总文档数: {{ total }}</p>
          <p>已处理: {{ processed }}</p>
          <p>成功: {{ successCount }}</p>
          <p>失败: {{ failureCount }}</p>
          <p class="message">{{ message }}</p>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listMyDocumentType } from '@/api/knowsource'
import { listRawDocuments } from '@/api/knowdata'
import { useUserStore } from '@/stores/user'

const form = ref({
  documentCode: ''
})

const userStore = useUserStore()
const loading = ref(false)
const progressVisible = ref(false)
const progress = ref(0)
const total = ref(0)
const processed = ref(0)
const successCount = ref(0)
const failureCount = ref(0)
const message = ref('')
const progressStatus = ref('')
const documentTypes = ref([])
const documents = ref([])

const progressFormat = (percentage) => {
  return `${percentage}%`
}

const loadDocumentTypes = async () => {
  try {
    const res = await listMyDocumentType()
    if (res.code === 200 && res.data && res.data.list) {
      documentTypes.value = res.data.list.filter(item => item.isDisabled !== 1)
    }
  } catch (error) {
    console.error('加载文档类型失败:', error)
    ElMessage.error('加载文档类型失败')
  }
}

const loadAuditedDocuments = async () => {
  loading.value = true
  try {
    const res = await listRawDocuments({
      documentCode: form.value.documentCode,
      isAudit: '1',
      page: 1,
      pageSize: 10000
    })
    if (res.code === 200 && res.data && res.data.list) {
      documents.value = res.data.list
      ElMessage.success(`加载成功，共 ${documents.value.length} 个已审核文档`)
    } else {
      ElMessage.error('加载已审核文档失败')
    }
  } catch (error) {
    console.error('加载已审核文档失败:', error)
    ElMessage.error('加载已审核文档失败')
  } finally {
    loading.value = false
  }
}

const handleBatchRegenerate = async () => {
  if (documents.value.length === 0) {
    ElMessage.warning('请先加载已审核文档')
    return
  }

  loading.value = true
  progressVisible.value = true
  progress.value = 0
  total.value = documents.value.length
  processed.value = 0
  successCount.value = 0
  failureCount.value = 0
  message.value = '准备开始处理...'
  progressStatus.value = ''

  // 直接处理所有文档，不分组
  try {
    // 发送 POST 请求
    const response = await fetch('/api/raw-documents/regenerate-summaries', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${userStore.token}`
      },
      body: JSON.stringify({
        documentCode: form.value.documentCode,
        fileIds: documents.value.map(doc => doc.id)
      })
    })

    if (!response.ok) {
      throw new Error('网络响应错误')
    }
    
    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    
    await reader.read().then(function processText({ done, value }) {
      if (done) {
        return
      }
      
      const chunk = decoder.decode(value, { stream: true })
      const lines = chunk.split('\n')
      
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          const dataStr = line.substring(6)
          try {
            const data = JSON.parse(dataStr)
            console.log('收到进度数据:', data)
            // 直接使用后端返回的值
            processed.value = data.processed
            successCount.value = data.successCount
            failureCount.value = data.failureCount
            message.value = data.message
            progress.value = Math.min(100, Math.round((data.processed / total.value) * 100))
          } catch (error) {
            console.error('解析进度数据失败:', error)
          }
        }
      }
      
      return reader.read().then(processText)
    })
  } catch (error) {
    console.error('处理失败:', error)
    failureCount.value = documents.value.length
    processed.value = documents.value.length
    progress.value = 100
    message.value = `处理失败: ${error.message}`
  }

  loading.value = false
  progressStatus.value = 'success'
  message.value = '处理完成'
  ElMessage.success('处理完成')
}

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

onMounted(() => {
  loadDocumentTypes()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.documents-list {
  margin-top: 20px;
}

.batch-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.progress-container {
  margin-top: 20px;
  padding: 10px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background-color: #f5f7fa;
}

.progress-info {
  margin-top: 10px;
  font-size: 14px;
}

.progress-info p {
  margin: 5px 0;
}

.message {
  color: #409eff;
  font-weight: 500;
}
</style>
