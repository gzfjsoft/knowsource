<template>
  <div class="ai-chat-container">
    <el-row :gutter="20" style="height: 100%">
      <!-- 左侧会话列表 -->
      <el-col :span="6" class="session-sidebar">
        <el-card class="session-card">
          <template #header>
            <div class="session-header">
              <div class="header-left">
                <el-checkbox v-model="selectAll" @change="handleSelectAll" v-if="sessionList.length > 0"></el-checkbox>
                <span>会话列表</span>
              </div>
              <div class="header-actions">
                <el-button
                  v-if="selectedSessions.length > 0"
                  type="danger"
                  size="small"
                  @click="handleBatchDelete"
                >
                  <el-icon><Delete /></el-icon>
                  批量删除
                </el-button>
                <el-button type="primary" size="small" @click="handleNewSession">
                  <el-icon><Plus /></el-icon>
                  新建会话
                </el-button>
              </div>
            </div>
          </template>
          <el-checkbox-group v-model="selectedSessions" class="session-list">
            <div
              v-for="session in sessionList"
              :key="session.session"
              :class="['session-item', { active: currentSession === session.session }]"
            >
              <div class="session-checkbox">
                <el-checkbox :label="session.session"></el-checkbox>
              </div>
              <div class="session-content" @click="handleSelectSession(session.session)">
                <div class="session-title-row">
                  <div class="session-title">{{ session.title || '新会话' }}</div>
                  <div class="session-time">{{ formatTime(session.updateTime) }}</div>
                </div>
                <div v-if="session.documentTypeCode || session.tags" class="session-meta-tags">
                  <el-tag v-if="session.documentTypeCode" size="small" type="primary" style="margin-right: 4px; margin-bottom: 2px">
                    {{ getDocumentTypeName(session.documentTypeCode) }}
                  </el-tag>
                  <el-tag
                    v-for="(tag, index) in parseTags(session.tags)"
                    :key="index"
                    size="small"
                    type="info"
                    style="margin-right: 2px; margin-bottom: 2px"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </div>
              <el-button
                type="danger"
                size="small"
                text
                @click.stop="handleDeleteSession(session.session)"
                class="delete-btn"
              >
                <el-icon><Delete /></el-icon>
              </el-button>
            </div>
          </el-checkbox-group>
        </el-card>
      </el-col>

      <!-- 右侧聊天区域 -->
      <el-col :span="18" class="chat-area">
        <el-card class="chat-card">
          <div class="chat-wrapper">
          <!-- 知识库和标签信息栏 -->
          <div v-if="currentDocumentTypeName || currentTags.length > 0" class="chat-header-info">
            <div class="header-info-content">
              <span v-if="currentDocumentTypeName" class="document-type-name">
                <el-icon><Document /></el-icon>
                {{ currentDocumentTypeName }}
              </span>
              <div v-if="currentTags.length > 0" class="tags-container">
                <el-tag
                  v-for="(tag, index) in currentTags"
                  :key="index"
                  size="small"
                  type="info"
                  style="margin-right: 8px; margin-bottom: 4px"
                >
                  {{ tag }}
                </el-tag>
              </div>
            </div>
          </div>
          <div class="chat-messages" ref="messagesContainer">
            <div
              v-for="(message, index) in messages"
              :key="index"
              :class="['message-item', message.role === 'user' ? 'user-message' : 'assistant-message']"
            >
              <div class="message-avatar">
                <el-icon v-if="message.role === 'user'"><User /></el-icon>
                  <el-icon v-else><ChatDotRound /></el-icon>
              </div>
              <div class="message-content">
                <div v-if="message.thinking" class="message-thinking markdown-body">
                  <strong>思考过程：</strong>
                  <div v-html="formatMessage(message.thinking)"></div>
                </div>
                <div
                  class="message-text markdown-body"
                  v-if="message.content"
                  v-html="formatMessage(message.content)"
                ></div>
                <div v-else-if="message.role === 'assistant' && loading" class="message-text">
                  <span class="typing-indicator">.</span>
                </div>
                <div class="message-time" v-if="message.created_at">
                  {{ formatMessageTime(message.created_at) }}
                  <!-- 性能统计信息：i 图标 + 悬浮提示，紧跟在时间后面 -->
                  <span v-if="message.stats" class="message-stats">
                    <el-tooltip :content="formatStats(message.stats)" placement="top">
                      <el-icon class="message-stats-icon">
                        <InfoFilled />
                      </el-icon>
                    </el-tooltip>
                  </span>
                </div>
                <!-- 参考资料放在最后，显示所有文件名 -->
                <div v-if="getMessageFiles(message).length" class="message-references">
                  <div class="references-title">参考资料</div>
                  <ul class="references-list">
                    <li
                      v-for="(file, idx) in getMessageFiles(message)"
                      :key="idx"
                      class="reference-item"
                    >
                      <span class="reference-link" @click="handlePreviewReference(file)">
                        {{ file }}
                      </span>
                    </li>
                  </ul>
                </div>
              </div>
            </div>
          </div>

          <div class="chat-input-area">
              <!-- 上传进度条 -->
              <div v-if="uploading" class="upload-progress-row">
                <el-progress
                  :percentage="uploadProgress"
                  :stroke-width="8"
                  status="success"
                />
                <span class="upload-progress-text">正在识别文档...</span>
              </div>
              <!-- 已上传待参考的文档列表（可删除） -->
              <div v-if="pendingChatDocNames.length > 0 && !uploading" class="pending-docs-tip">
                <div class="pending-docs-header">
                  <el-icon><Document /></el-icon>
                  <span>已上传 {{ pendingChatDocNames.length }} 个文档，发送消息时将作为参考（不进行知识库检索）</span>
                </div>
                <div class="pending-docs-list">
                  <div
                    v-for="(name, idx) in pendingChatDocNames"
                    :key="idx"
                    class="pending-doc-item"
                  >
                    <span class="pending-doc-name">{{ name }}</span>
                    <el-button
                      type="danger"
                      link
                      size="small"
                      @click="handleRemoveUploadedDoc(name)"
                    >
                      <el-icon><Close /></el-icon>
                    </el-button>
                  </div>
                </div>
              </div>
              <!-- 输入框 -->
            <div class="input-box">
              <el-input
                ref="messageInputRef"
                v-model="inputMessage"
                type="textarea"
                  :rows="4"
                  placeholder="发消息或输入 / 选择技能"
                @keydown.ctrl.enter="handleSend"
                  @keydown.enter.exact.prevent="handleSend"
                  @keydown.shift.enter
                :disabled="loading || uploading"
                  class="message-input"
              />
              </div>
              
              <!-- 操作按钮栏 -->
              <div class="input-actions-bar">
                <div class="actions-left">
                  <input
                    ref="chatDocInputRef"
                    type="file"
                    accept=".txt,.docx,.pdf"
                    style="display: none"
                    @change="onChatDocFileChange"
                  />
                  <el-tooltip content="上传文档（txt/docx/pdf），发送下一条消息时将作为参考" placement="top">
                    <el-button 
                      text 
                      class="action-btn"
                      :disabled="uploading"
                      @click="handleAttachment"
                      title="上传文档"
                    >
                      <el-icon><Paperclip /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <!-- <el-button 
                    text 
                    class="action-btn"
                    :type="chatOptions.think ? 'primary' : ''"
                    @click="chatOptions.think = !chatOptions.think"
                    title="深度思考"
                  >
                    <el-icon><Connection /></el-icon>
                    <span>深度思考</span>
                  </el-button> -->
                  <!-- <el-button 
                    text 
                    class="action-btn"
                    @click="handleSkills"
                    title="技能"
                  >
                    <el-icon><Grid /></el-icon>
                    <span>技能</span>
                  </el-button> -->
                  <el-tooltip content="清空" placement="top">
                    <el-button 
                      text 
                      class="action-btn"
                      @click="handleClear"
                    >
                      <el-icon><Close /></el-icon>
                    </el-button>
                  </el-tooltip>
                  <!-- <el-button 
                    text 
                    class="action-btn"
                    @click="handleVoice"
                    title="语音输入"
                  >
                    <el-icon><Microphone /></el-icon>
                  </el-button> -->
                  <!-- 管理员工具栏 + 深度思考 -->
                  <div class="admin-tools-row" v-if="userStore.role && userStore.role !== 'user'">
                    <el-checkbox v-model="chatOptions.think" class="deep-think-checkbox">
                      深度思考
                    </el-checkbox>
                    <el-tooltip :content="isFullscreen ? '退出全屏' : '全屏'" placement="top">
                      <el-button 
                        text 
                        class="action-btn"
                        @click="toggleFullscreen"
                      >
                        <el-icon><component :is="isFullscreen ? Fold : FullScreen" /></el-icon>
                      </el-button>
                    </el-tooltip>
                    <el-tooltip :content="menuVisible ? '隐藏菜单' : '显示菜单'" placement="top">
                      <el-button 
                        text 
                        class="action-btn"
                        @click="toggleMenu"
                      >
                        <el-icon><component :is="menuVisible ? Hide : View" /></el-icon>
                      </el-button>
                    </el-tooltip>
                  </div>
                </div>
                <div class="actions-divider"></div>
                <div class="actions-right">
                  <el-tooltip content="发送 (Ctrl+Enter)" placement="top">
                    <el-button 
                      circle
                      class="send-btn"
                      @click="handleSend" 
                      :loading="loading" 
                      :disabled="!inputMessage.trim() || loading || uploading"
                    >
                      <el-icon><ArrowUp /></el-icon>
                    </el-button>
                  </el-tooltip>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 参考资料 Markdown 预览对话框 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="参考资料预览"
      width="80%"
      :close-on-click-modal="false"
    >
      <div v-loading="previewLoading" class="md-preview">
        <div v-if="previewError" class="preview-error">
          <el-alert type="error" :title="previewError" show-icon />
        </div>
        <template v-else-if="previewDoc">
          <div class="preview-header">
            <el-tag type="info">{{ previewDoc.fileName }}</el-tag>
            <el-tag v-if="previewDoc.documentCode" type="success">
              {{ getDocumentTypeName(previewDoc.documentCode) }}
            </el-tag>
            <el-tag v-if="previewDoc.tag" type="warning">{{ previewDoc.tag }}</el-tag>
          </div>
          <el-tabs v-model="previewActiveTab" class="preview-tabs">
            <el-tab-pane label="预览" name="preview">
              <div
                v-if="previewHtmlContent"
                class="preview-content markdown-body"
                v-html="previewHtmlContent"
              />
              <el-empty v-else-if="!previewLoading" description="该文档暂无 Markdown 内容" />
            </el-tab-pane>
            <el-tab-pane label="源文件" name="source">
              <pre class="preview-source">{{ previewRawContent || '（无内容）' }}</pre>
            </el-tab-pane>
          </el-tabs>
        </template>
        <el-empty v-else-if="!previewLoading && !previewError" description="该文档暂无 Markdown 内容" />
      </div>
      <template #footer>
        <el-button @click="previewDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 新建会话知识库选择对话框 -->
    <el-dialog
      v-model="documentTypeDialogVisible"
      title="选择知识库"
      width="600px"
      :close-on-click-modal="false"
      :close-on-press-escape="true"
      :show-close="false"
      @close="handleCancelNewSession"
    >
      <div class="document-type-selector">
        <div v-if="availableDocumentTypes.length === 0" class="empty-tip">
          <el-empty description="暂无可用知识库" :image-size="80" />
        </div>
        <div v-else class="document-type-selector-body">
          <!-- 可滚动的知识库卡片区域 -->
          <div class="document-type-cards-scroll">
            <div class="document-type-cards">
              <div
                v-for="docType in availableDocumentTypes"
                :key="docType.documentTypeCode"
                :class="['document-type-card', { active: newSessionDocumentCode === docType.documentTypeCode }]"
                @click="handleSelectDocumentType(docType.documentTypeCode)"
              >
                <div class="card-content">
                  <div class="card-title">{{ docType.documentTypeName }}</div>
                  <div class="card-description" v-if="docType.description">{{ docType.description }}</div>
                </div>
                <el-icon class="check-icon" v-if="newSessionDocumentCode === docType.documentTypeCode">
                  <Check />
                </el-icon>
              </div>
            </div>
          </div>
          <!-- 固定的标签选择区域（不随上方滚动） -->
          <div v-if="newSessionDocumentCode" class="tag-selector tag-selector-fixed">
            <div class="tag-selector-label">选择标签（可选）：</div>
            <div class="tags-display-container">
              <el-tag
                v-for="tag in availableTags"
                :key="tag"
                size="small"
                :type="selectedTags.includes(tag) ? 'success' : 'info'"
                class="selectable-tag"
                @click="toggleTag(tag)"
              >
                <span v-if="selectedTags.includes(tag)" class="tag-check-icon">✓</span>
                {{ tag }}
              </el-tag>
              <div v-if="availableTags.length === 0" class="no-tags-tip">该知识库暂无标签</div>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="handleCancelNewSession">取消</el-button>
        <el-button 
          @click="handleConfirmNewSessionWithoutDoc"
          type="info"
        >
          不选择知识库
        </el-button>
        <el-button 
          v-if="newSessionDocumentCode"
          @click="handleConfirmNewSession"
          type="primary"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { parseMarkdown, parseMarkdownWithAssets } from '@/utils/markdown'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { 
  Plus, 
  Delete, 
  User, 
  ChatDotRound,
  Paperclip,
  Connection,
  Grid,
  Close,
  Microphone,
  ArrowUp,
  FullScreen,
  Fold,
  Hide,
  View,
  Check,
  Document,
  InfoFilled
} from '@element-plus/icons-vue'
import {
  sessionChat,
  getHistoryList,
  getSessionDetail,
  deleteSession,
  batchDeleteSession,
  getAIGreet,
  uploadChatDocument,
  removeChatDocument
} from '@/api/ai'
import { listMyDocumentType } from '@/api/knowsource'
import { getDistinctTags, getRawDocumentsByFilename } from '@/api/knowdata'

