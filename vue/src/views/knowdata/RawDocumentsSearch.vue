<template>
  <div class="raw-documents-search">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>库内全文检索</span>
        </div>
      </template>

      <el-form :inline="true" :model="form" class="search-form">
        <el-form-item label="关键字">
          <el-input
            v-model="form.keyword"
            placeholder="请输入关键字"
            clearable
            style="width: 260px"
            @keyup.enter="handleSearch(true)"
          />
        </el-form-item>
        <el-form-item label="知识库">
          <el-select
            v-model="form.documentCode"
            placeholder="全部"
            clearable
            filterable
            style="width: 220px"
          >
            <el-option
              v-for="item in documentsTypeList"
              :key="item.code"
              :label="item.name"
              :value="item.code"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="标签">
          <el-select
            v-model="form.tag"
            placeholder="全部"
            clearable
            filterable
            style="width: 200px"
          >
            <el-option
              v-for="tag in availableTags"
              :key="tag"
              :label="tag"
              :value="tag"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="审核状态">
          <el-select v-model="form.isAudit" placeholder="全部" clearable style="width: 140px">
            <el-option label="已审核" value="1" />
            <el-option label="未审核" value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch(true)" :loading="loading">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="handleReset">
            <el-icon><Refresh /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>

      <el-table v-loading="loading" :data="tableData" border stripe style="width: 100%">
        <el-table-column prop="documentCode" label="知识库" width="160">
          <template #default="{ row }">
            {{ getDocumentTypeName(row.documentCode) }}
          </template>
        </el-table-column>
        <el-table-column prop="fileName" label="文件名" width="320">
          <template #default="{ row }">
            <span class="file-name-link" @click="handleOpen(row)">{{ row.fileName }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="tag" label="标签" width="140">
          <template #default="{ row }">
            <el-tag v-if="row.tag" size="small" type="info">{{ row.tag }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="snippet" label="命中片段">
          <template #default="{ row }">
            <div class="snippet" v-html="highlight(row.snippet)"></div>
          </template>
        </el-table-column>
        <el-table-column prop="updatedAt" label="更新时间" width="180">
          <template #default="{ row }">
            {{ row.updatedAt ? formatTime(row.updatedAt) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              :disabled="row.isAudit !== 1"
              :loading="qaLoadingId === row.id"
              @click="handleViewQaPairs(row)"
            >
              问答
            </el-button>
          </template>
        </el-table-column>
      </el-table>

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

      <el-dialog
        v-model="qaDialogVisible"
        title="文档问答对"
        width="72%"
        @closed="qaData = null"
      >
        <div class="qa-header" v-if="qaData">
          <el-tag type="info">文件：{{ qaData.fileName || '-' }}</el-tag>
          <el-tag type="success">知识库：{{ getDocumentTypeName(qaData.documentCode) || qaData.documentCode || '-' }}</el-tag>
          <el-tag>总数：{{ qaData.total || 0 }}</el-tag>
        </div>
        <el-table :data="qaData?.list || []" border stripe max-height="460" size="small">
          <el-table-column prop="chunkIndex" label="分块" width="90" />
          <el-table-column prop="question" label="Q（content）" min-width="260" show-overflow-tooltip />
          <el-table-column prop="answer" label="A（辅助结果）" min-width="360" show-overflow-tooltip />
          <el-table-column prop="qdrantPointId" label="Qdrant点ID" min-width="220" show-overflow-tooltip />
        </el-table>
        <template #footer>
          <el-button type="primary" @click="qaDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { searchRawDocuments, listDocumentsType, listRawDocumentQaPairs } from '@/api/knowdata'
import { listFrTag } from '@/api/knowsource'

const router = useRouter()
const loading = ref(false)
const tableData = ref([])
const documentsTypeList = ref([])
const availableTags = ref([])
const qaLoadingId = ref(null)
const qaDialogVisible = ref(false)
const qaData = ref(null)

const form = reactive({
  keyword: '',
  documentCode: '',
  tag: '',
  isAudit: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

const escapeHtml = (str) => {
  const s = String(str || '')
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}
const escapeRegExp = (str) => String(str || '').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')

const highlight = (text) => {
  const kw = String(form.keyword || '').trim()
  const safe = escapeHtml(text || '')
  if (!kw) return safe
  const re = new RegExp(escapeRegExp(kw), 'gi')
  return safe.replace(re, (m) => `<mark class="hit-highlight">${m}</mark>`)
}

const getDocumentTypeName = (code) => {
  if (!code) return '-'
  const docType = documentsTypeList.value.find(item => item.code === code)
  return docType ? docType.name : code
}

const loadDocumentsTypeList = async () => {
  try {
    const res = await listDocumentsType({})
    if (res.code === 200 && res.data) {
      documentsTypeList.value = res.data.list || []
    }
  } catch (e) {
    // ignore
  }
}

const loadAvailableTags = async () => {
  try {
    const res = await listFrTag({ page: 1, pageSize: 1000 })
    if (res.code === 200 && res.data) {
      availableTags.value = (res.data.list || []).map(item => item.tag).filter(tag => tag)
    }
  } catch (e) {
    availableTags.value = []
  }
}

const handleSearch = async (resetPage = false) => {
  const kw = String(form.keyword || '').trim()
  if (!kw) {
    ElMessage.warning('请输入关键字')
    return
  }
  if (resetPage) pagination.page = 1

  loading.value = true
  try {
    const res = await searchRawDocuments({
      keyword: kw,
      documentCode: form.documentCode,
      tag: form.tag,
      isAudit: form.isAudit,
      page: pagination.page,
      pageSize: pagination.pageSize
    })
    if (res.code === 200 && res.data) {
      tableData.value = res.data.list || []
      pagination.total = res.data.total || 0
    } else {
      ElMessage.error(res.message || res.msg || '查询失败')
      tableData.value = []
      pagination.total = 0
    }
  } catch (e) {
    ElMessage.error('查询失败，请稍后重试')
    tableData.value = []
    pagination.total = 0
  } finally {
    loading.value = false
  }
}

const handleReset = () => {
  form.keyword = ''
  form.documentCode = ''
  form.tag = ''
  form.isAudit = ''
  tableData.value = []
  pagination.page = 1
  pagination.total = 0
}

const handleOpen = (row) => {
  const kw = String(form.keyword || '').trim()
  router.push({
    name: 'RawDocumentContent',
    params: { id: row.id },
    query: kw ? { q: kw } : {}
  })
}

const handleViewQaPairs = async (row) => {
  if (!row || row.isAudit !== 1) return
  qaLoadingId.value = row.id
  try {
    const res = await listRawDocumentQaPairs({
      id: row.id,
      page: 1,
      pageSize: 200
    })
    if (res.code === 200) {
      qaData.value = {
        list: res?.data?.list || [],
        total: res?.data?.total || 0,
        fileName: row.fileName,
        documentCode: row.documentCode
      }
      qaDialogVisible.value = true
    } else {
      ElMessage.error(res.message || res.msg || res.info || '查询问答失败')
    }
  } catch (e) {
    ElMessage.error('查询问答失败，请稍后重试')
  } finally {
    qaLoadingId.value = null
  }
}

const handleSizeChange = () => {
  pagination.page = 1
  handleSearch(false)
}

const handlePageChange = () => {
  handleSearch(false)
}

onMounted(() => {
  loadDocumentsTypeList()
  loadAvailableTags()
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

.file-name-link {
  color: #409eff;
  cursor: pointer;
  text-decoration: none;
}

.file-name-link:hover {
  color: #66b1ff;
  text-decoration: underline;
}

.snippet {
  white-space: pre-wrap;
  word-break: break-word;
  color: #303133;
}

.snippet :deep(mark.hit-highlight) {
  background: #fff3bf;
  color: #1f2328;
  padding: 0 2px;
  border-radius: 2px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.qa-header {
  margin-bottom: 10px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
</style>

