<template>
  <div class="ai-call-stats">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>AI 调用统计</span>
          <el-button type="primary" @click="query" :loading="loading">
            <el-icon><Search /></el-icon>
            查询
          </el-button>
        </div>
      </template>
      <p class="desc">按时间范围统计 AI 调用总次数、去重用户数、模型种类数、平均耗时与总字数。</p>
      <div class="range-buttons">
        <span class="range-label">快捷范围：</span>
        <el-button size="small" @click="setRange('today')">今天</el-button>
        <el-button size="small" @click="setRange('yesterday')">昨天</el-button>
        <el-button size="small" @click="setRange('thisWeek')">本周</el-button>
        <el-button size="small" @click="setRange('lastWeek')">上周</el-button>
        <el-button size="small" @click="setRange('thisMonth')">本月</el-button>
        <el-button size="small" @click="setRange('lastMonth')">上月</el-button>
        <el-button size="small" @click="setRange('thisQuarter')">本季度</el-button>
        <el-button size="small" @click="setRange('lastQuarter')">上季度</el-button>
        <el-button size="small" @click="setRange('thisYear')">本年</el-button>
        <el-button size="small" @click="setRange('lastYear')">上年</el-button>
      </div>
      <el-form :inline="true" class="query-form">
        <el-form-item label="开始时间">
          <el-date-picker
            v-model="startTime"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            format="YYYY-MM-DD HH:mm:ss"
            placeholder="选择开始时间"
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item label="结束时间">
          <el-date-picker
            v-model="endTime"
            type="datetime"
            value-format="YYYY-MM-DD HH:mm:ss"
            format="YYYY-MM-DD HH:mm:ss"
            placeholder="选择结束时间"
            style="width: 200px"
          />
        </el-form-item>
      </el-form>

      <div v-if="data" class="stats-result">
        <el-row :gutter="16">
          <el-col :xs="24" :sm="8">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-value">{{ data.totalCount }}</div>
              <div class="stat-label">调用总次数</div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-value">{{ data.userCount }}</div>
              <div class="stat-label">去重用户数</div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-value">{{ data.modelCount }}</div>
              <div class="stat-label">模型种类数</div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-value">{{ formatAvgMs(data.avgCostMs) }}</div>
              <div class="stat-label">平均耗时（ms）</div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-value">{{ data.sumQuestionCharCount }}</div>
              <div class="stat-label">问题总字数</div>
            </el-card>
          </el-col>
          <el-col :xs="24" :sm="8">
            <el-card shadow="hover" class="stat-card">
              <div class="stat-value">{{ data.sumOutputCharCount }}</div>
              <div class="stat-label">输出总字数</div>
            </el-card>
          </el-col>
        </el-row>
        <div v-if="data.modelNames && data.modelNames.length" class="model-names">
          <div class="model-names-title">模型列表</div>
          <el-tag
            v-for="name in data.modelNames"
            :key="name"
            class="model-tag"
            size="small"
          >
            {{ name }}
          </el-tag>
        </div>
      </div>
      <div v-else-if="hasQueried && !loading" class="empty-tip">
        请选择时间范围后点击「查询」，或当前时间范围内无数据。
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { queryAiCallStats } from '@/api/knowsource'

const loading = ref(false)
const hasQueried = ref(false)
const data = ref(null)

// 默认最近 7 天
const now = new Date()
const defaultEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
const defaultStart = new Date(defaultEnd)
defaultStart.setDate(defaultStart.getDate() - 6)
defaultStart.setHours(0, 0, 0, 0)

const startTime = ref(formatDateTime(defaultStart))
const endTime = ref(formatDateTime(defaultEnd))

