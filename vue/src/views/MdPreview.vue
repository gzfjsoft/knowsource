<template>
  <div class="md-preview">
    <div v-loading="loading" class="preview-container">
      <div v-if="error" class="preview-error">
        <el-alert type="error" :title="error" show-icon />
      </div>
      <template v-else-if="doc">
        <div class="preview-header">
          <el-button type="primary" link @click="goToList">
            <el-icon><ArrowLeft /></el-icon>
            返回列表
          </el-button>
          <el-tag type="info">{{ doc.fileName }}</el-tag>
          <el-tag v-if="documentTypeName" type="success">{{ documentTypeName }}</el-tag>
          <el-tag v-if="doc.tag" type="warning">{{ doc.tag }}</el-tag>
        </div>
        <el-tabs v-model="activeTab" class="preview-tabs">
          <el-tab-pane label="预览" name="preview">
            <div
              v-if="htmlContent"
              class="preview-content markdown-body"
              v-html="htmlContent"
            />
            <el-empty v-else-if="!loading" description="该文档暂无 Markdown 内容" />
          </el-tab-pane>
          <el-tab-pane label="源文件" name="source">
            <pre class="preview-source">{{ rawContent || '（无内容）' }}</pre>
          </el-tab-pane>
        </el-tabs>
      </template>
      <el-empty v-else-if="!loading && !error" description="该文档暂无 Markdown 内容" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { parseMarkdownWithAssets } from '@/utils/markdown'
import { getRawDocuments } from '@/api/knowdata'
import { listMyDocumentType } from '@/api/knowsource'
import { navigateBackToRawDocumentsList } from '@/utils/rawDocumentsListNavigation'

const route = useRoute()
const router = useRouter()

function goToList () {
  navigateBackToRawDocumentsList(router, route)
}
const loading = ref(false)
const error = ref('')
const doc = ref(null)
const documentTypes = ref([])

// 知识库名称：按 documentCode 从文档类型列表取名称，不展示 code
const documentTypeName = computed(() => {
  const code = doc.value?.documentCode
  if (!code) return ''
  const item = documentTypes.value.find(d => d.documentTypeCode === code)
  return item ? item.documentTypeName : ''
})
const activeTab = ref('preview')

// 图片 base：/api/v1/md/<真实相对目录>/（基于 filePath + fileList 推导）
const baseUrl = computed(() => {
  const d = doc.value
  if (!d?.filePath) return ''
  const origin = window.location.origin
  const normalizedFilePath = stripDocumentRootPrefix(String(d.filePath || ''))
  if (!normalizedFilePath) return ''

  const fileList = parseFileList(d.fileList)
  const mdEntry = fileList.find(v => String(v || '').toLowerCase().endsWith('.md')) || ''
  const mdDirRaw = mdEntry ? normalizeRelPath(mdEntry).split('/').slice(0, -1).join('/') : ''
  const mdDir = normalizeMdDirByFileName(mdDirRaw, d.fileName)
  const lower = normalizedFilePath.toLowerCase()

  let mdBasePath = ''
  if (lower.endsWith('.md') || lower.endsWith('.txt')) {
    // filePath 已是文本文件：以所在目录作为相对资源根
    mdBasePath = normalizedFilePath.split('/').slice(0, -1).join('/')
    if (mdDir) mdBasePath = normalizeRelPath([mdBasePath, mdDir].filter(Boolean).join('/'))
  } else {
    // filePath 指向原始文件（pdf/doc/docx/xlsx...）：资源在 <filePath>.file 下
    mdBasePath = normalizedFilePath + '.file'
    if (mdDir) mdBasePath = normalizeRelPath(`${mdBasePath}/${mdDir}`)
  }
  if (!mdBasePath) return ''

  const encodedPath = encodePathSegments(mdBasePath)
  return `${origin}/api/v1/md/${encodedPath}/`
})

const rawContent = computed(() => doc.value?.content ?? '')

