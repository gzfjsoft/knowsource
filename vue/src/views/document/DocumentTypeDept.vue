<template>
  <div class="document-type-dept">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>按部门授权</span>
          <el-button type="primary" @click="handleSave" :loading="saveLoading" :disabled="!selectedDocumentTypeCode">
            <el-icon><Check /></el-icon>
            保存绑定
          </el-button>
        </div>
      </template>
      
      <div class="content-container">
        <!-- 左侧知识库列表 -->
        <div class="left-panel">
          <el-card>
            <template #header>
              <div class="panel-header">
                <span>知识库列表</span>
                <el-button type="primary" @click="loadDocumentTypes" :loading="docTypeLoading" size="small">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </div>
            </template>
            
            <div class="search-box">
              <el-input
                v-model="docTypeSearchKeyword"
                placeholder="搜索知识库"
                clearable
                prefix-icon="Search"
              />
            </div>
            
            <div v-loading="docTypeLoading" class="doc-type-list">
              <div
                v-for="docType in filteredDocumentTypes"
                :key="docType.code"
                class="doc-type-item"
                :class="{ active: selectedDocumentTypeCode === docType.code }"
                @click="handleSelectDocumentType(docType)"
              >
                <div class="doc-type-name">{{ docType.name }}</div>
                <div v-if="docType.description" class="doc-type-desc">{{ docType.description }}</div>
              </div>
              <el-empty v-if="filteredDocumentTypes.length === 0" description="暂无知识库" />
            </div>
          </el-card>
        </div>

        <!-- 右侧部门树 -->
        <div class="right-panel">
          <el-card>
            <template #header>
              <div class="panel-header">
                <span>部门树</span>
                <el-button type="primary" @click="loadDeptTree" :loading="deptTreeLoading" size="small">
                  <el-icon><Refresh /></el-icon>
                  刷新
                </el-button>
              </div>
            </template>
            
            <div v-loading="deptTreeLoading" class="tree-container">
              <div v-if="!selectedDocumentTypeCode" class="empty-tip">
                <el-empty description="请先选择左侧知识库" />
              </div>
              <el-tree
                v-else-if="deptTreeData.length > 0"
                ref="deptTreeRef"
                :data="deptTreeData"
                :props="{ children: 'children', label: 'deptName' }"
                node-key="deptCode"
                show-checkbox
                :default-expanded-keys="defaultExpandedKeys"
                :check-strictly="false"
                class="dept-tree-multi-select"
                @check="handleDeptTreeCheck"
              >
                <template #default="{ node, data }">
                  <span class="tree-node">
                    <span class="node-label">{{ data.deptName }}</span>
                    <span class="node-code">{{ data.deptCode }}</span>
                  </span>
                </template>
              </el-tree>
              <div v-else class="empty-tip">
                <el-empty description="暂无部门数据" />
              </div>
            </div>
          </el-card>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Check, Refresh, Search } from '@element-plus/icons-vue'
import { getDeptTree, listDeptDocumentType, createDeptDocumentType, deleteDeptDocumentType } from '@/api/knowsource'
import { listDocumentsType } from '@/api/knowdata'

const saveLoading = ref(false)
const docTypeLoading = ref(false)
const deptTreeLoading = ref(false)
const documentTypeList = ref([])
const deptTreeData = ref([])
const deptTreeRef = ref(null)
const selectedDocumentTypeCode = ref('')
const selectedDeptCodes = ref([])
const docTypeSearchKeyword = ref('')

// 默认展开第一层节点
const defaultExpandedKeys = computed(() => {
  const data = deptTreeData.value || []
  return data.map(node => node.deptCode).filter(Boolean)
})

// 过滤后的知识库列表
const filteredDocumentTypes = computed(() => {
  if (!docTypeSearchKeyword.value) {
    return documentTypeList.value
  }
  const keyword = docTypeSearchKeyword.value.toLowerCase()
  return documentTypeList.value.filter(item => 
    item.name.toLowerCase().includes(keyword) || 
    item.code.toLowerCase().includes(keyword) ||
    (item.description && item.description.toLowerCase().includes(keyword))
  )
})

