<template>
  <div class="documents-type">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>知识库类型</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增知识库
          </el-button>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="知识库名称">
          <el-input v-model="searchForm.name" placeholder="请输入知识库名称" clearable />
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
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="description" label="描述" />
        <el-table-column prop="tags" label="标签" width="200">
          <template #default="{ row }">
            <div v-if="row.tags && row.tags.length > 0" class="tags-container">
              <el-tag
                v-for="(tag, index) in row.tags"
                :key="index"
                size="small"
                type="info"
                style="margin-right: 5px; margin-bottom: 5px"
              >
                {{ tag }}
              </el-tag>
            </div>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="isDisabled" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.isDisabled === 1 ? 'danger' : 'success'">
              {{ row.isDisabled === 1 ? '已禁止' : '正常' }}
            </el-tag>
          </template>
        </el-table-column>
        <!-- <el-table-column prop="createdAt" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.createdAt) }}
          </template>
        </el-table-column> -->
        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.updatedAt) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="row.isDisabled === 0"
              type="warning"
              size="small"
              @click="handleDisable(row)"
            >
              禁止
            </el-button>
            <el-button
              v-else
              type="success"
              size="small"
              @click="handleEnable(row)"
            >
              允许
            </el-button>
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
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入知识库名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="4"
            placeholder="请输入描述"
          />
        </el-form-item>
        <el-form-item label="关联部门">
          <div class="dept-tree-selector">
            <el-tree
              ref="deptTreeRef"
              :data="deptTreeData"
              :props="{ children: 'children', label: 'deptName' }"
              node-key="deptCode"
              show-checkbox
              :default-expand-all="false"
              :check-strictly="false"
              class="dept-tree-multi-select"
              v-loading="deptTreeLoading"
              @check="handleDeptTreeCheck"
            >
              <template #default="{ node, data }">
                <span class="tree-node">
                  <span class="node-label">{{ data.deptName }}</span>
                  <span class="node-code">{{ data.deptCode }}</span>
                </span>
              </template>
            </el-tree>
          </div>
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
import {
  listDocumentsType,
  createDocumentsType,
  updateDocumentsType,
  deleteDocumentsType
} from '@/api/knowdata'
import { getDeptTree, listDeptDocumentType, createDeptDocumentType, deleteDeptDocumentType } from '@/api/knowsource'
import { ArrowDown } from '@element-plus/icons-vue'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const formRef = ref(null)
const tableData = ref([])
const isEdit = ref(false)
const deptTreeData = ref([])
const deptTreeLoading = ref(false)
const selectedDeptCodes = ref([])
const deptTreeRef = ref(null)

const searchForm = reactive({
  code: '',
  name: ''
})

const form = reactive({
  code: '',
  name: '',
  description: '',
  isDisabled: 0
})

const rules = {
  name: [
    { required: true, message: '请输入知识库名称', trigger: 'blur' }
  ]
}

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await listDocumentsType({})
    if (res.code === 200 && res.data) {
      let list = res.data.list || []
      // 前端过滤
      if (searchForm.name) {
        list = list.filter(item => item.name.includes(searchForm.name))
      }
      pagination.total = list.length
      // 前端分页
      const start = (pagination.page - 1) * pagination.pageSize
      const end = start + pagination.pageSize
      tableData.value = list.slice(start, end)
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
    name: ''
  })
  handleSearch()
}

const handleAdd = async () => {
  isEdit.value = false
  Object.assign(form, {
    code: '', // 后端会自动生成，这里保持为空
    name: '',
    description: ''
  })
  selectedDeptCodes.value = []
  if (deptTreeRef.value) {
    deptTreeRef.value.setCheckedKeys([])
  }
  dialogVisible.value = true
  // 加载部门树
  await loadDeptTree()
}

