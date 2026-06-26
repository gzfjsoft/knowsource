<template>
  <div class="dept-tree">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>部门管理</span>
          <div class="header-actions">
            <el-segmented
              v-model="viewMode"
              :options="viewModeOptions"
              size="small"
            />
            <el-button type="primary" @click="openCreateRoot" size="small">
              <el-icon><Plus /></el-icon>
              新增根部门
            </el-button>
            <el-button type="primary" @click="loadTree" :loading="loading" size="small">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>
      
      <div v-loading="loading" class="tree-container">
        <!-- TreeTable（带表头，更像后台管理） -->
        <el-table
          v-if="viewMode === 'table' && treeData.length > 0"
          :data="treeData"
          row-key="deptCode"
          border
          stripe
          class="dept-tree-table"
          :tree-props="{ children: 'children' }"
          :default-expand-all="true"
        >
          <el-table-column prop="deptName" label="部门名称" min-width="220">
            <template #default="{ row }">
              <span class="dept-name-cell">
                <el-icon class="dept-name-icon">
                  <OfficeBuilding v-if="!row.parentCode" />
                  <Folder v-else />
                </el-icon>
                <span>{{ row.deptName }}</span>
              </span>
            </template>
          </el-table-column>
          <el-table-column prop="deptCode" label="部门编码" width="140" />
          <el-table-column prop="kind" label="类型" width="120" />
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openCreateChild(row)">新增下级</el-button>
              <el-button link type="warning" size="small" @click="openEdit(row)">编辑</el-button>
              <el-button link type="danger" size="small" @click="handleDeleteTree(row)">删除子树</el-button>
            </template>
          </el-table-column>
        </el-table>

        <!-- 拖拽模式：保留原 el-tree draggable 能力 -->
        <el-tree
          v-else-if="viewMode === 'drag' && treeData.length > 0"
          ref="treeRef"
          :data="treeData"
          :props="treeProps"
          :expand-on-click-node="false"
          node-key="deptCode"
          class="dept-tree-view"
          draggable
          :allow-drop="allowDrop"
          @node-drop="handleNodeDrop"
          :default-expand-all="true"
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
                <el-tag v-if="data.kind" size="small" type="info">
                  类型: {{ data.kind }}
                </el-tag>
              </span>
              <span class="node-actions">
                <el-button link type="primary" size="small" @click.stop="openCreateChild(data)">新增下级</el-button>
                <el-button link type="warning" size="small" @click.stop="openEdit(data)">编辑</el-button>
                <el-button link type="danger" size="small" @click.stop="handleDeleteTree(data)">删除子树</el-button>
              </span>
            </span>
          </template>
        </el-tree>
        <el-empty v-else description="暂无数据" />
      </div>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新增部门' : '编辑部门'"
      width="640px"
      :close-on-click-modal="false"
    >
      <el-form :model="form" label-width="110px">
        <el-form-item label="上级部门">
          <el-input v-model="form.parentCode" placeholder="根部门留空" disabled />
        </el-form-item>
        <el-form-item label="部门编码" required>
          <el-input v-model="form.deptCode" :disabled="dialogMode === 'edit'" placeholder="请输入部门编码" />
        </el-form-item>
        <el-form-item label="部门名称" required>
          <el-input v-model="form.deptName" placeholder="请输入部门名称" />
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
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, OfficeBuilding, Folder } from '@element-plus/icons-vue'
import { getDeptTree, createDept, updateDept, deleteDept, moveDept } from '@/api/knowsource'

const loading = ref(false)
const submitLoading = ref(false)
const treeData = ref([])
const treeRef = ref(null)

const viewMode = ref('table') // table | drag
const viewModeOptions = [
  { label: '表格', value: 'table' },
  { label: '拖拽', value: 'drag' }
]

const dialogVisible = ref(false)
const dialogMode = ref('create') // create | edit
const form = reactive({
  deptCode: '',
  deptName: '',
  parentCode: '',
  grade: 0,
  endMark: '',
  kind: '',
  b0110: ''
})

