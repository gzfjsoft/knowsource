<template>
  <div class="emp-document-type">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>按员工授权</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增绑定
          </el-button>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="员工编码">
          <el-input v-model="searchForm.empCode" placeholder="请输入员工编码" clearable />
        </el-form-item>
        <el-form-item label="员工姓名">
          <el-input v-model="searchForm.empName" placeholder="请输入员工姓名" clearable />
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
        <el-table-column prop="empCode" label="员工编码" width="120" />
        <el-table-column prop="empName" label="员工姓名" width="120" />
        <el-table-column prop="companyName" label="公司名称" width="140" />
        <el-table-column prop="deptName" label="员工部门" width="150" />
        <el-table-column label="知识库" min-width="300">
          <template #default="{ row }">
            <el-tag
              v-for="(name, index) in row.documentTypeNames"
              :key="row.documentTypeCodes[index]"
              type="primary"
              style="margin-right: 8px; margin-bottom: 4px"
              closable
              @close="handleDeleteTag(row, index)"
            >
              {{ name }}
            </el-tag>
            <span v-if="!row.documentTypeNames || row.documentTypeNames.length === 0" style="color: #909399">
              暂无知识库
            </span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              @click="handleEdit(row)"
            >
              编辑
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
      :title="form.empCode ? '编辑员工知识库权限' : '新增员工知识库权限'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="员工" prop="empCode">
          <el-select
            v-model="form.empCode"
            placeholder="请搜索并选择员工"
            filterable
            remote
            :remote-method="searchEmployees"
            :loading="employeeSearchLoading"
            clearable
            style="width: 100%"
            @change="handleEmployeeChange"
          >
            <el-option
              v-for="emp in employeeList"
              :key="emp.empCode"
              :label="`${emp.empName} (${emp.empCode})`"
              :value="emp.empCode"
            >
              <span style="float: left">{{ emp.empName }}</span>
              <span style="float: right; color: #8492a6; font-size: 13px">{{ emp.empCode }}</span>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item label="知识库" prop="documentTypeCodes">
          <el-select
            v-model="form.documentTypeCodes"
            placeholder="请选择知识库（可多选）"
            filterable
            multiple
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
import { Plus, Search, Refresh } from '@element-plus/icons-vue'
import {
  listEmpDocumentType,
  listEmpDocumentTypeGroup,
  createEmpDocumentType,
  deleteEmpDocumentType,
  listEmp
} from '@/api/knowsource'
import { listDocumentsType } from '@/api/knowdata'

const loading = ref(false)
const submitLoading = ref(false)
const dialogVisible = ref(false)
const formRef = ref(null)
const tableData = ref([])
const documentTypeList = ref([])
const employeeList = ref([])
const employeeSearchLoading = ref(false)
const employeeSearchKeyword = ref('')

const searchForm = reactive({
  empCode: '',
  empName: '',
  documentTypeCode: ''
})

const form = reactive({
  empCode: '',
  documentTypeCodes: []
})