const userStore = useUserStore()
const messagesContainer = ref(null)
const messageInputRef = ref(null)
const loading = ref(false)
const inputMessage = ref('')
const currentSession = ref('')
const sessionList = ref([])
const messages = ref([])
const showOptions = ref(false)
const isFullscreen = ref(false)
const menuVisible = ref(true)
const documentTypeDialogVisible = ref(false)
const newSessionDocumentCode = ref('')
const availableDocumentTypes = ref([])
const isDocumentTypeLocked = ref(false)
// 存储每个会话的知识库，key 为 sessionId，value 为 documentCode
const sessionDocumentTypes = ref(new Map())
// 存储每个会话的 tags，key 为 sessionId，value 为 tags 数组
const sessionTags = ref(new Map())
const availableTags = ref([])
const selectedTags = ref([])
const chatDocInputRef = ref(null)
// 已上传待发送的文档名（发下一条消息时会作为参考，发送后后端会清除）
const pendingChatDocNames = ref([])
const uploading = ref(false)
const uploadProgress = ref(0)

// 批量删除相关
const selectedSessions = ref([])
const selectAll = ref(false)

// 计算当前知识库名称
const currentDocumentTypeName = computed(() => {
  if (!chatOptions.documentCode) return ''
  const docType = availableDocumentTypes.value.find(
    item => item.documentTypeCode === chatOptions.documentCode
  )
  return docType ? docType.documentTypeName : ''
})

