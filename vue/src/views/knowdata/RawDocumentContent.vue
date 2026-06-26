<template>
  <div class="raw-document-content">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-button type="primary" @click="handleBack">
            <el-icon><ArrowLeft /></el-icon>
            返回
          </el-button>
          <span class="title">文档内容</span>
        </div>
      </template>

      <div v-loading="loading" class="content-container">
        <div v-if="doc" class="document-info">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="文件名">
              {{ doc.fileName }}
            </el-descriptions-item>
            <el-descriptions-item label="文件大小">
              {{ formatFileSize(doc.fileSize) }}
            </el-descriptions-item>
            <el-descriptions-item label="MD5">
              {{ doc.fileMd5 }}
            </el-descriptions-item>
            <el-descriptions-item label="创建时间">
              {{ formatTime(doc.createdAt) }}
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">
              {{ formatTime(doc.updatedAt) }}
            </el-descriptions-item>
          </el-descriptions>
        </div>

        <div v-if="doc && doc.content" class="content-section">
          <div class="content-header">
            <h3>文档内容</h3>
            <div class="content-actions">
              <el-button
                v-if="canEdit"
                :type="isEditing ? 'success' : 'primary'"
                size="small"
                @click="handleEditToggle"
              >
                <el-icon><Edit v-if="!isEditing" /><Check v-else /></el-icon>
                {{ isEditing ? '保存' : '编辑' }}
              </el-button>
              <el-button
                v-if="isEditing"
                size="small"
                @click="handleCancelEdit"
              >
                <el-icon><Close /></el-icon>
                取消
              </el-button>
              <el-button type="primary" size="small" @click="handleCopy">
                <el-icon><DocumentCopy /></el-icon>
                复制
              </el-button>
              <el-button type="info" size="small" @click="handleCompare">
                <el-icon><Files /></el-icon>
                比较
              </el-button>
              <!-- <el-button
                v-if="canEdit"
                type="warning"
                size="small"
                :loading="normalizeLoading"
                @click="handleMarkdownNormalizePreview"
              >
                <el-icon><MagicStick /></el-icon>
                LLM 规范化
              </el-button> -->
            </div>
          </div>
          <div class="content-body">
            <el-input
              v-if="isEditing"
              v-model="editedContent"
              type="textarea"
              :rows="20"
              placeholder="请输入文档内容"
              class="content-editor"
            />
            <pre v-else class="content-text" v-html="highlightedContentHtml"></pre>
          </div>
        </div>

        <el-empty v-else-if="!loading && doc && !doc.content" description="该文档暂无内容" />
      </div>
    </el-card>

    <!-- 比较对话框 -->
    <el-dialog
      v-model="diffDialogVisible"
      title="内容比较"
      width="95%"
      :close-on-click-modal="false"
      class="diff-dialog"
    >
      <div v-if="diffData && diffLines.length > 0" class="diff-container">
        <div class="diff-header">
          <div class="diff-label original">
            <span>原始内容 (content_org)</span>
          </div>
          <div class="diff-label current">
            <span>当前内容 (content)</span>
          </div>
        </div>
        <div class="diff-content-wrapper">
          <div class="diff-content">
            <table class="diff-table">
              <tbody>
                <tr
                  v-for="(line, index) in diffLines"
                  :key="index"
                  :class="getDiffLineClass(line)"
                >
                  <td class="diff-line-number original" :class="{ 'empty': !line.oldLine }">
                    {{ line.oldLine || '' }}
                  </td>
                  <td class="diff-line-content original">
                    <span v-if="line.oldContent">{{ line.oldContent }}</span>
                    <span v-else class="empty-line">&nbsp;</span>
                  </td>
                  <td class="diff-line-number current" :class="{ 'empty': !line.newLine }">
                    {{ line.newLine || '' }}
                  </td>
                  <td class="diff-line-content current">
                    <span v-if="line.newContent">{{ line.newContent }}</span>
                    <span v-else class="empty-line">&nbsp;</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
      <div v-else-if="diffData && diffLines.length === 0" class="diff-no-changes">
        <el-icon><Check /></el-icon>
        <span>内容相同，没有差异</span>
      </div>
      <div v-else class="diff-loading">
        <el-icon class="is-loading"><Loading /></el-icon>
        <span>加载中...</span>
      </div>
      <template #footer>
        <el-button @click="diffDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- LLM Markdown 规范化预览 -->
    <el-dialog
      v-model="normalizeDialogVisible"
      title="LLM Markdown 规范化"
      width="92%"
      :close-on-click-modal="false"
      class="normalize-md-dialog"
      destroy-on-close
    >
      <p class="normalize-hint">
        左侧为参与规范化的原文，右侧为模型输出。确认无误后可保存到知识库或仅应用到编辑区继续修改。
      </p>
      <el-row :gutter="16" class="normalize-panels">
        <el-col :span="12">
          <div class="normalize-panel-title">原文</div>
          <el-input
            v-model="normalizeOriginal"
            type="textarea"
            :rows="22"
            readonly
            class="normalize-textarea"
          />
        </el-col>
        <el-col :span="12">
          <div class="normalize-panel-title">规范化后</div>
          <el-input
            v-model="normalizeFormatted"
            type="textarea"
            :rows="22"
            class="normalize-textarea"
          />
        </el-col>
      </el-row>
      <template #footer>
        <el-button @click="normalizeDialogVisible = false">关闭</el-button>
        <el-button type="primary" plain @click="handleNormalizeApplyEditor">
          应用到编辑区
        </el-button>
        <el-button type="success" :loading="normalizeSaveLoading" @click="handleNormalizeApplySave">
          保存到知识库
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, DocumentCopy, Edit, Check, Close, Files, Loading, MagicStick } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getRawDocuments, updateRawDocumentsContent, getRawDocumentsContentDiff, previewRawDocumentsMarkdownNormalize } from '@/api/knowdata'
import { diffLines as diffLinesFn } from 'diff'
import { navigateBackToRawDocumentsList } from '@/utils/rawDocumentsListNavigation'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const hasPermission = (p) => (userStore.empPermissions || []).includes(p)
const loading = ref(false)
const doc = ref(null)
const isEditing = ref(false)
const editedContent = ref('')
const diffDialogVisible = ref(false)
const diffData = ref(null)
const diffLines = ref([])