// 规范化相对路径（去 ./、统一 /）
function normalizeRelPath(relative) {
  let p = (relative || '').trim()
  if (p.startsWith('./')) p = p.slice(2)
  if (p.startsWith('/')) p = p.slice(1)
  p = p.replace(/\\/g, '/')
  return p
}

function parseFileList(fileListRaw) {
  if (!fileListRaw) return []
  if (Array.isArray(fileListRaw)) return fileListRaw
  try {
    const arr = JSON.parse(fileListRaw)
    return Array.isArray(arr) ? arr : []
  } catch (_) {
    return []
  }
}

function stripDocumentRootPrefix(path) {
  const p = normalizeRelPath(path)
  const roots = ['files/document/', 'document/']
  for (const root of roots) {
    if (p.startsWith(root)) return p.slice(root.length)
  }
  return p
}

function normalizeMdDirByFileName(mdDir, fileName) {
  const dir = normalizeRelPath(mdDir)
  if (!dir) return ''
  const stem = String(fileName || '').replace(/\.[^/.]+$/, '')
  if (!stem) return dir
  const prefix = `${stem}/auto`
  if (dir === prefix) return 'auto'
  if (dir.startsWith(`${prefix}/`)) return dir.slice(stem.length + 1)
  return dir
}

function encodePathSegments(path) {
  return String(path || '')
    .split('/')
    .filter(Boolean)
    .map(seg => encodeURIComponent(seg))
    .join('/')
}

// 相对路径替换规则：
// - 若为 <folder>/<otherfiles> → /api/v1/md/<doctypecode>/<filename>.file/<folder>/<otherfiles>
// - 若为 <otherfiles> → /api/v1/md/<doctypecode>/<filename>.file/<otherfiles>
function resolveImageSrc(src) {
  if (/^https?:\/\//i.test(src) || src.startsWith('data:')) return src
  const base = baseUrl.value
  if (!base) return src
  const norm = normalizeRelPath(src)
  if (!norm) return src
  const slash = norm.startsWith('/') ? '' : '/'
  return base.replace(/\/$/, '') + slash + norm
}

// Markdown 转 HTML：相对路径的图片和链接按上述规则加前缀
const htmlContent = computed(() => {
  const md = rawContent.value
  if (!md) return ''
  const base = baseUrl.value
  return parseMarkdownWithAssets(md, {
    resolveImageSrc: (src) => resolveImageSrc(src),
    resolveLinkHref: (href) => {
      if (!base) return href
      const norm = normalizeRelPath(href)
      if (!norm) return href
      return base.replace(/\/$/, '') + (norm.startsWith('/') ? '' : '/') + norm
    },
  })
})

async function loadDocument() {
  const id = route.params.id
  if (!id) {
    error.value = '缺少文档 ID'
    return
  }
  loading.value = true
  error.value = ''
  doc.value = null
  try {
    const res = await getRawDocuments({ id: parseInt(id, 10) })
    if (res.code === 200 && res.data) {
      doc.value = res.data
    } else {
      error.value = res.message || '获取文档失败'
      doc.value = null
    }
  } catch (e) {
    error.value = e?.response?.data?.info || e?.message || '加载文档失败'
    doc.value = null
  } finally {
    loading.value = false
  }
}

watch(
  () => route.params.id,
  (newId) => {
    if (newId) loadDocument()
  },
  { immediate: false }
)
onMounted(async () => {
  try {
    const res = await listMyDocumentType()
    if (res?.code === 200 && res?.data?.list) documentTypes.value = res.data.list
  } catch (_) {}
  loadDocument()
})
</script>

<style scoped>
.md-preview {
  min-height: 0;
  background: #f5f5f5;
}
.preview-container {
  max-width: 900px;
  margin: 0 auto;
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}
.preview-error {
  margin-bottom: 16px;
}
.preview-header {
  margin-bottom: 16px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.preview-tabs {
  margin-top: 12px;
}
.preview-source {
  margin: 0;
  padding: 16px;
  background: #f6f8fa;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.5;
  overflow: auto;
  max-height: 70vh;
  white-space: pre-wrap;
  word-break: break-all;
}
</style>
