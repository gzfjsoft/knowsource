<template>
  <div class="role-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>角色管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            新增角色
          </el-button>
        </div>
      </template>
      
      <!-- 搜索表单 -->
      <el-form :inline="true" :model="searchForm" class="search-form">
        <el-form-item label="角色编码">
          <el-input v-model="searchForm.role" placeholder="请输入角色编码" clearable />
        </el-form-item>
        <el-form-item label="角色名称">
          <el-input v-model="searchForm.name" placeholder="请输入角色名称" clearable />
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
        <el-table-column prop="role" label="角色编码" width="150" />
        <el-table-column prop="name" label="角色名称" width="200" />
        <el-table-column label="操作" width="280" fixed="right">
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
              type="success"
              link
              size="small"
              @click="handleBindPermission(row)"
            >
              权限绑定
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
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="角色编码" prop="role">
          <el-input
            v-model="formData.role"
            placeholder="请输入角色编码"
            :disabled="isEdit"
          />
        </el-form-item>
        <el-form-item label="角色名称" prop="name">
          <el-input
            v-model="formData.name"
            placeholder="请输入角色名称"
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

    <!-- 权限绑定对话框 -->
    <el-dialog
      v-model="permissionDialogVisible"
      title="角色权限绑定"
      width="900px"
      @close="handlePermissionDialogClose"
    >
      <div class="permission-bind-container">
        <div class="permission-header">
          <span class="role-info">角色：{{ currentRole.role }} - {{ currentRole.name }}</span>
        </div>

        <div class="permission-columns">
          <!-- 左列：已选中的权限 -->
          <div class="permission-column">
            <div class="column-header">
              <span class="column-title">已选权限 ({{ selectedPermissions.length }})</span>
            </div>
            <div class="column-content" v-loading="permissionLoading">
              <div
                v-for="perm in selectedPermissions"
                :key="perm.permission"
                class="permission-item"
              >
                <div class="permission-info">
                  <div class="permission-code">{{ perm.permission }}</div>
                  <div class="permission-desc">{{ perm.description }}</div>
                </div>
                <el-button
                  type="danger"
                  link
                  size="small"
                  @click="handleRemovePermission(perm)"
                  :loading="permissionLoading"
                >
                  <el-icon><Delete /></el-icon>
                  删除
                </el-button>
              </div>
              <el-empty v-if="selectedPermissions.length === 0" description="暂无已选权限" />
            </div>
          </div>

          <!-- 右列：未选中的权限 -->
          <div class="permission-column">
            <div class="column-header">
              <span class="column-title">未选权限 ({{ unselectedPermissions.length }})</span>
            </div>
            <div class="column-content" v-loading="permissionLoading">
              <div
                v-for="perm in unselectedPermissions"
                :key="perm.permission"
                class="permission-item"
              >
                <div class="permission-info">
                  <div class="permission-code">{{ perm.permission }}</div>
                  <div class="permission-desc">{{ perm.description }}</div>
                </div>
                <el-button
                  type="primary"
                  link
                  size="small"
                  @click="handleAddPermission(perm)"
                  :loading="permissionLoading"
                >
                  <el-icon><Plus /></el-icon>
                  添加
                </el-button>
              </div>
              <el-empty v-if="unselectedPermissions.length === 0" description="暂无未选权限" />
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Plus, Delete } from '@element-plus/icons-vue'
import {
  listFrRole,
  createFrRole,
  updateFrRole,
  deleteFrRole,
  listFrRolePermission,
  createFrRolePermission,
  deleteFrRolePermission,
  listFrPermission
} from '@/api/knowsource'

const loading = ref(false)
const submitLoading = ref(false)
const tableData = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)

// 权限绑定相关
const permissionDialogVisible = ref(false)
const permissionLoading = ref(false)
const permissionList = ref([]) // 已绑定的权限列表（包含 id）
const currentRole = ref({ role: '', name: '' })
const allPermissions = ref([]) // 所有权限列表

// 计算属性：已选中的权限
const selectedPermissions = computed(() => {
  const selectedCodes = permissionList.value.map(p => p.permission)
  return allPermissions.value.filter(p => selectedCodes.includes(p.permission))
})

// 计算属性：未选中的权限
const unselectedPermissions = computed(() => {
  const selectedCodes = permissionList.value.map(p => p.permission)
  return allPermissions.value.filter(p => !selectedCodes.includes(p.permission))
})

