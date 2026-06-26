<template>
  <div class="dify-option-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>Dify 应用管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增配置
          </el-button>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="应用名称">
          <el-input v-model="searchForm.name" placeholder="请输入应用名称" clearable />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch" :loading="loading">
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
        :data="tableData"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="name" label="应用名称" width="200" />
        <el-table-column prop="description" label="描述" width="200">
          <template #default="{ row }">
            <span class="value-ellipsis" :title="row.description">
              {{ row.description || '-' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="url" label="URL">
          <template #default="{ row }">
            <span class="value-ellipsis" :title="row.url">
              {{ row.url }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="apiKey" label="API Key">
          <template #default="{ row }">
            <span class="value-ellipsis" :title="row.apiKey">
              {{ row.apiKey ? '***' + row.apiKey.slice(-4) : '' }}
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              link
              size="small"
              @click="handleDelete(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="应用名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入应用名称"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="请输入应用描述（可选）"
          />
        </el-form-item>
        <el-form-item label="URL" prop="url">
          <el-input
            v-model="formData.url"
            placeholder="请输入 Dify URL"
          />
        </el-form-item>
        <el-form-item label="API Key" prop="apiKey">
          <el-input
            v-model="formData.apiKey"
            type="password"
            placeholder="请输入 API Key"
            show-password
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus } from '@element-plus/icons-vue'
import {
  listDifyOption,
  getDifyOption,
  createDifyOption,
  updateDifyOption,
  deleteDifyOption
} from '@/api/knowsource'

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref([])
const dialogVisible = ref(false)
const formRef = ref(null)
const isEdit = ref(false)
const currentEditName = ref('')

const searchForm = reactive({
  name: ''
})

const formData = reactive({
  name: '',
  description: '',
  url: '',
  apiKey: ''
})

const formRules = {
  name: [
    { required: true, message: '请输入应用名称', trigger: 'blur' }
  ],
  url: [
    { required: true, message: '请输入 URL', trigger: 'blur' }
  ],
  apiKey: [
    { required: true, message: '请输入 API Key', trigger: 'blur' }
  ]
}

const dialogTitle = computed(() => {
  return isEdit.value ? '编辑配置' : '新增配置'
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await listDifyOption()
    if (res.code === 200) {
      let list = res.data?.list || []
      // 如果有搜索条件，进行过滤
      if (searchForm.name) {
        list = list.filter(item => 
          item.name && item.name.toLowerCase().includes(searchForm.name.toLowerCase())
        )
      }
      tableData.value = list
    } else {
      ElMessage.error(res.message || '获取数据失败')
    }
  } catch (error) {
    ElMessage.error('获取数据失败：' + error.message)
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  loadData()
}

const handleReset = () => {
  searchForm.name = ''
  loadData()
}

const handleAdd = () => {
  isEdit.value = false
  currentEditName.value = ''
  formData.name = ''
  formData.description = ''
  formData.url = ''
  formData.apiKey = ''
  dialogVisible.value = true
}

const handleEdit = async (row) => {
  isEdit.value = true
  currentEditName.value = row.name
  formData.name = row.name
  formData.description = row.description || ''
  formData.url = row.url || ''
  formData.apiKey = '' // 编辑时清空，让用户重新输入
  
  // 获取完整的配置信息
  try {
    const res = await getDifyOption({ name: row.name })
    if (res.code === 200 && res.data) {
      formData.description = res.data.description || ''
      formData.url = res.data.url || ''
      formData.apiKey = res.data.apiKey || ''
    }
  } catch (error) {
    console.error('获取配置详情失败:', error)
    // 如果获取失败，使用列表中的数据
    formData.description = row.description || ''
    formData.url = row.url || ''
  }
  
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该配置吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await deleteDifyOption({ name: row.name })
    if (res.code === 200) {
      ElMessage.success('删除成功')
      loadData()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败：' + error.message)
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        let res
        if (isEdit.value) {
          // 编辑时使用 update API
          res = await updateDifyOption({
            name: formData.name,
            description: formData.description || '',
            url: formData.url,
            apiKey: formData.apiKey
          })
        } else {
          // 创建时使用 create API
          res = await createDifyOption({
            name: formData.name,
            description: formData.description || '',
            url: formData.url,
            apiKey: formData.apiKey
          })
        }
        
        if (res.code === 200) {
          ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
          dialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.message || (isEdit.value ? '更新失败' : '创建失败'))
        }
      } catch (error) {
        ElMessage.error((isEdit.value ? '更新失败' : '创建失败') + '：' + error.message)
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
  isEdit.value = false
  currentEditName.value = ''
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.dify-option-management {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-form {
    margin-bottom: 20px;
  }

  .value-ellipsis {
    display: inline-block;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}
</style>