// 根据 documentTypeCode 获取知识库名称
const getDocumentTypeName = (documentTypeCode) => {
  if (!documentTypeCode) return ''
  const docType = availableDocumentTypes.value.find(
    item => item.documentTypeCode === documentTypeCode
  )
  return docType ? docType.documentTypeName : documentTypeCode
}

// 计算当前标签
const currentTags = computed(() => {
  return chatOptions.tags || []
})

const chatOptions = reactive({
  model: '',
  think: true,
  documentCode: '',
  categoryId: null,
  businessId: null,
  keys: '',
  tags: []
})

const defaultWelcomeMessage = '你好！我是AI咨询助手，有什么可以帮到您的？'

// 持久化 key：切换回 ai-chat 时恢复该会话；登出时清除
const AI_CHAT_RESTORE_SESSION_KEY = 'ai-chat-restore-session'

const saveRestoreSession = () => {
  const sessionId = currentSession.value || ''
  const documentCode = chatOptions.documentCode !== undefined ? chatOptions.documentCode : ''
  const tags = chatOptions.tags && Array.isArray(chatOptions.tags) ? chatOptions.tags : []
  const payload = { sessionId, documentCode, tags }
  localStorage.setItem(AI_CHAT_RESTORE_SESSION_KEY, JSON.stringify(payload))
}

const loadRestoreSession = () => {
  try {
    const raw = localStorage.getItem(AI_CHAT_RESTORE_SESSION_KEY)
    if (!raw) return null
    const data = JSON.parse(raw)
    if (!data || typeof data !== 'object') return null
    return {
      sessionId: typeof data.sessionId === 'string' ? data.sessionId : '',
      documentCode: typeof data.documentCode === 'string' ? data.documentCode : '',
      tags: Array.isArray(data.tags) ? data.tags : []
    }
  } catch {
    return null
  }
}

// 添加欢迎消息，内容从接口 api/conf/ai/name/greet 的 data.value 获取
const addWelcomeMessage = async () => {
  let content = defaultWelcomeMessage
  try {
    const res = await getAIGreet()
    if (res && (res.code === 200 || res.code === 0) && res.data && res.data.value) {
      content = res.data.value
    }
  } catch (error) {
    console.error('获取 AI 问候语失败，使用默认文案:', error)
  }

  messages.value.push({
    role: 'assistant',
    content,
    thinking: '',
    created_at: Date.now()
  })
  scrollToBottom()
}

const loadSessionList = async () => {
  // 先加载知识库列表，以便能够显示知识库名称
  if (availableDocumentTypes.value.length === 0) {
    try {
      const res = await listMyDocumentType()
      if (res.code === 200 && res.data && res.data.list) {
        availableDocumentTypes.value = (res.data.list || []).filter(item => item.isDisabled !== 1)
      }
    } catch (error) {
      console.error('加载知识库列表失败:', error)
    }
  }
  
  try {
    const res = await getHistoryList()
    if (res.code === 200 && res.sessions) {
      sessionList.value = res.sessions
      // 不再在此处自动选中第一个会话，由 onMounted 根据「恢复会话」或「首次进入」决定
    }
  } catch (error) {
    ElMessage.error('加载会话列表失败')
  }
}

const loadSessionDetail = async (sessionId) => {
  try {
    const res = await getSessionDetail({ session: sessionId })
    if (res.code === 200 && res.messages) {
      // 对历史消息也做 think 块拆分，保证展示一致
      messages.value = (res.messages || []).map((m) => {
        if (m.role === 'assistant' && typeof m.content === 'string') {
          const { content, thinking } = splitThinkBlocks(m.content)
          return { ...m, content, thinking: m.thinking || thinking }
        }
        return m
      })
      scrollToBottom()
    }
  } catch (error) {
    ElMessage.error('加载会话详情失败')
  }
}

const handleNewSession = async () => {
  // 如果当前有会话，先清理
  if (currentSession.value) {
    currentSession.value = ''
    messages.value = []
    inputMessage.value = ''
    chatOptions.documentCode = ''
    chatOptions.tags = []
    isDocumentTypeLocked.value = false
    notifySessionStatus()
  }
  
  // 重置标签选择
  selectedTags.value = []
  availableTags.value = []
  
  // 总是重新加载知识库列表，确保获取最新数据
  try {
    const res = await listMyDocumentType()
    if (res.code === 200 && res.data && res.data.list) {
      // 过滤掉已禁止的知识库（isDisabled === 1）
      availableDocumentTypes.value = (res.data.list || []).filter(item => item.isDisabled !== 1)
    } else {
      availableDocumentTypes.value = []
    }
  } catch (error) {
    console.error('加载知识库列表失败:', error)
    ElMessage.error('加载知识库列表失败')
    availableDocumentTypes.value = []
  }
  
  // 如果有知识库，弹出选择对话框（允许不选择）
  if (availableDocumentTypes.value.length > 0) {
    // 默认不选择，让用户自己选择
    newSessionDocumentCode.value = ''
    // 确保对话框显示
    documentTypeDialogVisible.value = true
    // 使用 nextTick 确保对话框已渲染
    nextTick(() => {
      // 焦点设置会在 watch 中处理
    })
  } else {
    // 如果没有知识库，直接创建新会话（不设置知识库）
    createNewSession('')
    // 添加AI助手的欢迎消息（从配置接口获取）
    addWelcomeMessage()
    // 将焦点移到输入框
    nextTick(() => {
      setTimeout(() => {
        if (messageInputRef.value) {
          // Element Plus 的 el-input 需要访问内部的 textarea 元素
          const textareaEl = messageInputRef.value.$el?.querySelector('textarea') || 
                            messageInputRef.value.textarea || 
                            messageInputRef.value.$el
          if (textareaEl && textareaEl.focus) {
            textareaEl.focus()
          }
        }
      }, 100)
    })
  }
}

