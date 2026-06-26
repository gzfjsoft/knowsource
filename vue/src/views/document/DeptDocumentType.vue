<template>
  <div class="dept-document-type">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门文档类型绑定</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增绑定
          </el-button>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="部门">
          <el-popover
            v-model:visible="searchDeptPopoverVisible"
            placement="bottom-start"
            :width="350"
            trigger="click"
            popper-class="dept-tree-popover"
          >
            <template #reference>
              <el-input
                v-model="searchDeptDisplayName"
                placeholder="请选择部门"
                readonly
                clearable
                style="width: 250px"
                @clear="handleClearSearchDept"
              >
                <template #suffix>
                  <el-icon class="el-input__icon"><ArrowDown /></el-icon>
                </template>
              </el-input>
            </template>
            <div class="tree-wrapper">
              <el-tree
                :data="deptTreeData"
                :props="{ children: 'children', label: 'deptName' }"
                node-key="deptCode"
                :default-expand-all="false"
                class="dept-tree-select"
                @node-click="handleSearchDeptSelect"
              />
            </div>
          </el-popover>
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
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="deptCode" label="部门编码" width="120" />
        <el-table-column prop="deptName" label="部门名称" width="200" />
        <el-table-column prop="documentTypeName" label="知识库名称" width="200" />
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column prop="updatedAt" label="更新时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right">
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

    <!-- 新增对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="新增部门知识库绑定"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="部门" prop="deptCode">
          <el-popover
            v-model:visible="formDeptPopoverVisible"
            placement="bottom-start"
            :width="350"
            trigger="click"
            popper-class="dept-tree-popover"
          >
            <template #reference>
              <el-input
                v-model="formDeptDisplayName"
                placeholder="请选择部门"
                readonly
                clearable
                style="width: 100%"
                @clear="handleClearFormDept"
              >
                <template #suffix>
                  <el-icon class="el-input__icon"><ArrowDown /></el-icon>
                </template>
              </el-input>
            </template>
            <div class="tree-wrapper">
              <el-tree
                :data="deptTreeData"
                :props="{ children: 'children', label: 'deptName' }"
                node-key="deptCode"
                :default-expand-all="false"
                class="dept-tree-select"
                @node-click="handleFormDeptSelect"
              />
            </div>
          </el-popover>
        </el-form-item>
        <el-form-item label="知识库" prop="documentTypeCode">
          <el-select
            v-model="form.documentTypeCode"
            placeholder="请选择知识库"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="docType in documentTypeList"
              :key="docType.code"
              :label="docType.name"
              :value="docType.code"
            />
          </el-select>
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
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, ArrowDown } from '@element-plus/icons-vue'
import {
  listDeptDocumentType,
  createDeptDocumentType,
  deleteDeptDocumentType,
  getDeptTree
} from '@/api/knowsource'
import { listDocumentsType } from '@/api/knowdata'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const formRef = ref(null)
const tableData = ref([])
const deptTreeData = ref([])
const deptTreeLoading = ref(false)
const searchDeptPopoverVisible = ref(false)
const formDeptPopoverVisible = ref(false)
const searchDeptDisplayName = ref('')
const formDeptDisplayName = ref('')
const documentTypeList = ref([])

const searchForm = reactive({
  deptCode: '',
  documentTypeCode: ''
})

const form = reactive({
  deptCode: '',
  documentTypeCode: ''
})

const rules = {
  deptCode: [
    { required: true, message: '请选择部门', trigger: 'change' }
  ],
  documentTypeCode: [
    { required: true, message: '请选择知识库', trigger: 'change' }
  ]
}

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await listDeptDocumentType({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchForm
    })
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
    deptCode: ''
  })
  handleSearch()
}

