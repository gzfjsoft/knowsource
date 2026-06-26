<template>
  <div class="dify-chat-container">
    <div class="chat-header-bar">
      <el-button type="primary" @click="handleOpenAgentSelect">
        选择智能体
      </el-button>
    </div>
    <el-card class="chat-card">
      <div class="chat-wrapper">
        <!-- Dify 配置选择对话框 -->
        <el-dialog
          v-model="configDialogVisible"
          title="选择智能体配置"
          width="600px"
          :close-on-click-modal="false"
          :close-on-press-escape="false"
          :show-close="false"
        >
          <div v-if="loadingConfigs" class="loading-container">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>加载配置中...</span>
          </div>
          <div v-else-if="difyOptions.length === 0" class="empty-tip">
            <el-empty description="暂无可用配置" :image-size="80" />
          </div>
          <div v-else class="config-selector">
            <div class="config-cards">
              <div
                v-for="option in difyOptions"
                :key="option.name"
                :class="['config-card', { active: selectedConfig?.name === option.name }]"
                @click="handleSelectConfig(option)"
              >
                <div class="card-content">
                  <div class="card-title">{{ option.name }}</div>
                  <div v-if="option.description" class="card-description" :title="option.description">
                    {{ option.description }}
                  </div>
                  <!-- <div class="card-url" :title="option.url">
                    <el-icon><Link /></el-icon>
                    {{ option.url }}
                  </div> -->
                </div>
                <el-icon class="check-icon" v-if="selectedConfig?.name === option.name">
                  <Check />
                </el-icon>
              </div>
            </div>
          </div>
          <template #footer>
            <el-button @click="handleCancelConfig">取消</el-button>
            <el-button 
              type="primary" 
              @click="handleConfirmConfig"
              :disabled="!selectedConfig"
            >
              确定
            </el-button>
          </template>
        </el-dialog>
        
        <!-- 当前配置显示在对话区域顶部 -->
        <div v-if="currentConfig" class="chat-header-config">
          <div class="config-info-bar">
            <el-tag type="info" size="large" class="config-name-tag">
              
              <span class="config-name-text">{{ currentConfig.name }}</span>
            </el-tag>
            <el-button 
              text 
              size="small"
              @click="handleChangeConfig"
              style="margin-left: 8px"
            >
              切换配置
            </el-button>
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
              <div
                class="message-text markdown-body"
                v-if="message.content"
                v-html="formatMessage(message.content)"
              ></div>
              <div v-else-if="message.role === 'assistant' && loading" class="message-text">
                <span class="typing-indicator">.</span>
              </div>
              <!-- 显示 message_end 的所有 document_name -->
              <div
                v-if="message.retriever_resources && message.retriever_resources.length > 0"
                class="message-resources"
              >
                <strong>附件文件名：</strong>
                <ul>
                  <li v-for="(resource, idx) in message.retriever_resources" :key="idx">
                    {{ resource.document_name }}
                  </li>
                </ul>
              </div>
              <div class="message-time" v-if="message.created_at">{{ formatMessageTime(message.created_at) }}</div>
            </div>
          </div>
        </div>

        <div class="chat-input-area">
          <div class="input-box">
            <el-input
              ref="messageInputRef"
              v-model="inputMessage"
              type="textarea"
              :rows="4"
              placeholder="输入您的问题..."
              @keydown.ctrl.enter="handleSend"
              @keydown.enter.exact.prevent="handleSend"
              @keydown.shift.enter
              :disabled="loading"
              class="message-input"
            />
          </div>
          
          <div class="input-actions-bar">
            <div class="actions-left">
              <el-button 
                text 
                class="action-btn"
                @click="handleClear"
                title="清空"
              >
                <el-icon><Close /></el-icon>
              </el-button>
            </div>
            <div class="actions-divider"></div>
            <div class="actions-right">
              <el-button 
                circle
                class="send-btn"
                @click="handleSend" 
                :loading="loading" 
                :disabled="!inputMessage.trim()"
                title="发送 (Ctrl+Enter)"
              >
                <el-icon><ArrowUp /></el-icon>
              </el-button>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'