const normalizeDialogVisible = ref(false)
const normalizeLoading = ref(false)
const normalizeSaveLoading = ref(false)
const normalizeOriginal = ref('')
const normalizeFormatted = ref('')

const keyword = computed(() => String(route.query.q || '').trim())

const formatTime = (timestamp) => {
  if (!timestamp) return '-'
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN')
}

const formatFileSize = (bytes) => {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
  return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
}

// 计算是否可以编辑：有「功能-更新文档内容」权限 且 非审核状态 且 isToMd = 1
const canEdit = computed(() => {
  return hasPermission('功能-更新文档内容') && doc.value && doc.value.isAudit !== 1 && doc.value.isToMd === 1
})

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

const highlightedContentHtml = computed(() => {
  const content = doc.value?.content || ''
  const kw = keyword.value
  const safe = escapeHtml(content)
  if (!kw) return safe
  const re = new RegExp(escapeRegExp(kw), 'gi')
  return safe.replace(re, (m) => `<mark class="hit-highlight">${m}</mark>`)
})

const navigateBackFromContent = () => navigateBackToRawDocumentsList(router, route)

const loadDocument = async () => {
  const id = route.params.id
  if (!id) {
    ElMessage.error('文档ID不能为空')
    navigateBackFromContent()
    return
  }

  loading.value = true
  try {
    const res = await getRawDocuments({ id: parseInt(id) })
    if (res.code === 200 && res.data) {
      doc.value = res.data
      editedContent.value = res.data.content || ''
      
      // 如果路由查询参数中有 edit=true 且文档可编辑，则直接进入编辑模式
      if (route.query.edit === 'true' && res.data.isAudit !== 1 && res.data.isToMd === 1) {
        isEditing.value = true
      }

      // 自动滚动到第一个命中
      if (keyword.value) {
        await nextTick()
        const firstHit = window.document.querySelector('.content-body mark.hit-highlight')
        if (firstHit && typeof firstHit.scrollIntoView === 'function') {
          firstHit.scrollIntoView({ block: 'center' })
        }
      }
    } else {
      ElMessage.error(res.message || '获取文档失败')
      navigateBackFromContent()
    }
  } catch (error) {
    ElMessage.error('获取文档失败，请稍后重试')
    navigateBackFromContent()
  } finally {
    loading.value = false
  }
}

const handleBack = () => {
  navigateBackFromContent()
}

const handleEditToggle = () => {
  if (isEditing.value) {
    handleSave()
  } else {
    isEditing.value = true
    editedContent.value = doc.value.content || ''
  }
}

const handleCancelEdit = () => {
  isEditing.value = false
  editedContent.value = doc.value.content || ''
}