const handleAdd = () => {
  Object.assign(form, {
    deptCode: '',
    documentTypeCode: ''
  })
  formDeptDisplayName.value = ''
  dialogVisible.value = true
  // 如果知识库列表为空，则加载
  if (documentTypeList.value.length === 0) {
    loadDocumentTypes()
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      submitLoading.value = true
      try {
        const res = await createDeptDocumentType(form)
        if (res.code === 200) {
          ElMessage.success('创建成功')
          dialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.msg || '创建失败')
        }
      } catch (error) {
        ElMessage.error('创建失败，请稍后重试')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除这条记录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteDeptDocumentType({ id: row.id })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        loadData()
      } else {
        ElMessage.error(res.msg || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
}

// 递归查找部门
const findDeptByCode = (tree, code) => {
  for (const node of tree) {
    if (node.deptCode === code) {
      return node
    }
    if (node.children && node.children.length > 0) {
      const found = findDeptByCode(node.children, code)
      if (found) return found
    }
  }
  return null
}

// 更新搜索表单的部门显示名称
const updateSearchDeptDisplay = () => {
  if (searchForm.deptCode && deptTreeData.value.length > 0) {
    const dept = findDeptByCode(deptTreeData.value, searchForm.deptCode)
    searchDeptDisplayName.value = dept ? dept.deptName : ''
  } else {
    searchDeptDisplayName.value = ''
  }
}

// 更新表单的部门显示名称
const updateFormDeptDisplay = () => {
  if (form.deptCode && deptTreeData.value.length > 0) {
    const dept = findDeptByCode(deptTreeData.value, form.deptCode)
    formDeptDisplayName.value = dept ? dept.deptName : ''
  } else {
    formDeptDisplayName.value = ''
  }
}

// 搜索表单部门选择
const handleSearchDeptSelect = (data) => {
  searchForm.deptCode = data.deptCode
  searchDeptDisplayName.value = data.deptName
  searchDeptPopoverVisible.value = false
}

// 清空搜索表单部门
const handleClearSearchDept = () => {
  searchForm.deptCode = ''
  searchDeptDisplayName.value = ''
}

// 表单部门选择
const handleFormDeptSelect = (data) => {
  form.deptCode = data.deptCode
  formDeptDisplayName.value = data.deptName
  formDeptPopoverVisible.value = false
}

// 清空表单部门
const handleClearFormDept = () => {
  form.deptCode = ''
  formDeptDisplayName.value = ''
}

const loadDeptTree = async () => {
  deptTreeLoading.value = true
  try {
    const res = await getDeptTree({})
    if (res.code === 200 && res.data && res.data.tree) {
      deptTreeData.value = res.data.tree
      // 更新显示名称
      updateSearchDeptDisplay()
      updateFormDeptDisplay()
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

const loadDocumentTypes = async () => {
  try {
    const res = await listDocumentsType({
      page: 1,
      pageSize: 1000 // 获取所有知识库
    })
    if (res.code === 200 && res.data && res.data.list) {
      documentTypeList.value = res.data.list || []
    } else {
      documentTypeList.value = []
    }
  } catch (error) {
    ElMessage.error('加载知识库列表失败')
    documentTypeList.value = []
  }
}

onMounted(() => {
  loadData()
  loadDeptTree()
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

.tree-wrapper {
  max-height: 400px;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 5px 0;
}

.tree-wrapper::-webkit-scrollbar {
  width: 8px;
}

.tree-wrapper::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.tree-wrapper::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.tree-wrapper::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

:deep(.dept-tree-select) {
  width: 100%;
}

:deep(.dept-tree-select .el-tree-node__content) {
  height: 32px;
  line-height: 32px;
  padding-right: 10px;
}

:deep(.dept-tree-popover) {
  padding: 8px;
}
</style>

<style>
/* 全局样式，用于 popper-class */
.dept-tree-popover {
  padding: 8px !important;
}

.dept-tree-popover .tree-wrapper {
  max-height: 400px;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 5px 0;
}

.dept-tree-popover .tree-wrapper::-webkit-scrollbar {
  width: 8px;
}

.dept-tree-popover .tree-wrapper::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
}

.dept-tree-popover .tree-wrapper::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 4px;
}

.dept-tree-popover .tree-wrapper::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>

