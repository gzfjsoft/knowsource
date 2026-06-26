<template>
  <div class="department-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门管理</span>
          <div class="header-actions">
            <el-button type="primary" @click="openCreateDialog">
              新增部门
            </el-button>
          </div>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="部门编码">
          <el-input v-model="searchForm.deptCode" placeholder="请输入部门编码" clearable />
        </el-form-item>
        <el-form-item label="部门名称">
          <el-input v-model="searchForm.deptName" placeholder="请输入部门名称" clearable />
        </el-form-item>
        <el-form-item label="父部门编码">
          <el-input v-model="searchForm.parentCode" placeholder="请输入父部门编码" clearable />
        </el-form-item>
        <el-form-item label="公司代码">
          <el-input v-model="searchForm.gsdm" placeholder="请输入公司代码" clearable />
        </el-form-item>
        <el-form-item label="级别">
          <el-input-number v-model="searchForm.grade" :min="0" placeholder="请输入级别" clearable />
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
        <el-table-column prop="deptCode" label="部门编码" width="120" />
        <el-table-column prop="deptName" label="部门名称" width="200" />
        <el-table-column prop="parentCode" label="父部门编码" width="120" />
        <el-table-column prop="gsdm" label="公司代码" width="100" />
        <el-table-column prop="grade" label="级别" width="80" />
        <el-table-column prop="endMark" label="结束标记" width="100" />
        <el-table-column prop="kind" label="类型" width="100" />
        <el-table-column prop="b0110" label="备注" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="openEditDialog(row)">编辑</el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 新增/编辑弹窗 -->
      <el-dialog
        v-model="dialogVisible"
        :title="dialogMode === 'create' ? '新增部门' : '编辑部门'"
        width="640px"
        :close-on-click-modal="false"
      >
        <el-form :model="form" label-width="110px">
          <el-form-item label="部门编码" required>
            <el-input v-model="form.deptCode" :disabled="dialogMode === 'edit'" placeholder="请输入部门编码" />
          </el-form-item>
          <el-form-item label="部门名称" required>
            <el-input v-model="form.deptName" placeholder="请输入部门名称" />
          </el-form-item>
          <el-form-item label="父部门编码">
            <el-input v-model="form.parentCode" placeholder="请输入父部门编码" />
          </el-form-item>
          <el-form-item label="公司代码">
            <el-input v-model="form.gsdm" placeholder="请输入公司代码" />
          </el-form-item>
          <el-form-item label="级别">
            <el-input-number v-model="form.grade" :min="0" placeholder="请输入级别" />
          </el-form-item>
          <el-form-item label="结束标记">
            <el-input v-model="form.endMark" placeholder="默认0" />
          </el-form-item>
          <el-form-item label="类型">
            <el-input v-model="form.kind" placeholder="请输入类型" />
          </el-form-item>
          <el-form-item label="备注">
            <el-input v-model="form.b0110" type="textarea" :rows="3" placeholder="请输入备注" />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitLoading" @click="handleSubmit">确定</el-button>
        </template>
      </el-dialog>

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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { listDept, createDept, updateDept, deleteDept } from '@/api/knowsource'

const loading = ref(false)
const tableData = ref([])

const searchForm = reactive({
  deptCode: '',
  deptName: '',
  parentCode: '',
  gsdm: '',
  grade: null
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const dialogVisible = ref(false)
const dialogMode = ref('create') // create | edit
const submitLoading = ref(false)
const form = reactive({
  deptCode: '',
  deptName: '',
  parentCode: '',
  gsdm: '',
  grade: 0,
  endMark: '',
  kind: '',
  b0110: ''
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await listDept({
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
    deptCode: '',
    deptName: '',
    parentCode: '',
    gsdm: '',
    grade: null
  })
  handleSearch()
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const resetForm = () => {
  Object.assign(form, {
    deptCode: '',
    deptName: '',
    parentCode: '',
    gsdm: '',
    grade: 0,
    endMark: '',
    kind: '',
    b0110: ''
  })
}

const openCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}

const openEditDialog = (row) => {
  dialogMode.value = 'edit'
  Object.assign(form, {
    deptCode: row.deptCode || '',
    deptName: row.deptName || '',
    parentCode: row.parentCode || '',
    gsdm: row.gsdm || '',
    grade: row.grade ?? 0,
    endMark: row.endMark || '',
    kind: row.kind || '',
    b0110: row.b0110 || ''
  })
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!form.deptCode || !form.deptName) {
    ElMessage.warning('请填写部门编码和部门名称')
    return
  }
  submitLoading.value = true
  try {
    const payload = {
      deptCode: form.deptCode,
      deptName: form.deptName,
      parentCode: form.parentCode,
      gsdm: form.gsdm,
      grade: form.grade ?? 0,
      endMark: form.endMark,
      kind: form.kind,
      b0110: form.b0110
    }
    const res = dialogMode.value === 'create' ? await createDept(payload) : await updateDept(payload)
    if (res.code === 200) {
      ElMessage.success('操作成功')
      dialogVisible.value = false
      loadData()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (e) {
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除部门 ${row.deptName}（${row.deptCode}）？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  submitLoading.value = true
  try {
    const res = await deleteDept({ deptCode: row.deptCode })
    if (res.code === 200) {
      ElMessage.success('删除成功')
      loadData()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (e) {
    ElMessage.error('删除失败')
  } finally {
    submitLoading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.card-header {
  font-size: 18px;
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.search-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>

