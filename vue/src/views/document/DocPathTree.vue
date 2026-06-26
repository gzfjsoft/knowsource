<template>
  <div class="doc-path-tree">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>文档路径文件树</span>
          <el-button type="primary" @click="loadTree" :loading="loading" size="small">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      
      <div v-loading="loading" class="tree-container">
        <el-tree
          v-if="treeData.length > 0"
          :data="treeData"
          :props="treeProps"
          default-expand-all
          :expand-on-click-node="false"
          node-key="path"
          class="doc-tree-view"
        >
          <template #default="{ node, data }">
            <span class="tree-node">
              <el-icon v-if="data.isDir" class="node-icon">
                <Folder />
              </el-icon>
              <el-icon v-else class="node-icon">
                <Document />
              </el-icon>
              <span class="node-label">{{ data.name }}</span>
              <span class="node-info">
                <el-tag v-if="data.isDir" size="small" type="info">目录</el-tag>
                <el-tag v-else size="small" type="success">文件</el-tag>
                <el-tag v-if="!data.isDir && data.size" size="small" type="warning">
                  {{ formatSize(data.size) }}
                </el-tag>
                <el-tag v-if="data.path" size="small" type="info" class="path-tag">
                  {{ data.path }}
                </el-tag>
              </span>
            </span>
          </template>
        </el-tree>
        <el-empty v-else description="暂无数据" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Folder, Document } from '@element-plus/icons-vue'
import { getDocPathTree } from '@/api/knowsource'

const loading = ref(false)
const treeData = ref([])

const treeProps = {
  children: 'children',
  label: 'name'
}

const formatSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const loadTree = async () => {
  loading.value = true
  try {
    const res = await getDocPathTree({})
    if (res.code === 200 && res.data && res.data.tree) {
      treeData.value = res.data.tree
    } else {
      ElMessage.warning(res.message || '暂无数据')
      treeData.value = []
    }
  } catch (error) {
    ElMessage.error('加载文件树失败: ' + (error.message || '未知错误'))
    treeData.value = []
  } finally {
    loading.value = false
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

.tree-container {
  min-height: 400px;
}

.doc-tree-view {
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
}

.node-icon {
  margin-right: 8px;
  color: #409eff;
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

.path-tag {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.el-tree-node__content) {
  height: 40px;
  line-height: 40px;
}

:deep(.el-tree-node__label) {
  width: 100%;
}
</style>