import { 
  User, 
  ChatDotRound,
  Close,
  ArrowUp,
  Loading,
  Link,
  Check,
  Setting
} from '@element-plus/icons-vue'
import { difyChat } from '@/api/dify'
import { listDifyOption } from '@/api/knowsource'
import { parseMarkdown } from '@/utils/markdown'

const userStore = useUserStore()
const messagesContainer = ref(null)
const messageInputRef = ref(null)
const loading = ref(false)
const inputMessage = ref('')
const messages = ref([])
const conversationId = ref('')
const configDialogVisible = ref(false)
const loadingConfigs = ref(false)
const difyOptions = ref([])
const selectedConfig = ref(null)
const currentConfig = ref(null) // 当前使用的配置

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
  loading.value = true
  scrollToBottom()

  // 创建助手消息占位符
  const assistantMessageIndex = messages.value.length
  messages.value.push({
    role: 'assistant',
    content: '',
    created_at: Date.now(),
    retriever_resources: []
  })

  // 检查是否已选择配置
  if (!currentConfig.value) {
    ElMessage.warning('请先选择 Dify 配置')
    configDialogVisible.value = true
    loading.value = false
    messages.value.pop() // 移除助手消息占位符
    messages.value.pop() // 移除用户消息
    return
  }

  try {
    await difyChat(
      {
        query: messageText,
        conversation_id: conversationId.value || '',
        user: userStore.empCode || 'user-123'
      },
      // onChunk
      (chunk) => {
        messages.value[assistantMessageIndex].content = chunk.content
        scrollToBottom()
      },
      // onComplete
      (result) => {
        messages.value[assistantMessageIndex].content = result.content
        
        if (result.conversation_id) {
          conversationId.value = result.conversation_id
        }
        
        const resources = result?.metadata?.retriever_resources || []
        const uniqueResources = deduplicateResources(resources)
        messages.value[assistantMessageIndex].retriever_resources = uniqueResources
        
        loading.value = false
        scrollToBottom()
      },
      // onError
      (error) => {
        ElMessage.error(error.message || '发送失败，请稍后重试')
        messages.value.pop()
        messages.value.pop()
        loading.value = false
      },
      // baseURL 和 apiKey
      currentConfig.value.url,
      currentConfig.value.apiKey
    )
  } catch (error) {
    ElMessage.error('发送失败，请稍后重试')
    messages.value.pop()
    messages.value.pop()
    loading.value = false
  }
}