const handleCancelNewSession = () => {
  documentTypeDialogVisible.value = false
  newSessionDocumentCode.value = ''
  selectedTags.value = []
  availableTags.value = []
}

const handleConfirmNewSession = () => {
  // 允许不选择知识库（空字符串）
  const selectedCode = newSessionDocumentCode.value || ''
  const tags = selectedTags.value || []
  documentTypeDialogVisible.value = false
  // 使用 nextTick 确保对话框关闭后再创建会话，这样状态更新更及时
  nextTick(() => {
    createNewSession(selectedCode, tags)
    // 添加AI助手的欢迎消息（从配置接口获取）
    addWelcomeMessage()
    // 将焦点移到输入框
    setTimeout(() => {
      if (messageInputRef.value) {
        // Element Plus 的 el-input 需要访问内部的 textarea 元素
        const textareaEl = messageInputRef.value.$el?.querySelector('textarea') || 
                          messageInputRef.value.textarea || 
                          messageInputRef.value.$el
        if (textareaEl && textareaEl.focus) {
          textareaEl.focus()
        }
      }
    }, 100)
  })
}

const handleSelectDocumentType = async (documentTypeCode) => {
  // 点击卡片选中知识库
  newSessionDocumentCode.value = documentTypeCode
  const tags = await loadTagsForDocument(documentTypeCode)
  if (!documentTypeCode) return
  const tagCount = tags?.length ?? 0
  // 0 或 1 个标签时无需再选手动点「确定」，直接进入 chat
  if (tagCount > 1) return
  const tagsForSession = tagCount === 1 ? [tags[0]] : []
  selectedTags.value = tagsForSession
  documentTypeDialogVisible.value = false
  nextTick(() => {
    createNewSession(documentTypeCode, tagsForSession)
    addWelcomeMessage()
    setTimeout(() => {
      if (messageInputRef.value) {
        const textareaEl = messageInputRef.value.$el?.querySelector('textarea') ||
                          messageInputRef.value.textarea ||
                          messageInputRef.value.$el
        if (textareaEl && textareaEl.focus) {
          textareaEl.focus()
        }
      }
    }, 100)
  })
}

const loadTagsForDocument = async (documentCode) => {
  if (!documentCode) {
    availableTags.value = []
    selectedTags.value = []
    return []
  }
  try {
    const res = await getDistinctTags({ documentCode })
    if (res.code === 200 && res.data && res.data.tags) {
      availableTags.value = res.data.tags || []
      // 默认全选所有标签
      selectedTags.value = [...availableTags.value]
      return availableTags.value
    } else {
      availableTags.value = []
      selectedTags.value = []
      return []
    }
  } catch (error) {
    console.error('加载标签列表失败:', error)
    availableTags.value = []
    selectedTags.value = []
    return []
  }
}

// 解析 tags 字符串（逗号分割）为数组
const parseTags = (tagsString) => {
  if (!tagsString || typeof tagsString !== 'string') {
    return []
  }
  return tagsString.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0)
}

const toggleTag = (tag) => {
  const index = selectedTags.value.indexOf(tag)
  if (index > -1) {
    // 如果已选中，则取消选中
    selectedTags.value.splice(index, 1)
  } else {
    // 如果未选中，则选中
    selectedTags.value.push(tag)
  }
}

const handleConfirmNewSessionWithoutDoc = () => {
  // 不选择知识库，documentCode 为空字符串
  newSessionDocumentCode.value = ''
  selectedTags.value = []
  handleConfirmNewSession()
}

const createNewSession = (documentCode, tags = []) => {
  currentSession.value = ''
  messages.value = []
  inputMessage.value = ''
  // 设置当前会话的知识库
  chatOptions.documentCode = documentCode
  chatOptions.tags = tags
  // 如果还没有会话ID，先设置为临时值，等创建会话后再更新
  if (documentCode) {
    sessionDocumentTypes.value.set('', documentCode)
    isDocumentTypeLocked.value = true
  }
  if (tags.length > 0) {
    sessionTags.value.set('', tags)
  }
  notifySessionStatus()
  saveRestoreSession()
}

const handleSelectSession = async (sessionId) => {
  currentSession.value = sessionId
  await loadSessionDetail(sessionId)
  // 从会话列表中查找该会话的知识库
  const session = sessionList.value.find(s => s.session === sessionId)
  if (session) {
    // 使用会话中的 documentTypeCode（可能为空字符串）
    const documentCode = session.documentTypeCode || ''
    chatOptions.documentCode = documentCode
    // 保存到映射中
    sessionDocumentTypes.value.set(sessionId, documentCode)
    // 如果有会话，锁定知识库（不允许修改）
    isDocumentTypeLocked.value = true
    // 解析并加载该会话的标签（从 session.tags 字符串解析）
    let tags = []
    if (session.tags) {
      tags = parseTags(session.tags)
    } else {
      // 如果没有 tags，尝试从 sessionTags Map 中获取
      const savedTags = sessionTags.value.get(sessionId)
      if (savedTags) {
        tags = savedTags
      }
    }
    chatOptions.tags = tags
    // 保存到 sessionTags Map 中
    if (tags.length > 0) {
      sessionTags.value.set(sessionId, tags)
    }
  } else {
    // 如果没有找到会话，检查是否有保存的知识库
    const savedDocumentCode = sessionDocumentTypes.value.get(sessionId)
    if (savedDocumentCode !== undefined) {
      chatOptions.documentCode = savedDocumentCode
      isDocumentTypeLocked.value = true
    } else {
      // 如果没有保存的知识库，说明是新会话，允许修改
      isDocumentTypeLocked.value = false
    }
    // 加载标签
    const savedTags = sessionTags.value.get(sessionId)
    if (savedTags) {
      chatOptions.tags = savedTags
    } else {
      chatOptions.tags = []
    }
  }
  notifySessionStatus()
  saveRestoreSession()
}