const searchForm = reactive({
  role: '',
  name: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const formData = reactive({
  role: '',
  name: ''
})

const formRules = {
  role: [
    { required: true, message: '请输入角色编码', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入角色名称', trigger: 'blur' }
  ]
}

const dialogTitle = computed(() => {
  return isEdit.value ? '编辑角色' : '新增角色'
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await listFrRole({
      page: pagination.page,
      pageSize: pagination.pageSize,
      ...searchForm
    })
    if (res.code === 200) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
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
  pagination.page = 1
  loadData()
}

const handleReset = () => {
  searchForm.role = ''
  searchForm.name = ''
  pagination.page = 1
  loadData()
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const handleAdd = () => {
  isEdit.value = false
  formData.role = ''
  formData.name = ''
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  formData.role = row.role
  formData.name = row.name
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该角色吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const res = await deleteFrRole({ role: row.role })
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
          res = await updateFrRole({
            role: formData.role,
            name: formData.name
          })
        } else {
          res = await createFrRole({
            role: formData.role,
            name: formData.name
          })
        }
        
        if (res.code === 200) {
          ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
          dialogVisible.value = false
          loadData()
        } else {
          ElMessage.error(res.message || '操作失败')
        }
      } catch (error) {
        ElMessage.error('操作失败：' + error.message)
      } finally {
        submitLoading.value = false
      }
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
}

// 权限绑定相关方法
const handleBindPermission = async (row) => {
  currentRole.value = { role: row.role, name: row.name }
  permissionDialogVisible.value = true
  await loadPermissionList()
  await loadAllPermissions()
}

const loadPermissionList = async () => {
  permissionLoading.value = true
  try {
    const res = await listFrRolePermission({
      role: currentRole.value.role,
      page: 1,
      pageSize: 1000
    })
    if (res.code === 200) {
      permissionList.value = res.data.list || []
    } else {
      ElMessage.error(res.message || '获取权限列表失败')
    }
  } catch (error) {
    ElMessage.error('获取权限列表失败：' + error.message)
  } finally {
    permissionLoading.value = false
  }
}

const loadAllPermissions = async () => {
  try {
    const res = await listFrPermission({
      page: 1,
      pageSize: 1000
    })
    if (res.code === 200) {
      allPermissions.value = res.data.list || []
    }
  } catch (error) {
    ElMessage.error('获取所有权限失败：' + error.message)
  }
}

const handleAddPermission = async (perm) => {
  permissionLoading.value = true
  try {
    const res = await createFrRolePermission({
      role: currentRole.value.role,
      permission: perm.permission
    })
    if (res.code === 200) {
      ElMessage.success('添加权限成功')
      await loadPermissionList()
    } else {
      ElMessage.error(res.message || '添加权限失败')
    }
  } catch (error) {
    ElMessage.error('添加权限失败：' + error.message)
  } finally {
    permissionLoading.value = false
  }
}

const handleRemovePermission = async (perm) => {
  try {
    await ElMessageBox.confirm('确定要移除该权限吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 找到对应的权限绑定记录（包含 id）
    const rolePermission = permissionList.value.find(p => p.permission === perm.permission)
    if (!rolePermission) {
      ElMessage.error('未找到对应的权限绑定记录')
      return
    }
    
    permissionLoading.value = true
    const res = await deleteFrRolePermission({ id: rolePermission.id })
    if (res.code === 200) {
      ElMessage.success('移除权限成功')
      await loadPermissionList()
    } else {
      ElMessage.error(res.message || '移除权限失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('移除权限失败：' + error.message)
    }
  } finally {
    permissionLoading.value = false
  }
}

const handlePermissionDialogClose = () => {
  permissionList.value = []
  currentRole.value = { role: '', name: '' }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.role-management {
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .search-form {
    margin-bottom: 20px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }

  .permission-bind-container {
    .permission-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 10px 0;
      margin-bottom: 20px;

      .role-info {
        font-size: 16px;
        font-weight: 500;
        color: #303133;
      }
    }

    .permission-columns {
      display: flex;
      gap: 20px;
      height: 500px;

      .permission-column {
        flex: 1;
        display: flex;
        flex-direction: column;
        border: 1px solid #dcdfe6;
        border-radius: 4px;
        overflow: hidden;

        .column-header {
          padding: 12px 16px;
          background-color: #f5f7fa;
          border-bottom: 1px solid #dcdfe6;

          .column-title {
            font-size: 14px;
            font-weight: 500;
            color: #303133;
          }
        }

        .column-content {
          flex: 1;
          overflow-y: auto;
          padding: 8px;

          .permission-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px;
            margin-bottom: 8px;
            border: 1px solid #e4e7ed;
            border-radius: 4px;
            background-color: #fff;
            transition: all 0.3s;

            &:hover {
              border-color: #409eff;
              box-shadow: 0 2px 4px rgba(64, 158, 255, 0.1);
            }

            .permission-info {
              flex: 1;
              min-width: 0;

              .permission-code {
                font-size: 14px;
                font-weight: 500;
                color: #303133;
                margin-bottom: 4px;
              }

              .permission-desc {
                font-size: 12px;
                color: #909399;
                overflow: hidden;
                text-overflow: ellipsis;
                white-space: nowrap;
              }
            }
          }
        }
      }
    }
  }
}
</style>