const handleClear = () => {
  inputMessage.value = ''
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

const formatMessage = (text) => parseMarkdown(text)

const formatMessageTime = (timestamp) => {
  if (!timestamp) return ''
  // created_at 可能是秒级时间戳
  const ts = timestamp < 10000000000 ? timestamp * 1000 : timestamp
  const date = new Date(ts)
  return date.toLocaleString('zh-CN')
}

// 对 retriever_resources 进行去重，基于 document_name
const deduplicateResources = (resources) => {
  if (!resources || !Array.isArray(resources)) return []
  
  const seen = new Set()
  const unique = []
  
  for (const resource of resources) {
    const documentName = resource?.document_name || ''
    if (documentName && !seen.has(documentName)) {
      seen.add(documentName)
      unique.push(resource)
    }
  }
  
  return unique
}

// 加载 Dify 配置列表
const loadDifyOptions = async () => {
  loadingConfigs.value = true
  try {
    const res = await listDifyOption()
    if (res.code === 200 && res.data && res.data.list) {
      difyOptions.value = res.data.list || []
      // 如果只有一个配置，自动选中
      if (difyOptions.value.length === 1) {
        selectedConfig.value = difyOptions.value[0]
        handleConfirmConfig()
      } else if (difyOptions.value.length > 1) {
        // 多个配置时显示选择对话框
        configDialogVisible.value = true
      } else {
        ElMessage.warning('暂无可用配置，请联系管理员')
      }
    } else {
      ElMessage.error(res.message || '加载配置失败')
    }
  } catch (error) {
    ElMessage.error('加载配置失败：' + error.message)
  } finally {
    loadingConfigs.value = false
  }
}

// 选择配置
const handleSelectConfig = (option) => {
  selectedConfig.value = option
}

// 确认选择配置
const handleConfirmConfig = () => {
  if (!selectedConfig.value) {
    ElMessage.warning('请选择一个配置')
    return
  }
  currentConfig.value = selectedConfig.value
  configDialogVisible.value = false
  ElMessage.success(`已选择配置：${selectedConfig.value.name}`)
}

// 取消选择配置（关闭弹窗）
const handleCancelConfig = () => {
  configDialogVisible.value = false
}

// 打开选择智能体弹窗
const handleOpenAgentSelect = () => {
  selectedConfig.value = currentConfig.value
  configDialogVisible.value = true
}

// 切换配置
const handleChangeConfig = () => {
  // 重置选择，显示对话框
  selectedConfig.value = currentConfig.value
  configDialogVisible.value = true
}

onMounted(() => {
  // 先加载配置列表
  loadDifyOptions()
  
  // 添加欢迎消息
  messages.value.push({
    role: 'assistant',
    content: '你好！我是智能体助手，有什么可以帮到您的？',
    created_at: Date.now()
  })
  scrollToBottom()
  
  // 将焦点移到输入框
  nextTick(() => {
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
})
</script>

<style scoped>
.dify-chat-container {
  height: calc(100vh - 120px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.chat-header-bar {
  flex-shrink: 0;
  margin-bottom: 10px;
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
  padding: 12px 16px;
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

.message-resources {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #e4e7ed;
  font-size: 12px;
  color: #909399;
}

.message-resources ul {
  margin: 5px 0 0 0;
  padding-left: 20px;
}

.message-resources li {
  margin: 3px 0;
}

.message-time {
  margin-top: 8px;
  font-size: 12px;
  color: #c0c4cc;
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

.chat-header-config {
  flex-shrink: 0;
  border-bottom: 1px solid #e4e7ed;
  background-color: #fff;
  padding: 12px 20px;
}

.config-info-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.config-name-tag {
  max-width: 300px;
  width: 100%;
  display: inline-flex;
  align-items: left;
  gap: 4px;
}

.config-name-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 600px;
}

.chat-input-area {
  flex-shrink: 0;
  border-top: 1px solid #e4e7ed;
  padding: 16px;
  background-color: #fff;
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

/* 配置选择对话框样式 */
.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
  gap: 12px;
  color: #606266;
}

.empty-tip {
  padding: 40px 0;
  text-align: center;
}

.config-selector {
  max-height: 400px;
  overflow-y: auto;
  padding: 10px 0;
}

.config-cards {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.config-card {
  position: relative;
  padding: 16px;
  border: 2px solid #e4e7ed;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  background-color: #fff;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.config-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
  transform: translateY(-2px);
}

.config-card.active {
  border-color: #409eff;
  background-color: #ecf5ff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.3);
}

.card-content {
  flex: 1;
  min-width: 0;
}

.card-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}

.card-description {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-url {
  font-size: 12px;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.check-icon {
  color: #409eff;
  font-size: 20px;
  flex-shrink: 0;
  margin-left: 8px;
  opacity: 0;
  transition: opacity 0.3s;
}

.config-card:hover .check-icon,
.config-card.active .check-icon {
  opacity: 1;
}

/* 滚动条样式 */
.config-selector::-webkit-scrollbar {
  width: 6px;
}

.config-selector::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.config-selector::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.config-selector::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>