const handleSave = async () => {
  if (!doc.value) {
    return
  }

  try {
    const res = await updateRawDocumentsContent({
      id: doc.value.id,
      content: editedContent.value
    })
    
    if (res.code === 200) {
      ElMessage.success('保存成功')
      isEditing.value = false
      // 重新加载文档以获取最新数据
      await loadDocument()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (error) {
    ElMessage.error('保存失败，请稍后重试')
  }
}

const handleCopy = async () => {
  const contentToCopy = isEditing.value ? editedContent.value : (doc.value?.content || '')
  if (!contentToCopy) {
    ElMessage.warning('没有可复制的内容')
    return
  }

  try {
    await navigator.clipboard.writeText(contentToCopy)
    ElMessage.success('复制成功')
  } catch (error) {
    // 降级方案
    const textarea = window.document.createElement('textarea')
    textarea.value = contentToCopy
    window.document.body.appendChild(textarea)
    textarea.select()
    try {
      window.document.execCommand('copy')
      ElMessage.success('复制成功')
    } catch (err) {
      ElMessage.error('复制失败')
    }
    window.document.body.removeChild(textarea)
  }
}

const processDiff = (contentOrg, content) => {
  const org = contentOrg || ''
  const cur = content || ''
  
  // 使用 diffLinesFn 计算差异
  const changes = diffLinesFn(org, cur)
  const lines = []
  let oldLineNum = 0
  let newLineNum = 0

  changes.forEach((change) => {
    const changeLines = change.value.split('\n')
    // 移除最后一个空行（如果存在）
    if (changeLines.length > 0 && changeLines[changeLines.length - 1] === '') {
      changeLines.pop()
    }

    if (change.added) {
      // 新增的行
      changeLines.forEach((line) => {
        newLineNum++
        lines.push({
          oldLine: null,
          newLine: newLineNum,
          oldContent: null,
          newContent: line,
          type: 'added'
        })
      })
    } else if (change.removed) {
      // 删除的行
      changeLines.forEach((line) => {
        oldLineNum++
        lines.push({
          oldLine: oldLineNum,
          newLine: null,
          oldContent: line,
          newContent: null,
          type: 'removed'
        })
      })
    } else {
      // 未改变的行
      changeLines.forEach((line) => {
        oldLineNum++
        newLineNum++
        lines.push({
          oldLine: oldLineNum,
          newLine: newLineNum,
          oldContent: line,
          newContent: line,
          type: 'normal'
        })
      })
    }
  })

  return lines
}

const getDiffLineClass = (line) => {
  return {
    'diff-line-removed': line.type === 'removed',
    'diff-line-added': line.type === 'added',
    'diff-line-normal': line.type === 'normal'
  }
}

const handleMarkdownNormalizePreview = async () => {
  if (!doc.value) {
    return
  }
  const payload = { id: doc.value.id }
  if (isEditing.value) {
    payload.content = editedContent.value
  }
  normalizeLoading.value = true
  try {
    const res = await previewRawDocumentsMarkdownNormalize(payload)
    if (res.code === 200 && res.data) {
      normalizeOriginal.value = res.data.originalContent ?? ''
      normalizeFormatted.value = res.data.formattedContent ?? ''
      normalizeDialogVisible.value = true
    } else {
      ElMessage.error(res.message || 'LLM 规范化失败')
    }
  } catch (e) {
    ElMessage.error('请求失败，请稍后重试')
  } finally {
    normalizeLoading.value = false
  }
}

const handleNormalizeApplyEditor = () => {
  editedContent.value = normalizeFormatted.value
  isEditing.value = true
  normalizeDialogVisible.value = false
  ElMessage.success('已应用到编辑区，可继续修改后点「保存」')
}

const handleNormalizeApplySave = async () => {
  if (!doc.value) {
    return
  }
  normalizeSaveLoading.value = true
  try {
    const res = await updateRawDocumentsContent({
      id: doc.value.id,
      content: normalizeFormatted.value
    })
    if (res.code === 200) {
      ElMessage.success('已保存规范化内容')
      normalizeDialogVisible.value = false
      isEditing.value = false
      await loadDocument()
    } else {
      ElMessage.error(res.message || '保存失败')
    }
  } catch (e) {
    ElMessage.error('保存失败，请稍后重试')
  } finally {
    normalizeSaveLoading.value = false
  }
}

const handleCompare = async () => {
  if (!doc.value) {
    return
  }

  diffDialogVisible.value = true
  diffData.value = null
  diffLines.value = []

  try {
    const res = await getRawDocumentsContentDiff({ id: doc.value.id })
    if (res.code === 200 && res.data) {
      diffData.value = res.data
      diffLines.value = processDiff(res.data.contentOrg, res.data.content)
    } else {
      ElMessage.error(res.message || '获取内容差异失败')
      diffDialogVisible.value = false
    }
  } catch (error) {
    ElMessage.error('获取内容差异失败，请稍后重试')
    diffDialogVisible.value = false
  }
}

onMounted(() => {
  loadDocument()
})
</script>

<style scoped>
.card-header {
  display: flex;
  align-items: center;
  gap: 20px;
  font-size: 18px;
  font-weight: 500;
}

.title {
  flex: 1;
}

.content-container {
  min-height: 400px;
}

.document-info {
  margin-bottom: 20px;
}

.content-section {
  margin-top: 20px;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e4e7ed;
}

.content-actions {
  display: flex;
  gap: 8px;
}

.content-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 500;
}

.content-body {
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  padding: 20px;
  max-height: 600px;
  overflow-y: auto;
}

.content-text {
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.6;
  color: #303133;
}

.content-text :deep(mark.hit-highlight) {
  background: #fff3bf;
  color: #1f2328;
  padding: 0 2px;
  border-radius: 2px;
}

.content-editor {
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
}

.normalize-hint {
  margin: 0 0 12px;
  font-size: 13px;
  color: #606266;
  line-height: 1.5;
}

.normalize-panel-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: #303133;
}

