<template>
  <div class="check-raw-documents-file-exists">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>检查文档文件是否存在</span>
          <el-button type="primary" @click="handleCheck" :loading="loading">
            <el-icon><Refresh /></el-icon>
            开始检查
          </el-button>
        </div>
      </template>

      <!-- 统计信息 -->
      <div v-if="checkResult" class="statistics">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-value">{{ checkResult.total }}</div>
                <div class="stat-label">总文档数</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-value" style="color: #67C23A;">{{ checkResult.existsCount }}</div>
                <div class="stat-label">存在</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-value" style="color: #F56C6C;">{{ checkResult.missingCount }}</div>
                <div class="stat-label">缺失</div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="6">
            <el-card class="stat-card">
              <div class="stat-content">
                <div class="stat-value" style="color: #409EFF;">{{ existenceRate }}%</div>
                <div class="stat-label">存在率</div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form" style="margin-top: 20px">
        <el-form-item label="知识库">
          <el-select
            v-model="searchForm.documentCode"
            placeholder="请选择知识库"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="item in documentsTypeList"
              :key="item.code"
              :label="item.name"
              :value="item.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="文件名">
          <el-input v-model="searchForm.fileName" placeholder="请输入文件名" clearable />
        </el-form-item>
        <el-form-item label="状态">
          <el-select
            v-model="searchForm.exists"
            placeholder="请选择状态"
            clearable
            style="width: 150px"
          >
            <el-option label="存在" :value="true" />
            <el-option label="缺失" :value="false" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <!-- 表格 -->
      <el-table
        v-loading="loading"
        :data="filteredTableData"
        border
        stripe
        style="width: 100%; margin-top: 20px"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="documentCode" label="知识库" width="150">
          <template #default="{ row }">
            {{ getDocumentTypeName(row.documentCode) }}
          </template>
        </el-table-column>
        <el-table-column prop="fileName" label="文件名" width="250" show-overflow-tooltip />
        <el-table-column prop="filePath" label="文件路径" width="300" show-overflow-tooltip />
        <el-table-column prop="fileSize" label="文件大小" width="100">
          <template #default="{ row }">
            {{ formatFileSize(row.fileSize) }}
          </template>
        </el-table-column>
        <el-table-column prop="exists" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.exists ? 'success' : 'danger'">
              {{ row.exists ? '存在' : '缺失' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="actualPath" label="实际路径" width="300" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.actualPath" :title="row.actualPath">{{ row.actualPath }}</span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="error" label="错误信息" width="200" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.error" :title="row.error" class="error-text">{{ row.error }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column v-if="hasPermission('功能-删除文档')" label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="hasPermission('功能-删除文档') && selectedRows.length > 0" class="batch-actions">
        <el-button type="danger" @click="handleBatchDelete">
          批量删除 ({{ selectedRows.length }})
        </el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Search } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { checkRawDocumentsFileExists, deleteRawDocuments, listDocumentsType } from '@/api/knowdata'

const userStore = useUserStore()
const hasPermission = (p) => (userStore.empPermissions || []).includes(p)
const loading = ref(false)
const checkResult = ref(null)
const tableData = ref([])
const documentsTypeList = ref([])
const selectedRows = ref([])

const searchForm = ref({
  documentCode: '',
  fileName: '',
  exists: null
})

// 计算存在率
const existenceRate = computed(() => {
  if (!checkResult.value || checkResult.value.total === 0) {
    return 0
  }
  return ((checkResult.value.existsCount / checkResult.value.total) * 100).toFixed(2)
})

// 过滤表格数据
const filteredTableData = computed(() => {
  let data = tableData.value

  if (searchForm.value.documentCode) {
    data = data.filter(item => item.documentCode === searchForm.value.documentCode)
  }

  if (searchForm.value.fileName) {
    const fileName = searchForm.value.fileName.toLowerCase()
    data = data.filter(item => item.fileName.toLowerCase().includes(fileName))
  }

  if (searchForm.value.exists !== null) {
    data = data.filter(item => item.exists === searchForm.value.exists)
  }

  return data
})

// 获取文档类型名称
const getDocumentTypeName = (code) => {
  const type = documentsTypeList.value.find(item => item.code === code)
  return type ? type.name : code
}

// 格式化文件大小
const formatFileSize = (bytes) => {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

// 开始检查
const handleCheck = async () => {
  loading.value = true
  try {
    const res = await checkRawDocumentsFileExists()
    if (res.code === 200) {
      checkResult.value = res.data
      tableData.value = res.data.list || []
      ElMessage.success('检查完成')
    } else {
      ElMessage.error(res.message || '检查失败')
    }
  } catch (error) {
    console.error('检查失败:', error)
    ElMessage.error('检查失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 搜索
const handleSearch = () => {
  // 搜索逻辑已在 computed 中实现
}

// 重置
const handleReset = () => {
  searchForm.value = {
    documentCode: '',
    fileName: '',
    exists: null
  }
}

// 加载文档类型列表
const loadDocumentsTypeList = async () => {
  try {
    const res = await listDocumentsType({})
    if (res.code === 200 && res.data && res.data.list) {
      documentsTypeList.value = res.data.list
    }
  } catch (error) {
    console.error('加载文档类型列表失败:', error)
  }
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedRows.value = selection
}

// 删除单个记录
const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除这条记录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteRawDocuments({ ids: [row.id] })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        // 从表格数据中移除已删除的项
        const index = tableData.value.findIndex(item => item.id === row.id)
        if (index > -1) {
          tableData.value.splice(index, 1)
          // 更新统计信息
          if (checkResult.value) {
            checkResult.value.total--
            if (row.exists) {
              checkResult.value.existsCount--
            } else {
              checkResult.value.missingCount--
            }
          }
        }
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

// 批量删除
const handleBatchDelete = () => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('请选择要删除的记录')
    return
  }
  
  ElMessageBox.confirm(`确定要删除选中的 ${selectedRows.value.length} 条记录吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const ids = selectedRows.value.map(row => row.id)
      const res = await deleteRawDocuments({ ids })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        // 从表格数据中移除已删除的项
        const deletedIds = new Set(ids)
        const deletedItems = tableData.value.filter(item => deletedIds.has(item.id))
        tableData.value = tableData.value.filter(item => !deletedIds.has(item.id))
        // 更新统计信息
        if (checkResult.value) {
          checkResult.value.total -= deletedItems.length
          deletedItems.forEach(item => {
            if (item.exists) {
              checkResult.value.existsCount--
            } else {
              checkResult.value.missingCount--
            }
          })
        }
        selectedRows.value = []
      } else {
        ElMessage.error(res.message || '删除失败')
      }
    } catch (error) {
      console.error('删除失败:', error)
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

onMounted(() => {
  loadDocumentsTypeList()
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

.statistics {
  margin-bottom: 20px;
}

.stat-card {
  text-align: center;
}

.stat-content {
  padding: 10px 0;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
}

.search-form {
  margin-bottom: 20px;
}

.text-muted {
  color: #909399;
}

.error-text {
  color: #f56c6c;
  font-size: 12px;
}

.batch-actions {
  margin-top: 20px;
  padding: 15px;
  background-color: #f5f7fa;
  border-radius: 4px;
  text-align: right;
}
</style>
