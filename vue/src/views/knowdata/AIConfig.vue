<template>
  <div class="ai-config">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>AI 提示词配置</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增配置
          </el-button>
        </div>
      </template>

      

      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="名称">
          <el-input v-model="searchForm.name" placeholder="请输入名称" clearable />
        </el-form-item>
        <el-form-item label="知识库">
          <el-select
            v-model="searchForm.documentCode"
            placeholder="请选择知识库"
            clearable
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
        <!-- <el-table-column prop="id" label="ID" width="80" /> -->
        <el-table-column prop="documentCode" label="知识库" width="150">
          <template #default="{ row }">
            {{ getDocumentTypeName(row.documentCode) }}
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="value" label="值">
          <template #default="{ row }">
            <span class="value-ellipsis" :title="row.value">
              {{ row.value }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column prop="createdBy" label="创建人" width="120" />
        <el-table-column prop="updatedAt" label="更新时间" width="180" />
        <el-table-column prop="updatedBy" label="修改人" width="80" />
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
            </el-button>
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

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="pagination.total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
      <!-- 配置说明 -->
      <el-alert
        type="info"
        :closable="false"
        
        style="margin-top: 20px"
      >
        <template #title>
          <div class="config-description">
            <div class="description-title">配置项说明：</div>
            <div class="description-item">
              <strong>检索提示词：</strong>拿到 RAG 数据后拼接发给 LLM 的信息格式
            </div>
            <div class="description-item">
              <strong>角色提示词：</strong>给 LLM 的提示词
            </div>
            <div class="description-item">
              <strong>问候词：</strong>进入打招呼的语句
            </div>
             
          </div>
        </template>
      </el-alert>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="90vw"
      top="5vh"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="80px"
      >
        <el-form-item label="名称" prop="name">
          <el-select
            v-model="form.name"
            placeholder="请选择配置名称"
            style="width: 100%"
            :disabled="isEdit"
          >
            <el-option
              v-for="item in configNameOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="知识库" prop="documentCode">
          <el-select
            v-model="form.documentCode"
            placeholder="请选择知识库（可选）"
            clearable
            style="width: 100%"
          >
            <el-option
              v-for="item in documentsTypeList"
              :key="item.code"
              :label="item.name"
              :value="item.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="值" prop="value" class="editor-form-item">
          <codemirror
            v-model="form.value"
            :extensions="cmExtensions"
            :autofocus="true"
            :indent-with-tab="true"
            :tab-size="2"
            style="height: 70vh; width: 100%; border: 1px solid #dcdfe6; border-radius: 4px;"
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
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import { Codemirror } from 'vue-codemirror'
import { javascript } from '@codemirror/lang-javascript'
import { oneDark } from '@codemirror/theme-one-dark'
import {
  listAIConfig,
  createAIConfig,
  updateAIConfig,
  deleteAIConfig,
  listDocumentsType
} from '@/api/knowdata'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const formRef = ref(null)
const tableData = ref([])
const isEdit = ref(false)
const documentsTypeList = ref([])

// AI 配置名称选项
const configNameOptions = [
  { label: '检索提示词', value: '检索提示词' },
  { label: '角色提示词', value: '角色提示词' },
  { label: '问候词', value: '问候词' }
]

const searchForm = reactive({
  name: '',
  documentCode: ''
})

const form = reactive({
  id: null,
  name: '',
  value: '',
  documentCode: ''
})

const rules = {
  name: [
    { required: true, message: '请输入名称', trigger: 'blur' }
  ],
  value: [
    { required: true, message: '请输入值', trigger: 'blur' }
  ]
}

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const cmExtensions = [javascript({ jsx: true, typescript: true }), oneDark]

const loadData = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      name: searchForm.name || ''
    }
    if (searchForm.documentCode) {
      params.documentCode = searchForm.documentCode
    }
    const res = await listAIConfig(params)
    if (res.code === 200 && res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  Object.assign(searchForm, {
    name: '',
    documentCode: ''
  })
  handleSearch()
}

const handleAdd = () => {
  isEdit.value = false
  Object.assign(form, {
    id: null,
    name: '',
    value: '',
    documentCode: ''
  })
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  Object.assign(form, {
    id: row.id,
    name: row.name,
    value: row.value,
    documentCode: row.documentCode || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return

  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        let res
        if (isEdit.value) {
          res = await updateAIConfig(form.id, {
            id: form.id,
            name: form.name,
            value: form.value,
            documentCode: form.documentCode || ''
          })
        } else {
          res = await createAIConfig({
            name: form.name,
            value: form.value,
            documentCode: form.documentCode || ''
          })
        }
        if (res.code === 200) {
          ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
          dialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.message || res.msg || (isEdit.value ? '更新失败' : '创建失败'))
        }
      } catch (error) {
        ElMessage.error(isEdit.value ? '更新失败，请稍后重试' : '创建失败，请稍后重试')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除这条配置吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteAIConfig(row.id)
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.message || res.msg || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败，'+error.message)
    }
  }).catch(() => {})
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const dialogTitle = computed(() => {
  return isEdit.value ? '编辑配置' : '新增配置'
})

const getDocumentTypeName = (code) => {
  if (!code) return '-'
  const docType = documentsTypeList.value.find(item => item.code === code)
  return docType ? docType.name : code
}

const loadDocumentsTypeList = async () => {
  try {
    const res = await listDocumentsType({})
    if (res.code === 200 && res.data) {
      documentsTypeList.value = res.data.list || []
    }
  } catch (error) {
    console.error('加载知识库列表失败', error)
  }
}

onMounted(() => {
  loadDocumentsTypeList()
  loadData()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 18px;
  font-weight: 500;
}

.search-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.value-ellipsis {
  display: inline-block;
  max-width: 420px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.config-description {
  line-height: 1.8;
}

.description-title {
  font-weight: 600;
  margin-bottom: 8px;
  color: #303133;
}

.description-item {
  margin-bottom: 4px;
  color: #606266;
}

.description-item strong {
  color: #409eff;
}

.editor-form-item {
  align-items: stretch;
}

.editor-form-item :deep(.el-form-item__content) {
  flex: 1;
}

.editor-form-item :deep(.cm-editor) {
  width: 100%;
  height: 70vh;
  max-height: 70vh;
}

.editor-form-item :deep(.cm-scroller) {
  overflow: auto;
}
</style>


