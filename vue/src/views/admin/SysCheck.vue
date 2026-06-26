<template>
  <div class="sys-check">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>系统依赖检查</span>
          <el-button type="primary" @click="loadData" :loading="loading">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>
      <p class="desc">检查 llm chat、llm embedding、llm rerank、Qdrant、Redis、Mysql、MinerU、Mail(SMTP) 是否可访问。</p>
      <el-table
        v-loading="loading"
        :data="rows"
        border
        stripe
        style="width: 100%"
      >
        <el-table-column prop="name" label="服务" width="140" />
        <el-table-column prop="ok" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.ok ? 'success' : 'danger'" size="small">
              {{ row.ok ? '正常' : '异常' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="value" label="地址/配置值" min-width="200" show-overflow-tooltip />
        <el-table-column prop="message" label="说明" min-width="120" show-overflow-tooltip />
        <el-table-column label="类型" width="90">
          <template #default="{ row }">
            <el-tag v-if="row.type" size="small" :type="row.type === 'ollama' ? 'warning' : 'primary'">
              {{ row.type }}
            </el-tag>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="版本" width="80">
          <template #default="{ row }">
            <span v-if="row.version">{{ row.version }}</span>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
        <el-table-column label="模型 ID (vLLM)" min-width="200">
          <template #default="{ row }">
            <template v-if="row.modelIds && row.modelIds.length">
              <el-tag
                v-for="id in row.modelIds"
                :key="id"
                size="small"
                type="info"
                style="margin-right: 6px; margin-bottom: 4px"
              >
                {{ id }}
              </el-tag>
            </template>
            <span v-else class="text-muted">—</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { sysCheck } from '@/api/knowsource'

const loading = ref(false)
const data = ref(null)

const rows = computed(() => {
  const d = data.value?.data
  if (!d) return []
  return [
    { name: 'llm chat', ...d.vllmchat },
    { name: 'llm embedding', ...d.vllmembedding },
    { name: 'llm rerank', ...d.vllmreranker },
    { name: 'Qdrant', ...d.qdrant },
    { name: 'Redis', ...d.redis },
    { name: 'Mysql', ...d.mysql },
    { name: 'MinerU', ...d.mineru },
    { name: 'Mail(SMTP)', ...d.mail }
  ]
})

const loadData = async () => {
  loading.value = true
  try {
    const res = await sysCheck()
    if (res.code === 200 || res.code === 0) {
      data.value = res
    } else {
      ElMessage.error(res.message || '检查失败')
    }
  } catch (e) {
    ElMessage.error(e.message || '请求失败')
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
.desc {
  color: #606266;
  font-size: 14px;
  margin: 0 0 16px 0;
}
.text-muted {
  color: #909399;
  font-size: 13px;
}
</style>