const rules = {
  empCode: [
    { required: true, message: '请输入员工编码', trigger: 'blur' }
  ],
  documentTypeCodes: [
    {
      validator: (rule, value, callback) => {
        if (!Array.isArray(value) || value.length === 0) {
          callback(new Error('请选择知识库'))
          return
        }
        callback()
      },
      trigger: 'change'
    }
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
    const res = await listEmpDocumentTypeGroup({
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
    empCode: '',
    empName: ''
  })
  handleSearch()
}

const handleAdd = () => {
  Object.assign(form, {
    empCode: '',
    documentTypeCodes: []
  })
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
        const empCode = (form.empCode || '').trim()
        const codes = Array.isArray(form.documentTypeCodes) ? form.documentTypeCodes : []
        
        // 如果是编辑模式，需要先获取当前已有的绑定
        let existingCodes = []
        if (empCode) {
          try {
            const listRes = await listEmpDocumentType({
              page: 1,
              pageSize: 1000,
              empCode: empCode
            })
            if (listRes.code === 200 && listRes.data && listRes.data.list) {
              existingCodes = listRes.data.list.map(item => item.documentTypeCode)
            }
          } catch (e) {
            console.warn('获取已有绑定失败:', e)
          }
        }
        
        // 需要新增的
        const toAdd = codes.filter(code => !existingCodes.includes(code))
        // 需要删除的
        const toDelete = existingCodes.filter(code => !codes.includes(code))
        
        let addSuccessCount = 0
        let deleteSuccessCount = 0
        const failed = []

        // 删除不再需要的绑定
        if (toDelete.length > 0) {
          for (const code of toDelete) {
            try {
              const listRes = await listEmpDocumentType({
                page: 1,
                pageSize: 100,
                empCode: empCode,
                documentTypeCode: code
              })
              
              if (listRes.code === 200 && listRes.data && listRes.data.list && listRes.data.list.length > 0) {
                const record = listRes.data.list.find(
                  item => item.empCode === empCode && item.documentTypeCode === code
                )
                
                if (record) {
                  const deleteRes = await deleteEmpDocumentType({ id: record.id })
                  if (deleteRes.code === 200) {
                    deleteSuccessCount++
                  } else {
                    failed.push({ code, message: deleteRes.msg || deleteRes.message || '删除失败' })
                  }
                }
              }
            } catch (e) {
              failed.push({ code, message: e?.message || '删除失败' })
            }
          }
        }

        // 新增绑定
        for (const code of toAdd) {
          try {
            const res = await createEmpDocumentType({
              empCode,
              documentTypeCode: code
            })
            if (res.code === 200) {
              addSuccessCount++
            } else {
              failed.push({ code, message: res.msg || res.message || '创建失败' })
            }
          } catch (e) {
            failed.push({ code, message: e?.message || '创建失败' })
          }
        }

        if (addSuccessCount > 0 || deleteSuccessCount > 0) {
          const messages = []
          if (addSuccessCount > 0) {
            messages.push(`成功绑定 ${addSuccessCount} 个知识库`)
          }
          if (deleteSuccessCount > 0) {
            messages.push(`成功删除 ${deleteSuccessCount} 个知识库`)
          }
          ElMessage.success(messages.join('，'))
          dialogVisible.value = false
          loadData()
        } else if (toAdd.length === 0 && toDelete.length === 0) {
          ElMessage.info('没有需要更新的内容')
          dialogVisible.value = false
        }

        if (failed.length > 0) {
          console.warn('员工知识库权限操作失败明细:', failed)
          ElMessage.warning(`失败 ${failed.length} 个：${failed.map(i => i.code).join(', ')}`)
        }
      } catch (error) {
        ElMessage.error('操作失败，请稍后重试')
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleDeleteTag = (row, index) => {
  const documentTypeCode = row.documentTypeCodes[index]
  const documentTypeName = row.documentTypeNames[index]
  
  ElMessageBox.confirm(
    `确定要删除员工 "${row.empName}" 的知识库 "${documentTypeName}" 吗？`,
    '提示',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      // 需要先查询该绑定记录的 id
      const listRes = await listEmpDocumentType({
        page: 1,
        pageSize: 100,
        empCode: row.empCode,
        documentTypeCode: documentTypeCode
      })
      
      if (listRes.code === 200 && listRes.data && listRes.data.list && listRes.data.list.length > 0) {
        const record = listRes.data.list.find(
          item => item.empCode === row.empCode && item.documentTypeCode === documentTypeCode
        )
        
        if (record) {
          const res = await deleteEmpDocumentType({ id: record.id })
          if (res.code === 200) {
            ElMessage.success('删除成功')
            loadData()
          } else {
            ElMessage.error(res.msg || res.message || '删除失败')
          }
        } else {
          ElMessage.error('未找到对应的绑定记录')
        }
      } else {
        ElMessage.error('未找到对应的绑定记录')
      }
    } catch (error) {
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleEdit = (row) => {
  // 编辑时，填充当前员工的知识库
  Object.assign(form, {
    empCode: row.empCode,
    documentTypeCodes: [...row.documentTypeCodes]
  })
  dialogVisible.value = true
  
  // 如果知识库列表为空，则加载
  if (documentTypeList.value.length === 0) {
    loadDocumentTypes()
  }
  
  // 如果员工不在列表中，添加到列表
  const empExists = employeeList.value.find(emp => emp.empCode === row.empCode)
  if (!empExists) {
    employeeList.value.push({
      empCode: row.empCode,
      empName: row.empName
    })
  }
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
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

// 搜索员工
const searchEmployees = async (query) => {
  if (!query || query.trim() === '') {
    employeeList.value = []
    return
  }
  
  employeeSearchLoading.value = true
  try {
    const keyword = query.trim()
    // 同时搜索编码和姓名，合并结果
    const [codeRes, nameRes] = await Promise.all([
      listEmp({
        page: 1,
        pageSize: 50,
        empCode: keyword
      }),
      listEmp({
        page: 1,
        pageSize: 50,
        empName: keyword
      })
    ])
    
    // 合并结果，去重
    const codeList = (codeRes.code === 200 && codeRes.data && codeRes.data.list) ? codeRes.data.list : []
    const nameList = (nameRes.code === 200 && nameRes.data && nameRes.data.list) ? nameRes.data.list : []
    
    // 使用 Map 去重（以 empCode 为 key）
    const empMap = new Map()
    codeList.forEach(emp => {
      empMap.set(emp.empCode, emp)
    })
    nameList.forEach(emp => {
      empMap.set(emp.empCode, emp)
    })
    
    employeeList.value = Array.from(empMap.values())
  } catch (error) {
    ElMessage.error('搜索员工失败')
    employeeList.value = []
  } finally {
    employeeSearchLoading.value = false
  }
}

// 处理员工选择变化
const handleEmployeeChange = (empCode) => {
  if (empCode) {
    const selectedEmp = employeeList.value.find(emp => emp.empCode === empCode)
    if (!selectedEmp) {
      // 如果当前列表中没有，尝试搜索
      searchEmployees(empCode)
    }
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
</style>