function formatDateTime(d) {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  const s = String(d.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${day} ${h}:${min}:${s}`
}

function startOfDay(d) {
  const x = new Date(d)
  x.setHours(0, 0, 0, 0)
  return x
}
function endOfDay(d) {
  const x = new Date(d)
  x.setHours(23, 59, 59, 999)
  return x
}

function getRange(rangeType) {
  const now = new Date()
  let start, end
  switch (rangeType) {
    case 'today':
      start = startOfDay(now)
      end = endOfDay(now)
      break
    case 'yesterday': {
      const y = new Date(now)
      y.setDate(y.getDate() - 1)
      start = startOfDay(y)
      end = endOfDay(y)
      break
    }
    case 'thisWeek': {
      const d = new Date(now)
      const day = d.getDay()
      const diff = day === 0 ? 6 : day - 1
      d.setDate(d.getDate() - diff)
      start = startOfDay(d)
      end = endOfDay(now)
      break
    }
    case 'lastWeek': {
      const d = new Date(now)
      const day = d.getDay()
      const diff = day === 0 ? 6 : day - 1
      d.setDate(d.getDate() - diff - 7)
      start = startOfDay(d)
      const lastSun = new Date(d)
      lastSun.setDate(lastSun.getDate() + 6)
      end = endOfDay(lastSun)
      break
    }
    case 'thisMonth':
      start = startOfDay(new Date(now.getFullYear(), now.getMonth(), 1))
      end = endOfDay(new Date(now.getFullYear(), now.getMonth() + 1, 0))
      break
    case 'lastMonth':
      start = startOfDay(new Date(now.getFullYear(), now.getMonth() - 1, 1))
      end = endOfDay(new Date(now.getFullYear(), now.getMonth(), 0))
      break
    case 'thisQuarter': {
      const q = Math.floor(now.getMonth() / 3) + 1
      const qStartMonth = (q - 1) * 3
      start = startOfDay(new Date(now.getFullYear(), qStartMonth, 1))
      end = endOfDay(new Date(now.getFullYear(), qStartMonth + 3, 0))
      break
    }
    case 'lastQuarter': {
      const q = Math.floor(now.getMonth() / 3) + 1
      const year = q === 1 ? now.getFullYear() - 1 : now.getFullYear()
      const qStartMonth = q === 1 ? 9 : (q - 2) * 3
      start = startOfDay(new Date(year, qStartMonth, 1))
      end = endOfDay(new Date(year, qStartMonth + 3, 0))
      break
    }
    case 'thisYear':
      start = startOfDay(new Date(now.getFullYear(), 0, 1))
      end = endOfDay(now)
      break
    case 'lastYear':
      start = startOfDay(new Date(now.getFullYear() - 1, 0, 1))
      end = endOfDay(new Date(now.getFullYear() - 1, 11, 31))
      break
    default:
      return null
  }
  return { start: formatDateTime(start), end: formatDateTime(end) }
}

function setRange(rangeType) {
  const r = getRange(rangeType)
  if (!r) return
  startTime.value = r.start
  endTime.value = r.end
  query()
}

function formatAvgMs(v) {
  if (v === null || v === undefined) return 0
  // 保留 2 位小数，避免展示过长
  return Number(v).toFixed(2)
}

const query = async () => {
  if (!startTime.value || !endTime.value) {
    ElMessage.warning('请选择开始时间和结束时间')
    return
  }
  loading.value = true
  hasQueried.value = true
  data.value = null
  try {
    const res = await queryAiCallStats({
      startTime: startTime.value,
      endTime: endTime.value
    })
    if (res.code === 200 || res.code === 0) {
      data.value = res.data || null
    } else {
      ElMessage.error(res.message || '查询失败')
    }
  } catch (e) {
    ElMessage.error(e.message || '请求失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  query()
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
.range-buttons {
  margin-bottom: 16px;
}
.range-label {
  font-size: 14px;
  color: #606266;
  margin-right: 8px;
  vertical-align: middle;
}
.range-buttons .el-button {
  margin-right: 8px;
  margin-bottom: 6px;
}
.query-form {
  margin-bottom: 20px;
}
.stats-result {
  margin-top: 16px;
}
.stat-card {
  text-align: center;
  margin-bottom: 16px;
}
.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #409eff;
}
.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 8px;
}
.model-names {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #ebeef5;
}
.model-names-title {
  font-size: 14px;
  color: #606266;
  margin-bottom: 12px;
}
.model-tag {
  margin-right: 8px;
  margin-bottom: 8px;
}
.empty-tip {
  color: #909399;
  font-size: 14px;
  margin-top: 20px;
}
</style>