// 加载知识库列表
const loadDocumentTypes = async () => {
  docTypeLoading.value = true
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
  } finally {
    docTypeLoading.value = false
  }
}

// 加载部门树
const loadDeptTree = async () => {
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

// 选择知识库
const handleSelectDocumentType = async (docType) => {
  selectedDocumentTypeCode.value = docType.code
  // 等待树组件渲染完成
  await nextTick()
  // 加载该知识库已关联的部门
  await loadDocumentTypeDepts(docType.code)
}

// 加载知识库已关联的部门
const loadDocumentTypeDepts = async (documentTypeCode) => {
  if (!documentTypeCode) {
    selectedDeptCodes.value = []
    await nextTick()
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
      // 等待 DOM 更新后再设置选中状态
      await nextTick()
      if (deptTreeRef.value) {
        deptTreeRef.value.setCheckedKeys(selectedDeptCodes.value)
      }
    } else {
      selectedDeptCodes.value = []
      await nextTick()
      if (deptTreeRef.value) {
        deptTreeRef.value.setCheckedKeys([])
      }
    }
  } catch (error) {
    ElMessage.error('加载关联部门失败')
    selectedDeptCodes.value = []
    await nextTick()
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

// 保存绑定
const handleSave = async () => {
  if (!selectedDocumentTypeCode.value) {
    ElMessage.warning('请先选择知识库')
    return
  }
  
  saveLoading.value = true
  try {
    // 获取当前已关联的部门
    const currentRes = await listDeptDocumentType({
      documentTypeCode: selectedDocumentTypeCode.value,
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
    let deleteCount = 0
    for (const deptCode of toDelete) {
      const deleteRes = await listDeptDocumentType({
        documentTypeCode: selectedDocumentTypeCode.value,
        deptCode: deptCode,
        page: 1,
        pageSize: 1
      })
      if (deleteRes.code === 200 && deleteRes.data && deleteRes.data.list && deleteRes.data.list.length > 0) {
        const delRes = await deleteDeptDocumentType({ id: deleteRes.data.list[0].id })
        if (delRes.code === 200) {
          deleteCount++
        }
      }
    }
    
    // 批量新增
    let addCount = 0
    for (const deptCode of toAdd) {
      const addRes = await createDeptDocumentType({
        deptCode: deptCode,
        documentTypeCode: selectedDocumentTypeCode.value
      })
      if (addRes.code === 200) {
        addCount++
      }
    }
    
    if (addCount > 0 || deleteCount > 0) {
      ElMessage.success(`保存成功：新增 ${addCount} 个，删除 ${deleteCount} 个`)
      // 重新加载关联的部门
      await loadDocumentTypeDepts(selectedDocumentTypeCode.value)
    } else {
      ElMessage.info('没有变更')
    }
  } catch (error) {
    ElMessage.error('保存失败，请稍后重试')
  } finally {
    saveLoading.value = false
  }
}

onMounted(() => {
  loadDocumentTypes()
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

.content-container {
  display: flex;
  gap: 20px;
  min-height: calc(100vh - 200px);
}

.left-panel {
  width: 350px;
  flex-shrink: 0;
}

.right-panel {
  flex: 1;
  min-width: 0;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 16px;
  font-weight: 500;
}

.search-box {
  margin-bottom: 15px;
}

.doc-type-list {
  max-height: calc(100vh - 300px);
  overflow-y: auto;
}

.doc-type-item {
  padding: 12px;
  margin-bottom: 8px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.doc-type-item:hover {
  border-color: #409eff;
  background-color: #f5f7fa;
}

.doc-type-item.active {
  border-color: #409eff;
  background-color: #ecf5ff;
}

.doc-type-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.doc-type-code {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.doc-type-desc {
  font-size: 12px;
  color: #606266;
  margin-top: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tree-container {
  min-height: 400px;
  max-height: calc(100vh - 300px);
  overflow-y: auto;
}

.empty-tip {
  padding: 40px 0;
  text-align: center;
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
</style>