const handleDeleteSession = (sessionId) => {
  ElMessageBox.confirm('确定要删除这个会话吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await deleteSession({ session: sessionId })
      if (res.code === 200) {
        ElMessage.success('删除成功')
        // 清理该会话的知识库映射
        sessionDocumentTypes.value.delete(sessionId)
        if (currentSession.value === sessionId) {
          currentSession.value = ''
          messages.value = []
          inputMessage.value = ''
          chatOptions.documentCode = ''
          chatOptions.tags = []
          isDocumentTypeLocked.value = false
          notifySessionStatus()
        }
        loadSessionList()
      } else {
        ElMessage.error(res.msg || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

// 全选/取消全选
const handleSelectAll = (val) => {
  if (val) {
    selectedSessions.value = sessionList.value.map(session => session.session)
  } else {
    selectedSessions.value = []
  }
}

// 批量删除
const handleBatchDelete = () => {
  if (selectedSessions.value.length === 0) return
  
  ElMessageBox.confirm(`确定要删除选中的 ${selectedSessions.value.length} 个会话吗？`, '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      const res = await batchDeleteSession({ sessions: selectedSessions.value })
      if (res.code === 200) {
        ElMessage.success('批量删除成功')
        // 清理被删除会话的知识库映射
        selectedSessions.value.forEach(sessionId => {
          sessionDocumentTypes.value.delete(sessionId)
          if (currentSession.value === sessionId) {
            currentSession.value = ''
            messages.value = []
            inputMessage.value = ''
            chatOptions.documentCode = ''
            chatOptions.tags = []
            isDocumentTypeLocked.value = false
            notifySessionStatus()
          }
        })
        selectedSessions.value = []
        selectAll.value = false
        loadSessionList()
      } else {
        ElMessage.error(res.msg || '删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败，请稍后重试')
    }
  }).catch(() => {})
}

const handleSend = async () => {
  if (!inputMessage.value.trim() || loading.value) return

  const userMessage = {
    role: 'user',
    content: inputMessage.value,
    created_at: Date.now()
  }
  messages.value.push(userMessage)
  const messageText = inputMessage.value
  inputMessage.value = ''
  pendingChatDocNames.value = [] // 发送后清空待参考文档提示（后端会使用并清除缓存）
  loading.value = true
  scrollToBottom()

  // 创建助手消息占位符
  const assistantMessageIndex = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '',
    thinking: '',
        filesinfos: [],
        stats: null,
    created_at: Date.now()
  })

  try {
    await sessionChat(
      {
      message: messageText,
      session: currentSession.value || undefined,
      model: chatOptions.model || undefined,
      think: chatOptions.think,
      documentCode: chatOptions.documentCode || undefined,
      categoryId: chatOptions.categoryId || undefined,
      businessId: chatOptions.businessId || undefined,
      keys: chatOptions.keys || undefined,
      tags: chatOptions.tags && chatOptions.tags.length > 0 ? chatOptions.tags : undefined
      },
      // onChunk: 每次收到数据块时调用
      (chunk) => {
        const fromSplit = splitThinkBlocks(chunk.content || '')
        // 优先使用 API 的 reasoning 字段（thinking），否则从 content 的 <think> 标签解析
        const thinking = chunk.reasoning !== undefined ? chunk.reasoning : fromSplit.thinking
        messages.value[assistantMessageIndex].content = fromSplit.content
        messages.value[assistantMessageIndex].thinking = thinking
        if (chunk.filesinfos && Array.isArray(chunk.filesinfos)) {
          messages.value[assistantMessageIndex].filesinfos = [...new Set(chunk.filesinfos)]
        }
        // attributedFiles：若存在则用于“附件/参考资料”展示（优先级高于 filesinfos）
        if (chunk.attributedFiles && Array.isArray(chunk.attributedFiles)) {
          messages.value[assistantMessageIndex].attributedFiles = [...new Set(chunk.attributedFiles)]
        }
        if (chunk.stats) {
          messages.value[assistantMessageIndex].stats = chunk.stats
        }
        scrollToBottom()
      },
      // onComplete: 流式响应完成时调用
      (result) => {
        const fromSplit = splitThinkBlocks(result.content || '')
        const thinking = result.reasoning !== undefined ? result.reasoning : fromSplit.thinking
        messages.value[assistantMessageIndex].content = fromSplit.content
        messages.value[assistantMessageIndex].thinking = thinking
        if (result.filesinfos && Array.isArray(result.filesinfos)) {
          messages.value[assistantMessageIndex].filesinfos = [...new Set(result.filesinfos)]
        }
        if (result.attributedFiles && Array.isArray(result.attributedFiles)) {
          messages.value[assistantMessageIndex].attributedFiles = [...new Set(result.attributedFiles)]
        }
        if (result.stats) {
          messages.value[assistantMessageIndex].stats = result.stats
        }
        if (result.session) {
          const oldSession = currentSession.value
          currentSession.value = result.session
          // 如果是从临时会话创建的，将知识库和标签迁移到新会话
          if (oldSession === '' && sessionDocumentTypes.value.has('')) {
            const documentCode = sessionDocumentTypes.value.get('')
            sessionDocumentTypes.value.set(result.session, documentCode)
            sessionDocumentTypes.value.delete('')
            isDocumentTypeLocked.value = true
          }
          // 迁移标签
          if (oldSession === '' && sessionTags.value.has('')) {
            const tags = sessionTags.value.get('')
            sessionTags.value.set(result.session, tags)
            sessionTags.value.delete('')
          }
          notifySessionStatus()
          saveRestoreSession()
      }
      
      loadSessionList()
        loading.value = false
        scrollToBottom()
      },
      // onError: 发生错误时调用
      (error) => {
        ElMessage.error(error.message || '发送失败，请稍后重试')
        messages.value.pop() // 移除助手消息占位符
      messages.value.pop() // 移除用户消息
        loading.value = false
    }
    )
  } catch (error) {
    ElMessage.error('发送失败，请稍后重试')
    messages.value.pop() // 移除助手消息占位符
    messages.value.pop() // 移除用户消息
    loading.value = false
  }
}

const handleClear = () => {
  inputMessage.value = ''
}

const handleAttachment = () => {
  if (uploading.value) return
  chatDocInputRef.value?.click()
}

const onChatDocFileChange = async (e) => {
  const file = e.target?.files?.[0]
  e.target.value = ''
  if (!file) return
  const ext = (file.name || '').toLowerCase().replace(/^.*\./, '.')
  if (!['.txt', '.docx', '.pdf'].includes(ext)) {
    ElMessage.warning('仅支持 .txt、.docx、.pdf 文件')
    return
  }
  uploading.value = true
  uploadProgress.value = 0
  const formData = new FormData()
  formData.append('file', file)
  try {
    const res = await uploadChatDocument(formData, {
      onUploadProgress: (progressEvent) => {
        if (progressEvent.total) {
          uploadProgress.value = Math.round((progressEvent.loaded / progressEvent.total) * 100)
        } else {
          uploadProgress.value = 50
        }
      }
    })
    if (res && res.code === 200 && res.data) {
      pendingChatDocNames.value = [...pendingChatDocNames.value, res.data.filename]
      ElMessage.success(res.data.message || '已识别并缓存')
    } else {
      ElMessage.error(res?.message || res?.info || '上传失败')
    }
  } catch (err) {
    ElMessage.error(err?.message || '上传失败，请稍后重试')
  } finally {
    uploading.value = false
    uploadProgress.value = 0
  }
}

const handleRemoveUploadedDoc = async (filename) => {
  try {
    await removeChatDocument({ filename })
    pendingChatDocNames.value = pendingChatDocNames.value.filter(n => n !== filename)
    ElMessage.success('已移除')
  } catch (err) {
    ElMessage.error(err?.message || '移除失败')
  }
}

const handleSkills = () => {
  ElMessage.info('技能功能开发中')
}

const handleVoice = () => {
  ElMessage.info('语音输入功能开发中')
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const formatMessage = (text) => parseMarkdown(text)

const formatTime = (timestamp) => {
  if (!timestamp) return ''
  // 后端返回的是秒级时间戳，需要转换为毫秒级
  const timestampMs = timestamp < 10000000000 ? timestamp * 1000 : timestamp
  const date = new Date(timestampMs)
  const now = new Date()
  const diff = now - date
  
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  
  return date.toLocaleDateString('zh-CN')
}

const formatMessageTime = (timestamp) => {
  if (!timestamp) return ''
  // 后端返回的是秒级时间戳，需要转换为毫秒级
  const timestampMs = timestamp < 10000000000 ? timestamp * 1000 : timestamp
  const date = new Date(timestampMs)
  return date.toLocaleString('zh-CN')
}

// 将 stats 对象格式化为提示文本
const formatStats = (stats) => {
  if (!stats) return ''
  const parts = []
  if (typeof stats.fullTextMs === 'number') parts.push(`全文 ${stats.fullTextMs}ms`)
  if (typeof stats.mainSearchMs === 'number') parts.push(`主检索 ${stats.mainSearchMs}ms`)
  if (typeof stats.subSearchMs === 'number') parts.push(`子检索 ${stats.subSearchMs}ms`)
  if (typeof stats.firstTokenMs === 'number') parts.push(`首字 ${stats.firstTokenMs}ms`)
  if (typeof stats.totalStreamMs === 'number') parts.push(`总耗时 ${stats.totalStreamMs}ms`)
  if (stats.modelName) parts.push(`模型 ${stats.modelName}`)
  return parts.join('；')
}

// 将带 <think>...</think> 的内容拆分为 thinking + 正式 content
function splitThinkBlocks(raw) {
  if (!raw) return { content: '', thinking: '' }
  const regex = /<think>([\s\S]*?)<\/think>/i
  const match = raw.match(regex)
  if (!match) {
    return { content: raw, thinking: '' }
  }
  const thinking = (match[1] || '').trim()
  const content = raw.replace(regex, '').trim()
  return { content, thinking }
}

// 消息参考资料：优先用 files（历史接口），否则用 filesinfos（流式），并在前端再做一次去重
const getMessageFiles = (message) => {
  if (!message) return []
  // attributedFiles 优先：如果后端提供了归因后的附件列表，只展示这一份
  if (message.attributedFiles && Array.isArray(message.attributedFiles) && message.attributedFiles.length) {
    return Array.from(new Set(message.attributedFiles))
  }
  let files = []
  if (message.files && Array.isArray(message.files) && message.files.length) {
    files = message.files
  } else if (message.filesinfos && Array.isArray(message.filesinfos) && message.filesinfos.length) {
    files = message.filesinfos
  }
  if (!files.length) return []
  // 使用 Set 再次去重，避免同一文件在一条消息中出现多次
  return Array.from(new Set(files))
  return []
}

// 只保留在返回内容（答案或思考）中出现过的文件名，未出现的从列表中删除。
// 若内容里没有 .pdf，且与文件名（去扩展名）约 90% 一致也认为存在，不删除。
// 匹配时【】与〔〕视为相同。
const getMessageFilesFiltered = (message) => {
  const files = getMessageFiles(message)
  if (!files.length) return []
  const content = [message.content, message.thinking].filter(Boolean).join('\n')
  if (!content) return files

  const normContent = normalizeBrackets(content)
  const hasPdfInContent = content.includes('.pdf')

  return files.filter((fileName) => {
    if (content.includes(fileName)) return true
    const nameOnly = fileName.replace(/\.[^.]+$/, '')
    if (!nameOnly) return true
    if (content.includes(nameOnly)) return true
    const normName = normalizeBrackets(nameOnly)
    if (normContent.includes(normName)) return true
    // 内容里没有 .pdf 时，90% 文件名一致也认为存在
    if (!hasPdfInContent) {
      const maxLen = longestSubstringLengthInContent(normName, normContent)
      if (maxLen >= normName.length * 0.9) return true
    }
    return false
  })
}

// 【】与〔〕视为相同，统一为【】
function normalizeBrackets(s) {
  if (typeof s !== 'string') return ''
  return s.replace(/〔/g, '【').replace(/〕/g, '】')
}

// 求 fileName 中最长的一段连续子串在 content 中出现的长度（用于 90% 一致判断）
function longestSubstringLengthInContent(fileName, content) {
  let maxLen = 0
  for (let i = 0; i < fileName.length; i++) {
    for (let len = fileName.length - i; len > maxLen; len--) {
      const sub = fileName.slice(i, i + len)
      if (content.includes(sub)) {
        maxLen = len
        break
      }
    }
  }
  return maxLen
}

// ========== 参考资料 Markdown 预览（复用 MdPreview 样式和逻辑） ==========
const previewDialogVisible = ref(false)
const previewLoading = ref(false)
const previewError = ref('')
const previewDoc = ref(null)
const previewActiveTab = ref('preview')

const previewBaseUrl = computed(() => {
  const d = previewDoc.value
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
    mdBasePath = normalizedFilePath.split('/').slice(0, -1).join('/')
    if (mdDir) mdBasePath = normalizeRelPath([mdBasePath, mdDir].filter(Boolean).join('/'))
  } else {
    mdBasePath = normalizedFilePath + '.file'
    if (mdDir) mdBasePath = normalizeRelPath(`${mdBasePath}/${mdDir}`)
  }
  if (!mdBasePath) return ''

  const encodedPath = encodePathSegments(mdBasePath)
  return `${origin}/api/v1/md/${encodedPath}/`
})

const previewRawContent = computed(() => previewDoc.value?.content ?? '')

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

function resolvePreviewImageSrc(src) {
  if (/^https?:\/\//i.test(src) || src.startsWith('data:')) return src
  const base = previewBaseUrl.value
  if (!base) return src
  const norm = normalizeRelPath(src)
  if (!norm) return src
  const slash = norm.startsWith('/') ? '' : '/'
  return base.replace(/\/$/, '') + slash + norm
}

const previewHtmlContent = computed(() => {
  const md = previewRawContent.value
  if (!md) return ''
  const base = previewBaseUrl.value
  return parseMarkdownWithAssets(md, {
    resolveImageSrc: (src) => resolvePreviewImageSrc(src),
    resolveLinkHref: (href) => {
      if (!base) return href
      const norm = normalizeRelPath(href)
      if (!norm) return href
      return base.replace(/\/$/, '') + (norm.startsWith('/') ? '' : '/') + norm
    },
  })
})

const handlePreviewReference = async (fileName) => {
  if (!fileName) return
  previewDialogVisible.value = true
  previewLoading.value = true
  previewError.value = ''
  previewDoc.value = null
  previewActiveTab.value = 'preview'
  try {
    const res = await getRawDocumentsByFilename({ fileName })
    if (res.code === 200 && res.data) {
      previewDoc.value = res.data
    } else {
      previewError.value = res.message || res.info || '获取文档失败'
    }
  } catch (e) {
    previewError.value = e?.response?.data?.info || e?.message || '加载文档失败'
  } finally {
    previewLoading.value = false
  }
}

watch(messages, () => {
  scrollToBottom()
}, { deep: true })

// 监听对话框打开状态（不再需要焦点设置，因为点击卡片直接关闭）
watch(documentTypeDialogVisible, (newVal) => {
  if (newVal) {
    // 对话框打开时，重置选择状态
    newSessionDocumentCode.value = ''
  }
})

const toggleFullscreen = () => {
  if (!document.fullscreenElement) {
    document.documentElement.requestFullscreen().then(() => {
      isFullscreen.value = true
    }).catch(err => {
      console.error('进入全屏失败:', err)
    })
  } else {
    document.exitFullscreen().then(() => {
      isFullscreen.value = false
    }).catch(err => {
      console.error('退出全屏失败:', err)
    })
  }
}

const toggleMenu = () => {
  menuVisible.value = !menuVisible.value
  // 通过事件通知 MainLayout 切换菜单显示
  window.dispatchEvent(new CustomEvent('toggle-menu', { detail: { visible: menuVisible.value } }))
}

// 监听知识库变化事件（从 MainLayout 触发）
const handleDocumentTypeChanged = (event) => {
  // 只有在没有锁定知识库的情况下才允许修改
  if (!isDocumentTypeLocked.value) {
    chatOptions.documentCode = event.detail.documentCode || ''
  }
}

// 通知 MainLayout 会话状态变化
const notifySessionStatus = () => {
  // 即使没有会话ID，如果有知识库，也应该显示
  const hasSession = !!currentSession.value || chatOptions.documentCode !== undefined
  // 检查是否有锁定的知识库（包括临时会话）
  const isLocked = (currentSession.value && sessionDocumentTypes.value.has(currentSession.value)) || 
                   (!currentSession.value && sessionDocumentTypes.value.has(''))
  // 明确传递 documentCode，包括空字符串的情况
  const documentCode = chatOptions.documentCode !== undefined ? chatOptions.documentCode : ''
  window.dispatchEvent(new CustomEvent('session-status-changed', {
    detail: {
      hasSession,
      isLocked,
      documentCode: documentCode
    }
  }))
}

// 监听全屏状态变化
const handleFullscreenChange = () => {
  isFullscreen.value = !!document.fullscreenElement
}

onMounted(async () => {
  window.addEventListener('document-type-changed', handleDocumentTypeChanged)
  document.addEventListener('fullscreenchange', handleFullscreenChange)

  await loadSessionList()
  const restored = loadRestoreSession()

  if (restored && restored.sessionId && sessionList.value.some(s => s.session === restored.sessionId)) {
    // 有保存的会话且在列表中：恢复该会话，不弹知识库选择
    await handleSelectSession(restored.sessionId)
    return
  }

  if (restored && (restored.documentCode !== undefined || restored.sessionId === '')) {
    // 有保存的「新会话」状态（已选知识库但可能尚未发消息）：恢复知识库与标签，不弹窗
    chatOptions.documentCode = restored.documentCode || ''
    chatOptions.tags = restored.tags || []
    if (restored.documentCode) {
      sessionDocumentTypes.value.set('', restored.documentCode)
      isDocumentTypeLocked.value = true
    }
    if ((restored.tags || []).length > 0) {
      sessionTags.value.set('', restored.tags)
    }
    addWelcomeMessage()
    notifySessionStatus()
    saveRestoreSession()
    return
  }

  // 首次进入或登出后：立刻弹出知识库选择
  if (availableDocumentTypes.value.length === 0) {
    try {
      const res = await listMyDocumentType()
      if (res.code === 200 && res.data && res.data.list) {
        availableDocumentTypes.value = (res.data.list || []).filter(item => item.isDisabled !== 1)
      }
    } catch (e) {
      console.error('加载知识库列表失败:', e)
    }
  }
  newSessionDocumentCode.value = ''
  selectedTags.value = []
  availableTags.value = []
  if (availableDocumentTypes.value.length > 0) {
    documentTypeDialogVisible.value = true
  } else {
    createNewSession('')
    addWelcomeMessage()
  }
})

onUnmounted(() => {
  window.removeEventListener('document-type-changed', handleDocumentTypeChanged)
  document.removeEventListener('fullscreenchange', handleFullscreenChange)
})
</script>

<style scoped>
.ai-chat-container {
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.ai-chat-container :deep(.el-row) {
  height: 100%;
  margin: 0 !important;
}

.ai-chat-container :deep(.el-row > .el-col) {
  height: 100%;
}

.session-sidebar {
  height: 100%;
}

.session-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.session-card :deep(.el-card__body) {
  padding-top: 8px;
}

.session-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 8px;
}



.session-item {
  display: flex;
  align-items: flex-start;
  padding: 6px 8px;
  margin-bottom: 4px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  cursor: pointer;
  position: relative;
  transition: all 0.3s;
  line-height: 1.3;
}

.session-checkbox {
  margin-right: 8px;
  flex-shrink: 0;
  margin-top: 2px;
}

.session-checkbox :deep(.el-checkbox__label) {
  display: none;
}

.session-content {
  flex: 1;
  min-width: 0;
}

.session-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}

.session-title {
  font-weight: 500;
  color: #303133;
  font-size: 13px;
  line-height: 1.3;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.session-document-code {
  margin-bottom: 2px;
}

.session-meta-tags {
  margin-bottom: 2px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px;
}

.session-preview {
  font-size: 11px;
  color: #909399;
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
}

.session-time {
  font-size: 11px;
  color: #c0c4cc;
  line-height: 1.2;
  flex-shrink: 0;
  white-space: nowrap;
}

.delete-btn {
  position: absolute;
  bottom: 4px;
  right: 4px;
  opacity: 0;
  transition: opacity 0.3s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.session-list {
  flex: 1;
  overflow-y: auto;
  margin-top: 10px;
}

.session-item {
  padding: 6px 8px;
  margin-bottom: 4px;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  cursor: pointer;
  position: relative;
  transition: all 0.3s;
  line-height: 1.3;
}

.session-item:hover {
  background-color: #f5f7fa;
  border-color: #409eff;
}

.session-item.active {
  background-color: #ecf5ff;
  border-color: #409eff;
}

.session-title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2px;
}

.session-title {
  font-weight: 500;
  color: #303133;
  font-size: 13px;
  line-height: 1.3;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 8px;
}

.session-document-code {
  margin-bottom: 2px;
}

.session-meta-tags {
  margin-bottom: 2px;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px;
}

.session-preview {
  font-size: 11px;
  color: #909399;
  margin-bottom: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  line-height: 1.2;
}

.session-time {
  font-size: 11px;
  color: #c0c4cc;
  line-height: 1.2;
  flex-shrink: 0;
  white-space: nowrap;
}

.delete-btn {
  position: absolute;
  bottom: 4px;
  right: 4px;
  opacity: 0;
  transition: opacity 0.3s;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.chat-area {
  height: 100%;
}

.chat-card {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.chat-card :deep(.el-card__body) {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 0;
  overflow: hidden;
}

.chat-wrapper {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-header-info {
  flex-shrink: 0;
  padding: 12px 20px;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.header-info-content {
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.document-type-name {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.document-type-name .el-icon {
  color: #409eff;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.chat-messages {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 0;
}

.message-item {
  display: flex;
  margin-bottom: 20px;
  animation: fadeIn 0.3s;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.user-message {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.user-message .message-avatar {
  background-color: #409eff;
  color: #fff;
  margin-left: 12px;
}

.assistant-message .message-avatar {
  background-color: #67c23a;
  color: #fff;
  margin-right: 12px;
}

.message-content {
  max-width: 70%;
  background-color: #fff;
  padding: 16px 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.user-message .message-content {
  background-color: #409eff;
  color: #fff;
}

.message-text {
  word-wrap: break-word;
}

/* 消息区 Markdown 细节由全局 markdown-body.css 提供 */

.message-thinking {
  margin-bottom: 12px;
  padding-bottom: 1.5em;
  /* border-top: 1px solid #e4e7ed; */
  font-size: 12px;
  background-color: #f5f7fa;
  color: #909399;
}

.user-message .message-thinking {
  border-top-color: rgba(255, 255, 255, 0.3);
  color: rgba(255, 255, 255, 0.8);
}

.message-references {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #606266;
}

.references-title {
  font-weight: 600;
  margin-bottom: 6px;
  color: #909399;
}

.references-list {
  margin: 0;
  padding-left: 18px;
  line-height: 1.6;
}

.references-list li {
  margin-bottom: 2px;
  word-break: break-all;
}

.reference-item {
  font-size: 12px;
}
.reference-link {
  color: #409eff;
  cursor: pointer;
}
.reference-link:hover {
  color: #66b1ff;
  text-decoration: underline;
}

.user-message .message-references {
  border-top-color: rgba(255, 255, 255, 0.3);
  color: rgba(255, 255, 255, 0.9);
}

.user-message .references-title {
  color: rgba(255, 255, 255, 0.85);
}

.message-time {
  margin-top: 8px;
  font-size: 12px;
  color: #c0c4cc;
  display: inline-flex;
  align-items: center;
  line-height: 1;
}

.message-stats {
  margin-left: 4px;
  font-size: 12px;
  color: inherit;
  display: inline-flex;
  align-items: center;
}

.message-stats-icon {
  cursor: pointer;
  font-size: 14px;
}

/* 参考资料预览复用 MdPreview 简化样式 */
.md-preview {
  min-height: 200px;
}
.preview-header {
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.preview-tabs {
  margin-top: 8px;
}
.preview-content.markdown-body {
  padding: 8px 0;
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
}

.user-message .message-time {
  color: rgba(255, 255, 255, 0.7);
}

.typing-indicator {
  display: inline-block;
}

.typing-indicator::after {
  content: '...';
  animation: dots 1.5s steps(4, end) infinite;
}

@keyframes dots {
  0%, 20% {
    content: '.';
  }
  40% {
    content: '..';
  }
  60%, 100% {
    content: '...';
  }
}

.chat-input-area {
  flex-shrink: 0;
  border-top: 1px solid #e4e7ed;
  padding: 16px;
  background-color: #fff;
}

.pending-docs-tip {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 10px;
  font-size: 13px;
  color: #409eff;
  background-color: #ecf5ff;
  border-radius: 6px;
}

.pending-docs-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pending-docs-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pending-doc-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #d9ecff;
}

.pending-doc-name {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 12px;
  color: #303133;
}

.upload-progress-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  margin-bottom: 10px;
  background-color: #f0f9ff;
  border-radius: 6px;
}

.upload-progress-row .el-progress {
  flex: 1;
}

.upload-progress-text {
  font-size: 13px;
  color: #409eff;
  white-space: nowrap;
}

.chat-options-bar {
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid #e4e7ed;
}

.chat-options {
  margin: 0;
}

.input-box {
  margin-bottom: 12px;
}

.message-input :deep(.el-textarea__inner) {
  border-radius: 8px;
  border: 1px solid #dcdfe6;
  font-size: 14px;
  line-height: 1.5;
  resize: none;
}

.message-input :deep(.el-textarea__inner):focus {
  border-color: #409eff;
}

.input-actions-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.actions-left {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1;
}

.admin-tools-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}
.deep-think-checkbox :deep(.el-checkbox__label) {
  font-size: 12px;
}

.action-btn {
  padding: 6px 12px;
  border-radius: 6px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 14px;
  color: #606266;
  transition: all 0.2s;
}

.action-btn:hover {
  background-color: #f5f7fa;
  color: #409eff;
}

.action-btn .el-icon {
  font-size: 16px;
}

.action-btn span {
  margin-left: 4px;
}

.actions-divider {
  width: 1px;
  height: 24px;
  background-color: #e4e7ed;
  flex-shrink: 0;
}

.actions-right {
  display: flex;
  align-items: center;
}

.send-btn {
  width: 36px;
  height: 36px;
  padding: 0;
  border-radius: 50%;
  background-color: #f5f7fa;
  border: 1px solid #dcdfe6;
  color: #606266;
  transition: all 0.2s;
}

.send-btn:hover:not(:disabled) {
  background-color: #409eff;
  border-color: #409eff;
  color: #fff;
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.send-btn .el-icon {
  font-size: 18px;
}

/* 知识库选择器样式 */
.document-type-selector {
  padding: 10px 0;
}

.document-type-selector-body {
  display: flex;
  flex-direction: column;
  max-height: 420px;
}

/* 仅知识库卡片区域可滚动 */
.document-type-cards-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding-right: 4px;
}

.empty-tip {
  padding: 40px 0;
  text-align: center;
}

.document-type-cards {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0;
  padding-top: 5px;
  justify-content: center;
  align-items: stretch;
}

/* 固定的标签选择区域 */
.tag-selector-fixed {
  flex-shrink: 0;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e4e7ed;
}
.tag-selector-fixed .tag-selector-label {
  margin-bottom: 12px;
  font-size: 14px;
  color: #606266;
}
.tag-selector-fixed .no-tags-tip {
  color: #909399;
  font-size: 12px;
  min-height: 24px; /* 与 el-tag size="small" 同高 */
  display: inline-flex;
  align-items: center;
  padding: 0;
}

.document-type-card {
  position: relative;
  padding: 10px 12px;
  border: 2px solid #e4e7ed;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.3s;
  background-color: #fff;
  min-height: 72px;
  width: 130px;
  flex: 0 0 auto;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.document-type-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
  transform: translateY(-2px);
}

.document-type-card.active {
  border-color: #409eff;
  background-color: #ecf5ff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
}

.card-content {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 13px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 2px;
  word-break: break-word;
}

.card-description {
  font-size: 11px;
  color: #606266;
  line-height: 1.35;
  margin-top: 4px;
  word-break: break-word;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

.check-icon {
  color: #409eff;
  font-size: 16px;
  flex-shrink: 0;
  margin-left: 4px;
  opacity: 0;
  transition: opacity 0.3s;
}

.document-type-card:hover .check-icon {
  opacity: 1;
}

/* 滚动条样式（仅卡片区域） */
.document-type-cards-scroll::-webkit-scrollbar {
  width: 6px;
}

.document-type-cards-scroll::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.document-type-cards-scroll::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.document-type-cards-scroll::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

/* 标签选择器样式（选择知识库对话框内） */
.tags-display-container {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  min-height: 24px; /* 与 el-tag small / no-tags-tip 一致，保证一行高度统一 */
}

.selectable-tag {
  cursor: pointer;
  user-select: none;
  transition: all 0.3s;
}

.selectable-tag:hover {
  opacity: 0.8;
  transform: scale(1.05);
}

.tag-check-icon {
  margin-right: 4px;
  font-weight: bold;
  font-size: 14px;
}

.no-tags-tip {
  min-height: 24px;
  display: inline-flex;
  align-items: center;
  padding: 0;
}
</style>