.normalize-textarea :deep(textarea) {
  font-family: 'Courier New', Courier, monospace;
  font-size: 13px;
  line-height: 1.5;
}

.diff-container {
  display: flex;
  flex-direction: column;
  height: 75vh;
  border: 1px solid #d0d7de;
  border-radius: 6px;
  overflow: hidden;
  background-color: #ffffff;
}

.diff-header {
  display: flex;
  background-color: #f6f8fa;
  border-bottom: 1px solid #d0d7de;
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 600;
}

.diff-label {
  flex: 1;
  color: #656d76;
}

.diff-label.original {
  border-right: 1px solid #d0d7de;
  padding-right: 16px;
}

.diff-label.current {
  padding-left: 16px;
}

.diff-content-wrapper {
  flex: 1;
  overflow: auto;
  background-color: #ffffff;
}

.diff-content {
  width: 100%;
}

.diff-table {
  width: 100%;
  border-collapse: collapse;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, 'Liberation Mono', monospace;
  font-size: 12px;
  line-height: 1.45;
}

.diff-table tbody tr {
  border-top: 1px solid transparent;
}

.diff-table tbody tr:hover {
  background-color: #f6f8fa;
}

.diff-line-number {
  width: 1%;
  min-width: 50px;
  padding: 0 10px;
  text-align: right;
  color: #656d76;
  background-color: #f6f8fa;
  border-right: 1px solid #d0d7de;
  user-select: none;
  font-variant-numeric: tabular-nums;
}

.diff-line-number.empty {
  background-color: #f6f8fa;
}

.diff-line-number.original {
  border-right: 1px solid #d0d7de;
}

.diff-line-number.current {
  border-right: 1px solid #d0d7de;
}

.diff-line-content {
  padding: 0 10px;
  white-space: pre-wrap;
  word-wrap: break-word;
  word-break: break-word;
  overflow-x: hidden;
  color: #24292f;
}

.diff-line-content.original {
  border-right: 1px solid #d0d7de;
}

.diff-line-removed {
  background-color: #fff1f2;
}

.diff-line-removed .diff-line-number.original {
  background-color: #ffebe9;
  color: #82071e;
}

.diff-line-removed .diff-line-content.original {
  background-color: #ffebe9;
  color: #82071e;
}

.diff-line-removed .diff-line-content.original {
  position: relative;
  padding-left: 20px;
}

.diff-line-removed .diff-line-content.original::before {
  content: '-';
  position: absolute;
  left: 10px;
  color: #cf222e;
}

.diff-line-added {
  background-color: #f0fff4;
}

.diff-line-added .diff-line-number.current {
  background-color: #ccfdf4;
  color: #116329;
}

.diff-line-added .diff-line-content.current {
  background-color: #ccfdf4;
  color: #116329;
  position: relative;
  padding-left: 20px;
}

.diff-line-added .diff-line-content.current::before {
  content: '+';
  position: absolute;
  left: 10px;
  color: #1a7f37;
}

.diff-line-normal {
  background-color: #ffffff;
}

.diff-line-normal .diff-line-content {
  color: #24292f;
}

.empty-line {
  display: inline-block;
  width: 100%;
}

.diff-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  gap: 10px;
  color: #909399;
}

.diff-no-changes {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
  gap: 10px;
  color: #67c23a;
  font-size: 14px;
}

.diff-dialog :deep(.el-dialog__body) {
  padding: 0;
}
</style>

