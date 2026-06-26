<template>
  <div class="qdrant-collection-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>Qdrant 集合列表</span>
          <el-button type="primary" @click="loadData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <!-- 集合列表 -->
      <el-table
        v-loading="loading"
        :data="collectionList"
        border
        stripe
        style="width: 100%"
        row-key="name"
      >
        <el-table-column prop="name" label="集合名称" width="400" show-overflow-tooltip />
        <el-table-column prop="index" label="序号" width="80" type="index" />
      </el-table>

      <!-- 空状态 -->
      <div v-if="!loading && collectionList.length === 0" class="empty-state">
        <el-empty description="暂无集合数据" />
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { getQdrantCollectionList } from '@/api/knowsource'

const loading = ref(false)
const collectionList = ref([])

// 加载集合列表
const loadData = async () => {
  loading.value = true
  try {
    const res = await getQdrantCollectionList()
    if (res.code === 200 && res.data && res.data.list) {
      // 将集合列表转换为表格需要的格式
      collectionList.value = res.data.list.map((name, index) => ({
        name,
        index: index + 1
      }))
    } else {
      ElMessage.error(res.message || '获取集合列表失败')
    }
  } catch (error) {
    console.error('加载集合列表失败:', error)
    ElMessage.error('加载集合列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 18px;
  font-weight: 500;
}

.empty-state {
  margin-top: 50px;
  text-align: center;
}
</style>