const handleEdit = async (row) => {
  isEdit.value = true
  Object.assign(form, {
    code: row.code,
    name: row.name,
    description: row.description || '',
    isDisabled: row.isDisabled || 0
  })
  dialogVisible.value = true
  // 加载部门树和已关联的部门
  await loadDeptTree()
  await loadDocumentTypeDepts(row.code)
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        let res
        if (isEdit.value) {
          res = await updateDocumentsType(form)
          if (res.code === 200) {
            // 更新部门关联
            await updateDocumentTypeDepts(form.code)
          }
        } else {
          res = await createDocumentsType(form)
          if (res.code === 200) {
            // 创建成功后，重新加载数据以获取生成的 code
            await loadData()
            // 查找刚创建的知识库（通过 name 匹配）
            const createdDoc = tableData.value.find(item => item.name === form.name)
            if (createdDoc && createdDoc.code && selectedDeptCodes.value.length > 0) {
              // 创建后立即关联部门（如果有选择）
              await updateDocumentTypeDepts(createdDoc.code)
            }
          }
        }
        if (res.code === 200) {
          ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
          dialogVisible.value = false
          // 如果创建时没有关联部门，或者已经关联完成，重新加载数据
          if (isEdit.value || !selectedDeptCodes.value.length) {
            loadData()
          } else {
            // 如果创建时关联了部门，已经在上面的 loadData() 中加载了，这里不再重复加载
          }
        } else {
          ElMessage.error(res.msg || (isEdit.value ? '更新失败' : '创建失败'))
        }
      } catch (error) {
        ElMessage.error(isEdit.value ? '更新失败，请稍后重试' : '创建失败，请稍后重试')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleDisable = async (row) => {
  ElMessageBox.confirm('确定要禁止该知识库吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await updateDocumentsType({
        code: row.code,
        name: row.name,
        description: row.description || '',
        isDisabled: 1
      })
      if (res.code === 200) {
        ElMessage.success('禁止成功')
        loadData()
      } else {
        ElMessage.error(res.msg || '禁止失败')
      }
    } catch (error) {
      ElMessage.error('禁止失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleEnable = async (row) => {
  ElMessageBox.confirm('确定要允许该知识库吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'info'
  }).then(async () => {
    try {
      const res = await updateDocumentsType({
        code: row.code,
        name: row.name,
        description: row.description || '',
        isDisabled: 0
      })
      if (res.code === 200) {
        ElMessage.success('允许成功')
        loadData()
      } else {
        ElMessage.error(res.msg || '允许失败')
      }
    } catch (error) {
      ElMessage.error('允许失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除这条记录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteDocumentsType({ code: row.code })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.message || res.msg || '删除失败')
      }
    } catch (error) {
      const errorMessage = error?.response?.data?.message || error?.message || '删除失败，请稍后重试'
      ElMessage.error(errorMessage)
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
  return isEdit.value ? '编辑知识库' : '新增知识库'
})

// 加载部门树
const loadDeptTree = async () => {
  if (deptTreeData.value.length > 0) return // 已加载过
  
  deptTreeLoading.value = true
  try {
    const res = await getDeptTree({})
    if (res.code === 200 && res.data && res.data.tree) {
      deptTreeData.value = res.data.tree
    } else {
      deptTreeData.value = []
    }
  } catch (error) {
    ElMessage.error('加载部门树失败')
    deptTreeData.value = []
  } finally {
    deptTreeLoading.value = false
  }
}

// 加载知识库已关联的部门
const loadDocumentTypeDepts = async (documentTypeCode) => {
  if (!documentTypeCode) {
    selectedDeptCodes.value = []
    if (deptTreeRef.value) {
      deptTreeRef.value.setCheckedKeys([])
    }
    return
  }
  
  try {
    const res = await listDeptDocumentType({
      documentTypeCode: documentTypeCode,
      page: 1,
      pageSize: 1000 // 获取所有关联的部门
    })
    if (res.code === 200 && res.data && res.data.list) {
      selectedDeptCodes.value = res.data.list.map(item => item.deptCode)
      if (deptTreeRef.value) {
        deptTreeRef.value.setCheckedKeys(selectedDeptCodes.value)
      }
    } else {
      selectedDeptCodes.value = []
      if (deptTreeRef.value) {
        deptTreeRef.value.setCheckedKeys([])
      }
    }
  } catch (error) {
    ElMessage.error('加载关联部门失败')
    selectedDeptCodes.value = []
    if (deptTreeRef.value) {
      deptTreeRef.value.setCheckedKeys([])
    }
  }
}

// 处理部门树选择变化
const handleDeptTreeCheck = (data, checked) => {
  const checkedKeys = deptTreeRef.value.getCheckedKeys()
  selectedDeptCodes.value = checkedKeys
}

// 更新知识库的部门关联
const updateDocumentTypeDepts = async (documentTypeCode) => {
  if (!documentTypeCode) return
  
  try {
    // 获取当前已关联的部门
    const currentRes = await listDeptDocumentType({
      documentTypeCode: documentTypeCode,
      page: 1,
      pageSize: 1000
    })
    
    const currentDeptCodes = []
    if (currentRes.code === 200 && currentRes.data && currentRes.data.list) {
      currentRes.data.list.forEach(item => {
        currentDeptCodes.push(item.deptCode)
      })
    }
    
    // 计算需要新增和删除的部门
    const newDeptCodes = selectedDeptCodes.value || []
    const toAdd = newDeptCodes.filter(code => !currentDeptCodes.includes(code))
    const toDelete = currentDeptCodes.filter(code => !newDeptCodes.includes(code))
    
    // 批量删除
    for (const deptCode of toDelete) {
      const deleteRes = await listDeptDocumentType({
        documentTypeCode: documentTypeCode,
        deptCode: deptCode,
        page: 1,
        pageSize: 1
      })
      if (deleteRes.code === 200 && deleteRes.data && deleteRes.data.list && deleteRes.data.list.length > 0) {
        await deleteDeptDocumentType({ id: deleteRes.data.list[0].id })
      }
    }
    
    // 批量新增
    for (const deptCode of toAdd) {
      await createDeptDocumentType({
        deptCode: deptCode,
        documentTypeCode: documentTypeCode
      })
    }
  } catch (error) {
    console.error('更新部门关联失败:', error)
    // 不显示错误，因为知识库本身已经保存成功了
  }
}

onMounted(() => {
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

.dept-tree-selector {
  width: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
  max-height: 400px;
  overflow-y: auto;
}

.dept-tree-multi-select {
  width: 100%;
}

.tree-node {
  display: flex;
  align-items: center;
  flex: 1;
  font-size: 14px;
}

.node-label {
  font-weight: 500;
  color: #303133;
  margin-right: 8px;
}

.node-code {
  color: #909399;
  font-size: 12px;
}

:deep(.dept-tree-multi-select .el-tree-node__content) {
  height: 32px;
  line-height: 32px;
}

:deep(.dept-tree-multi-select .el-checkbox) {
  margin-right: 8px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}
</style>