const treeProps = {
  children: 'children',
  label: 'deptName'
}

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getDeptTree({})
    if (res.code === 200 && res.data && res.data.tree) {
      treeData.value = res.data.tree
    } else {
      ElMessage.warning('暂无数据')
      treeData.value = []
    }
  } catch (error) {
    ElMessage.error('加载部门树失败')
    treeData.value = []
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  Object.assign(form, {
    deptCode: '',
    deptName: '',
    parentCode: '',
    grade: 0,
    endMark: '',
    kind: '',
    b0110: ''
  })
}

const openCreateRoot = () => {
  dialogMode.value = 'create'
  resetForm()
  form.parentCode = ''
  dialogVisible.value = true
}

const openCreateChild = (node) => {
  dialogMode.value = 'create'
  resetForm()
  form.parentCode = node.deptCode || ''
  form.grade = typeof node.grade === 'number' ? (node.grade + 1) : 0
  dialogVisible.value = true
}

const openEdit = (node) => {
  dialogMode.value = 'edit'
  Object.assign(form, {
    deptCode: node.deptCode || '',
    deptName: node.deptName || '',
    parentCode: node.parentCode || '',
    grade: node.grade ?? 0,
    endMark: node.endMark || '',
    kind: node.kind || '',
    b0110: node.b0110 || ''
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
      grade: form.grade ?? 0,
      endMark: form.endMark,
      kind: form.kind,
      b0110: form.b0110
    }
    const res = dialogMode.value === 'create' ? await createDept(payload) : await updateDept(payload)
    if (res.code === 200) {
      ElMessage.success('操作成功')
      dialogVisible.value = false
      await loadTree()
    } else {
      ElMessage.error(res.message || '操作失败')
    }
  } catch (e) {
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
  }
}

const handleDeleteTree = async (node) => {
  try {
    await ElMessageBox.confirm(
      `确认删除部门 ${node.deptName}（${node.deptCode}）及其所有下级部门？`,
      '删除确认',
      { type: 'warning' }
    )
  } catch {
    return
  }
  submitLoading.value = true
  try {
    const res = await deleteDept({ deptCode: node.deptCode, cascade: 1 })
    if (res.code === 200) {
      ElMessage.success('删除成功')
      await loadTree()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (e) {
    ElMessage.error('删除失败')
  } finally {
    submitLoading.value = false
  }
}

const allowDrop = (_draggingNode, _dropNode, type) => {
  return type === 'inner' || type === 'before' || type === 'after'
}

const handleNodeDrop = async (draggingNode, dropNode, dropType) => {
  const drag = draggingNode?.data
  const drop = dropNode?.data
  if (!drag || !drop) return

  let newParentCode = ''
  if (dropType === 'inner') {
    newParentCode = drop.deptCode
  } else {
    newParentCode = drop.parentCode || ''
  }

  if (drag.deptCode === newParentCode) {
    ElMessage.warning('不能移动到自身下面')
    await loadTree()
    return
  }

  try {
    const res = await moveDept({ deptCode: drag.deptCode, newParentCode })
    if (res.code === 200) {
      ElMessage.success('移动成功')
      await loadTree()
    } else {
      ElMessage.error(res.message || '移动失败')
      await loadTree()
    }
  } catch (e) {
    ElMessage.error('移动失败')
    await loadTree()
  }
}

onMounted(() => {
  loadTree()
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

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.tree-container {
  min-height: 400px;
}

.dept-tree-table {
  width: 100%;
}

.dept-name-cell {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.dept-name-icon {
  color: var(--el-text-color-secondary);
}

.dept-tree-view {
  background-color: #fff;
  padding: 20px;
  border-radius: 4px;
}

.tree-node {
  display: flex;
  align-items: center;
  flex: 1;
  font-size: 14px;
  padding-right: 8px;
  gap: 10px;
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

.node-actions {
  margin-left: auto;
  display: flex;
  gap: 6px;
  align-items: center;
}

:deep(.el-tree-node__content) {
  height: 40px;
  line-height: 40px;
}

:deep(.el-tree-node__label) {
  width: 100%;
}
</style>

