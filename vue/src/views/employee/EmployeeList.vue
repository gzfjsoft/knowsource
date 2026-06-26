<template>
  <div class="employee-list">
    <div class="employee-list-container">
      <!-- 左侧部门树 -->
      <div class="left-panel">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>部门树</span>
              <el-button type="primary" @click="loadDeptTree" :loading="deptTreeLoading" size="small">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
          </template>
          
          <div v-loading="deptTreeLoading" class="tree-container">
            <el-tree
              ref="deptTreeRef"
              v-if="deptTreeData.length > 0"
              :data="deptTreeData"
              :props="treeProps"
              :expand-on-click-node="false"
              node-key="deptCode"
              class="dept-tree-view"
              :highlight-current="true"
              :default-expanded-keys="defaultExpandedKeys"
              @node-click="handleDeptNodeClick"
            >
              <template #default="{ data }">
                <span class="tree-node">
                  <span class="dept-name-cell">
                    <el-icon class="dept-name-icon">
                      <OfficeBuilding v-if="!data.parentCode" />
                      <Folder v-else />
                    </el-icon>
                    <span class="node-label">{{ data.deptName }}</span>
                  </span>
                  <span class="node-info">
                    <el-tag size="small" type="info">{{ data.deptCode }}</el-tag>
                  </span>
                </span>
              </template>
            </el-tree>
            <el-empty v-else description="暂无数据" />
          </div>
        </el-card>
      </div>

      <!-- 右侧员工列表 -->
      <div class="right-panel">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>员工管理</span>
              <div class="header-actions">
                <el-button type="primary" @click="openCreateEmpDialog" size="small">
                  新增员工
                </el-button>
                <el-button type="success" @click="handleImportHrUserDept" :loading="importLoading" size="small">
                  <el-icon><Upload /></el-icon>
                  导入HR数据
                </el-button>
              </div>
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
        <el-form-item label="部门编码">
          <el-input v-model="searchForm.deptCode" placeholder="请输入部门编码" clearable />
        </el-form-item>
        <!-- <el-form-item label="公司代码">
          <el-input v-model="searchForm.gsdm" placeholder="请输入公司代码" clearable />
        </el-form-item> -->
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="请选择状态" clearable style="width: 160px">
            <el-option label="启用" value="0" />
            <el-option label="禁用" value="1" />
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
        <el-table-column prop="empCode" label="员工编码" width="120" />
        <el-table-column prop="empName" label="员工姓名" width="120" />
        <el-table-column prop="deptCode" label="部门编码" width="120" />
        <el-table-column prop="deptName" label="部门名称" width="150" />
        <!-- 隐藏公司代码列 -->
        <!-- <el-table-column prop="gsdm" label="公司代码" width="100" /> -->
        <!-- 隐藏部门代码列 -->
        <!-- <el-table-column prop="bmdm" label="部门代码" width="100" /> -->
        <el-table-column prop="position" label="职位" width="120" />
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'">
              {{ row.status === 0 ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="roles" label="角色" width="200">
          <template #default="{ row }">
            <div v-if="row.roles && row.roles.length > 0" style="display: flex; gap: 4px; flex-wrap: wrap;">
              <el-tag
                v-for="role in row.roles"
                :key="role"
                :type="getRoleTagType(role)"
                size="small"
              >
                {{ getRoleLabel(role) }}
              </el-tag>
            </div>
            <el-tag v-else type="info" size="small">无角色</el-tag>
          </template>
        </el-table-column>
        <!-- 隐藏年龄列 -->
        <!-- <el-table-column prop="age" label="年龄" width="80" /> -->
        <!-- 隐藏性别列 -->
        <!-- <el-table-column prop="sex" label="性别" width="80" /> -->
        <!-- 隐藏学历列 -->
        <!-- <el-table-column prop="education" label="学历" width="100" /> -->
        <!-- 隐藏政治面貌列 -->
        <!-- <el-table-column prop="politics" label="政治面貌" width="100" /> -->
        <!-- 隐藏职位级别列 -->
        <!-- <el-table-column prop="positionLevel" label="职位级别" width="100" /> -->
        <!-- 隐藏省份列 -->
        <!-- <el-table-column prop="province" label="省份" width="100" /> -->
        <el-table-column label="操作" width="168" fixed="right" align="center">
          <template #default="{ row }">
            <span class="table-op-actions">
              <el-tooltip content="编辑" placement="top">
                <el-button
                  type="primary"
                  link
                  size="small"
                  :icon="EditPen"
                  aria-label="编辑"
                  @click="openEditEmpDialog(row)"
                />
              </el-tooltip>
              <el-tooltip content="删除" placement="top">
                <el-button
                  type="danger"
                  link
                  size="small"
                  :icon="Delete"
                  aria-label="删除"
                  @click="handleDeleteEmp(row)"
                />
              </el-tooltip>
              <el-tooltip content="重置密码" placement="top">
                <el-button
                  type="primary"
                  link
                  size="small"
                  :icon="Key"
                  aria-label="重置密码"
                  @click="handleResetPassword(row)"
                />
              </el-tooltip>
              <el-tooltip content="绑定角色" placement="top">
                <el-button
                  type="warning"
                  link
                  size="small"
                  :icon="Avatar"
                  aria-label="绑定角色"
                  @click="handleBindRole(row)"
                />
              </el-tooltip>
            </span>
          </template>
        </el-table-column>
      </el-table>

      <!-- 新增/编辑员工弹窗 -->
      <el-dialog
        v-model="empDialogVisible"
        :title="empDialogMode === 'create' ? '新增员工' : '编辑员工'"
        width="560px"
        :close-on-click-modal="false"
      >
        <el-form :model="empForm" label-width="100px">
          <el-form-item label="员工编码" required>
            <el-input v-model="empForm.empCode" :disabled="empDialogMode === 'edit'" placeholder="请输入员工编码" />
          </el-form-item>
          <el-form-item label="员工姓名" required>
            <el-input v-model="empForm.empName" placeholder="请输入员工姓名" />
          </el-form-item>
          <el-form-item label="部门编码" required>
            <el-select
              v-model="empForm.deptCode"
              placeholder="请选择部门（可输入筛选名称/编码）"
              filterable
              :filter-method="filterDeptOptions"
              style="width: 100%"
              clearable
            >
              <el-option
                v-for="d in filteredDeptOptions"
                :key="d.deptCode"
                :label="`${d.deptName}（${d.deptCode}）`"
                :value="d.deptCode"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="职位">
            <el-input v-model="empForm.position" placeholder="请输入职位" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="empForm.status" placeholder="请选择状态" style="width: 160px">
              <el-option label="启用" :value="0" />
              <el-option label="禁用" :value="1" />
            </el-select>
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="empDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="empSubmitLoading" @click="handleSubmitEmp">确定</el-button>
        </template>
      </el-dialog>

      <!-- 重置密码弹窗 -->
      <el-dialog
        v-model="resetPasswordDialogVisible"
        title="重置密码"
        width="500px"
        :close-on-click-modal="false"
      >
        <el-form :model="resetPasswordForm" label-width="100px">
          <el-form-item label="员工编码">
            <el-input v-model="resetPasswordForm.empCode" disabled />
          </el-form-item>
          <el-form-item label="员工姓名">
            <el-input v-model="resetPasswordForm.empName" disabled />
          </el-form-item>
          <el-form-item label="新密码" required>
            <el-input
              v-model="resetPasswordForm.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="请输入新密码"
              clearable
            >
              <template #append>
                <el-button @click="generatePassword">生成密码</el-button>
              </template>
              <template #suffix>
                <el-icon
                  class="password-icon"
                  @click="showPassword = !showPassword"
                  style="cursor: pointer;"
                >
                  <View v-if="!showPassword" />
                  <Hide v-else />
                </el-icon>
              </template>
            </el-input>
          </el-form-item>
          <el-form-item>
            <el-alert
              title="密码要求：至少8位，包含大小写字母、数字和特殊符号"
              type="info"
              :closable="false"
              show-icon
            />
          </el-form-item>
        </el-form>
        <template #footer>
          <el-button @click="resetPasswordDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            @click="handleConfirmResetPassword"
            :loading="resetPasswordLoading"
          >
            确定
          </el-button>
        </template>
      </el-dialog>

      <!-- 绑定角色弹窗 -->
      <el-dialog
        v-model="bindRoleDialogVisible"
        title="绑定角色"
        width="600px"
        :close-on-click-modal="false"
      >
        <el-form :model="bindRoleForm" label-width="100px">
          <el-form-item label="员工编码">
            <el-input v-model="bindRoleForm.empCode" disabled />
          </el-form-item>
          <el-form-item label="员工姓名">
            <el-input v-model="bindRoleForm.empName" disabled />
          </el-form-item>
          <el-form-item label="选择角色">
            <el-select
              v-model="bindRoleForm.selectedRoles"
              placeholder="请选择角色（可多选）"
              multiple
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="role in availableRoles"
                :key="role.role"
                :label="role.name"
                :value="role.role"
              />
            </el-select>
          </el-form-item>
        </el-form>

        <template #footer>
          <el-button @click="bindRoleDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            @click="handleConfirmBindRoles"
            :loading="bindRoleLoading"
          >
            确定
          </el-button>
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
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  Refresh,
  Upload,
  View,
  Hide,
  OfficeBuilding,
  Folder,
  EditPen,
  Delete,
  Key,
  Avatar
} from '@element-plus/icons-vue'
import {
  listEmp,
  createEmp,
  updateEmp,
  deleteEmp,
  adminResetPassword,
  getDeptTree,
  importHrUserDept,
  listFrUserRole,
  createFrUserRole,
  deleteFrUserRole,
  listFrRole
} from '@/api/knowsource'

const loading = ref(false)
const tableData = ref([])
const deptTreeLoading = ref(false)
const deptTreeData = ref([])
const deptTreeRef = ref(null)
const selectedDeptCode = ref('')
const importLoading = ref(false)
const resetPasswordDialogVisible = ref(false)
const resetPasswordLoading = ref(false)
const showPassword = ref(false)
const resetPasswordForm = reactive({
  empCode: '',
  empName: '',
  password: ''
})
const bindRoleDialogVisible = ref(false)
const bindRoleLoading = ref(false)
const bindRoleForm = reactive({
  empCode: '',
  empName: '',
  selectedRoles: [] // 选中的角色数组
})
const availableRoles = ref([])
const userRoleList = ref([]) // 存储用户角色绑定列表（包含 id）

const empDialogVisible = ref(false)
const empDialogMode = ref('create') // create | edit
const empSubmitLoading = ref(false)
const empForm = reactive({
  empCode: '',
  empName: '',
  deptCode: '',
  position: '',
  status: 0,
  branch: ''
})

const deptOptions = ref([])
const filteredDeptOptions = ref([])

const searchForm = reactive({
  empCode: '',
  empName: '',
  deptCode: '',
  gsdm: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const treeProps = {
  children: 'children',
  label: 'deptName'
}

const flattenDeptTree = (nodes, acc = []) => {
  if (!Array.isArray(nodes)) return acc
  for (const n of nodes) {
    if (!n) continue
    acc.push({ deptCode: n.deptCode, deptName: n.deptName })
    if (Array.isArray(n.children) && n.children.length > 0) {
      flattenDeptTree(n.children, acc)
    }
  }
  return acc
}

const rebuildDeptOptions = () => {
  const list = flattenDeptTree(deptTreeData.value, [])
  // 去重并排序（按名称，再按编码）
  const map = new Map()
  for (const d of list) {
    if (!d.deptCode) continue
    if (!map.has(d.deptCode)) map.set(d.deptCode, d)
  }
  const uniq = Array.from(map.values()).sort((a, b) => {
    const an = a.deptName || ''
    const bn = b.deptName || ''
    if (an !== bn) return an.localeCompare(bn, 'zh-Hans-CN')
    return (a.deptCode || '').localeCompare(b.deptCode || '')
  })
  deptOptions.value = uniq
  filteredDeptOptions.value = uniq
}

const filterDeptOptions = (keyword) => {
  const k = (keyword || '').trim().toLowerCase()
  if (!k) {
    filteredDeptOptions.value = deptOptions.value
    return
  }
  filteredDeptOptions.value = deptOptions.value.filter(d => {
    const code = (d.deptCode || '').toLowerCase()
    const name = (d.deptName || '').toLowerCase()
    return code.includes(k) || name.includes(k)
  })
}

// 计算默认展开的节点（第一层节点）
const defaultExpandedKeys = computed(() => {
  return deptTreeData.value.map(node => node.deptCode)
})

// 加载部门树
const loadDeptTree = async () => {
  deptTreeLoading.value = true
  try {
    const res = await getDeptTree({})
    if (res.code === 200 && res.data && res.data.tree) {
      deptTreeData.value = res.data.tree
      rebuildDeptOptions()
    } else {
      ElMessage.warning('暂无数据')
      deptTreeData.value = []
      rebuildDeptOptions()
    }
  } catch (error) {
    ElMessage.error('加载部门树失败')
    deptTreeData.value = []
    rebuildDeptOptions()
  } finally {
    deptTreeLoading.value = false
  }
}

// 处理部门树节点点击
const handleDeptNodeClick = (data) => {
  selectedDeptCode.value = data.deptCode
  searchForm.deptCode = data.deptCode
  pagination.page = 1
  loadData()
}

const loadData = async () => {
  loading.value = true
  try {
    // 构建请求参数，status 为空字符串时忽略
    const params = {
      page: pagination.page,
      pageSize: pagination.pageSize,
      empCode: searchForm.empCode || '',
      empName: searchForm.empName || '',
      deptCode: searchForm.deptCode || '',
      status: searchForm.status || ''
    }
    const res = await listEmp(params)
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
    empName: '',
    deptCode: '',
    gsdm: '',
    status: ''
  })
  selectedDeptCode.value = ''
  // 清除树的选择状态
  if (deptTreeRef.value) {
    deptTreeRef.value.setCurrentKey(null)
  }
  handleSearch()
}

const handleSizeChange = () => {
  loadData()
}

const handlePageChange = () => {
  loadData()
}

const resetEmpForm = () => {
  Object.assign(empForm, {
    empCode: '',
    empName: '',
    deptCode: '',
    position: '',
    status: 0,
    branch: ''
  })
}

const openCreateEmpDialog = () => {
  empDialogMode.value = 'create'
  resetEmpForm()
  empForm.deptCode = selectedDeptCode.value || searchForm.deptCode || ''
  empDialogVisible.value = true
}

const openEditEmpDialog = (row) => {
  empDialogMode.value = 'edit'
  Object.assign(empForm, {
    empCode: row.empCode || '',
    empName: row.empName || '',
    deptCode: row.deptCode || '',
    position: row.position || '',
    status: typeof row.status === 'number' ? row.status : 0,
    branch: row.branch || ''
  })
  empDialogVisible.value = true
}

const handleSubmitEmp = async () => {
  if (!empForm.empCode || !empForm.empName || !empForm.deptCode) {
    ElMessage.warning('请填写员工编码、员工姓名、部门编码')
    return
  }
  empSubmitLoading.value = true
  try {
    const payload = {
      empCode: empForm.empCode,
      empName: empForm.empName,
      deptCode: empForm.deptCode,
      position: empForm.position,
      status: empForm.status,
      branch: empForm.branch
    }
    const res = empDialogMode.value === 'create' ? await createEmp(payload) : await updateEmp(payload)
    if (res.code === 200) {
      ElMessage.success('操作成功')
      empDialogVisible.value = false
      loadData()
      loadDeptTree()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (e) {
    ElMessage.error('操作失败')
  } finally {
    empSubmitLoading.value = false
  }
}

const handleDeleteEmp = async (row) => {
  try {
    await ElMessageBox.confirm(`确认删除员工 ${row.empName}（${row.empCode}）？`, '提示', { type: 'warning' })
  } catch {
    return
  }
  empSubmitLoading.value = true
  try {
    const res = await deleteEmp({ empCode: row.empCode })
    if (res.code === 200) {
      ElMessage.success('删除成功')
      loadData()
      loadDeptTree()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (e) {
    ElMessage.error('删除失败')
  } finally {
    empSubmitLoading.value = false
  }
}

// 生成8位密码：前7个字符为大小写字母和数字，最后一个字符为特殊符号
const generatePassword = () => {
  const uppercase = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
  const lowercase = 'abcdefghijklmnopqrstuvwxyz'
  const numbers = '0123456789'
  const symbols = '!@#$%^&*()_+-=[]{}|;:,.<>?'
  const lettersAndNumbers = uppercase + lowercase + numbers

  // 前7个字符：确保至少包含一个大写字母、一个小写字母和一个数字
  let password = ''
  password += uppercase[Math.floor(Math.random() * uppercase.length)]
  password += lowercase[Math.floor(Math.random() * lowercase.length)]
  password += numbers[Math.floor(Math.random() * numbers.length)]

  // 填充剩余4位（前7个字符中的剩余4位）
  for (let i = password.length; i < 7; i++) {
    password += lettersAndNumbers[Math.floor(Math.random() * lettersAndNumbers.length)]
  }

  // 打乱前7个字符的顺序
  const first7 = password.split('').sort(() => Math.random() - 0.5).join('')
  
  // 最后一个字符：特殊符号
  const lastChar = symbols[Math.floor(Math.random() * symbols.length)]
  
  resetPasswordForm.password = first7 + lastChar
}

// 打开重置密码弹窗
const handleResetPassword = (row) => {
  resetPasswordForm.empCode = row.empCode || ''
  resetPasswordForm.empName = row.empName || ''
  resetPasswordForm.password = ''
  showPassword.value = false
  resetPasswordDialogVisible.value = true
}

// 确认重置密码
const handleConfirmResetPassword = async () => {
  if (!resetPasswordForm.password) {
    ElMessage.warning('请输入新密码')
    return
  }

  // 验证密码格式：至少8位，包含大小写字母、数字和特殊符号
  const password = resetPasswordForm.password
  if (password.length < 8) {
    ElMessage.warning('密码长度至少为8位')
    return
  }
  
  const hasUpperCase = /[A-Z]/.test(password)
  const hasLowerCase = /[a-z]/.test(password)
  const hasNumber = /[0-9]/.test(password)
  const hasSpecialChar = /[!@#$%^&*()_+\-=\[\]{}|;:,.<>?]/.test(password)
  
  if (!hasUpperCase || !hasLowerCase || !hasNumber || !hasSpecialChar) {
    ElMessage.warning('密码必须包含大小写字母、数字和特殊符号')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要重置员工 ${resetPasswordForm.empName} (${resetPasswordForm.empCode}) 的密码吗？`,
      '确认重置',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    resetPasswordLoading.value = true
    const res = await adminResetPassword({
      empCode: resetPasswordForm.empCode,
      password: resetPasswordForm.password
    })

    if (res.code === 200) {
      ElMessage.success('密码重置成功')
      resetPasswordDialogVisible.value = false
      resetPasswordForm.password = ''
    } else {
      ElMessage.error(res.message || '密码重置失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '密码重置失败')
    }
  } finally {
    resetPasswordLoading.value = false
  }
}

// 加载可用角色列表
const loadAvailableRoles = async () => {
  try {
    const res = await listFrRole({
      page: 1,
      pageSize: 1000
    })
    if (res.code === 200) {
      availableRoles.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取角色列表失败：', error)
  }
}

// 加载用户当前的角色绑定
const loadUserRoles = async (empCode) => {
  bindRoleLoading.value = true
  try {
    const res = await listFrUserRole({
      page: 1,
      pageSize: 1000,
      empCode: empCode
    })
    if (res.code === 200) {
      userRoleList.value = res.data.list || []
    } else {
      ElMessage.error(res.message || '获取用户角色失败')
    }
  } catch (error) {
    ElMessage.error('获取用户角色失败：' + error.message)
  } finally {
    bindRoleLoading.value = false
  }
}

// 打开绑定角色弹窗
const handleBindRole = async (row) => {
  bindRoleForm.empCode = row.empCode || ''
  bindRoleForm.empName = row.empName || ''
  bindRoleForm.selectedRoles = []
  bindRoleDialogVisible.value = true
  
  // 加载用户当前的角色绑定（包含 id）
  await loadUserRoles(row.empCode)
  
  // 初始化选中的角色为当前已绑定的角色
  bindRoleForm.selectedRoles = userRoleList.value.map(item => item.role)
  
  // 如果还没有加载可用角色列表，则加载
  if (availableRoles.value.length === 0) {
    await loadAvailableRoles()
  }
}

// 确认绑定角色
const handleConfirmBindRoles = async () => {
  try {
    bindRoleLoading.value = true
    
    // 获取当前已绑定的角色
    const currentRoles = userRoleList.value.map(item => item.role)
    const selectedRoles = bindRoleForm.selectedRoles || []
    
    // 找出需要添加的角色（在 selectedRoles 中但不在 currentRoles 中）
    const rolesToAdd = selectedRoles.filter(role => !currentRoles.includes(role))
    
    // 找出需要删除的角色（在 currentRoles 中但不在 selectedRoles 中）
    const rolesToDelete = currentRoles.filter(role => !selectedRoles.includes(role))
    
    // 如果没有变化，直接返回
    if (rolesToAdd.length === 0 && rolesToDelete.length === 0) {
      ElMessage.info('角色未发生变化')
      bindRoleDialogVisible.value = false
      return
    }
    
    // 删除需要移除的角色
    for (const role of rolesToDelete) {
      const userRole = userRoleList.value.find(item => item.role === role)
      if (userRole && userRole.id) {
        const res = await deleteFrUserRole({ id: userRole.id })
        if (res.code !== 200) {
          ElMessage.error(`删除角色 ${role} 失败：${res.message}`)
          return
        }
      }
    }
    
    // 添加需要新增的角色
    for (const role of rolesToAdd) {
      const res = await createFrUserRole({
        empCode: bindRoleForm.empCode,
        role: role
      })
      if (res.code !== 200) {
        ElMessage.error(`添加角色 ${role} 失败：${res.message}`)
        return
      }
    }
    
    ElMessage.success('角色绑定成功')
    bindRoleDialogVisible.value = false
    // 刷新列表
    loadData()
  } catch (error) {
    ElMessage.error('角色绑定失败：' + error.message)
  } finally {
    bindRoleLoading.value = false
  }
}

// 获取角色标签文本
const getRoleLabel = (role) => {
  switch (role) {
    case 'superadmin':
      return '超级管理员'
    case 'admin':
      return '管理员'
    case 'user':
      return '普通用户'
    case 'demo':
      return '演示用户'
    default:
      return role || '普通用户'
  }
}

// 获取角色标签类型（用于 el-tag 的 type 属性）
const getRoleTagType = (role) => {
  switch (role) {
    case 'superadmin':
      return 'danger'
    case 'admin':
      return 'warning'
    case 'demo':
      return 'success'
    case 'user':
      return 'info'
    default:
      return 'info'
  }
}

// 导入 HR 用户和部门数据
const handleImportHrUserDept = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要导入 HR 用户和部门数据吗？此操作将从 MSSQL 数据库同步数据到 MySQL 数据库。',
      '确认导入',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    importLoading.value = true
    const res = await importHrUserDept()

    if (res.code === 200) {
      ElMessage.success(res.message || '导入成功')
      // 刷新列表和部门树
      loadDeptTree()
      loadData()
    } else {
      ElMessage.error(res.message || res.info || '导入失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '导入失败')
    }
  } finally {
    importLoading.value = false
  }
}

onMounted(() => {
  loadDeptTree()
  loadData()
  loadAvailableRoles()
})
</script>

<style scoped>
.employee-list-container {
  display: flex;
  gap: 20px;
  height: calc(100vh - 100px);
}

.left-panel {
  width: 350px;
  flex-shrink: 0;
}

.right-panel {
  flex: 1;
  min-width: 0;
}

.card-header {
  font-size: 18px;
  font-weight: 500;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.tree-container {
  min-height: 400px;
  max-height: calc(100vh - 200px);
  overflow-y: auto;
}

.dept-tree-view {
  background-color: #fff;
  padding: 10px;
  border-radius: 4px;
}

.tree-node {
  display: flex;
  align-items: center;
  flex: 1;
  font-size: 14px;
  padding-right: 8px;
}

.dept-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.dept-name-icon {
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
}

.node-label {
  font-weight: 500;
  color: #303133;
  margin-right: 12px;
}

.node-info {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
}

:deep(.el-tree-node__content) {
  height: 40px;
  line-height: 40px;
}

:deep(.el-tree-node__label) {
  width: 100%;
}

:deep(.el-tree-node.is-current > .el-tree-node__content) {
  background-color: #ecf5ff;
  color: #409eff;
}

.search-form {
  margin-bottom: 20px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.table-op-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  flex-wrap: nowrap;
}
</style>

