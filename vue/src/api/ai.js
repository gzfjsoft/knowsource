import request from '@/utils/request'
import { useUserStore } from '@/stores/user'

// 数组去重，保持首次出现顺序
function dedupeArray(arr) {
  if (!Array.isArray(arr) || arr.length === 0) return []
  return [...new Set(arr)]
}

// 过滤 <think></think> 标签内的空白字符串和换行符
// 移除所有 \n 直到实际文本内容出现，同时移除 </think> 后的换行和空格
function filterThinkTags(content) {
  if (!content) return content
  
  // 使用正则表达式匹配 <think>...</think> 标签及其后面的空白字符
  return content.replace(/<think>([\s\S]*?)<\/think>([\s\n]*)/g, (match, thinkContent, trailingWhitespace) => {
    // 移除开头的所有换行符和空白字符，直到遇到实际文本
    let filtered = thinkContent.replace(/^[\s\n]+/, '')
    
    // 如果过滤后只剩下空白或为空（没有实际文本），完全移除该标签和后面的空白
    if (!filtered.trim()) {
      return ''
    }
    
    // 返回处理后的 think 标签（保留实际文本内容，移除开头的空白和换行，移除 </think> 后的空白）
    return `<think>${filtered}</think>`
  })
}

// AI 问答
export function aiChat(data) {
  return request({
    url: '/ai/chat',
    method: 'post',
    data
  })
}

// 会话聊天选项（OPTIONS）
export function sessionChatOptions(data) {
  return request({
    url: '/ai/session/chat',
    method: 'options',
    data
  })
}

// 会话聊天（流式）
export async function sessionChat(data, onChunk, onComplete, onError) {
  const userStore = useUserStore()
  const token = userStore.token
  
  try {
    const response = await fetch('/api/ai/session/chat', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': token ? `Bearer ${token}` : ''
      },
      body: JSON.stringify(data)
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ msg: '请求失败' }))
      throw new Error(errorData.msg || `HTTP error! status: ${response.status}`)
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let fullContent = ''
    let fullReasoning = ''
    let sessionId = response.headers.get('SessionUuid') || null
    const llmbackend = response.headers.get('llmbackend') || null // 若为 "ollama" 表示后端使用 Ollama 原生流式格式（每行一个 JSON，含 message.content）
    let filesinfos = []
    let stats = null

    while (true) {
      const { done, value } = await reader.read()
      
      if (done) {
        if (onComplete) {
          onComplete({ content: filterThinkTags(fullContent), reasoning: fullReasoning, session: sessionId, filesinfos, llmbackend })
        }
        break
      }

      const chunk = decoder.decode(value, { stream: true })
      buffer += chunk

      // 按行分割并处理每个JSON对象
      const lines = buffer.split('\n')
      buffer = lines.pop() || '' // 保留最后一行（可能不完整）

      for (const line of lines) {
        const trimmedLine = line.trim()
        if (!trimmedLine) continue

        // 处理 vLLM/OpenAI 格式: "data: {...}"
        let jsonStr = trimmedLine
        if (trimmedLine.startsWith('data: ')) {
          jsonStr = trimmedLine.slice(6) // 移除 'data: '
          if (jsonStr === '[DONE]') {
            continue
          }
        }

        try {
          const jsonData = JSON.parse(jsonStr)

          // 处理 metadata：参考资料 filesinfos（后端先发一行 data: {"filesinfos":[...], "stats": {...}}）
          if ((jsonData.filesinfos && Array.isArray(jsonData.filesinfos)) || jsonData.stats) {
            if (jsonData.filesinfos && Array.isArray(jsonData.filesinfos)) {
              filesinfos = dedupeArray(jsonData.filesinfos)
            }
            if (jsonData.stats) {
              stats = jsonData.stats
            }
            if (onChunk) {
              onChunk({
                content: filterThinkTags(fullContent),
                reasoning: fullReasoning,
                chunk: '',
                filesinfos,
                stats
              })
            }
            continue
          }
          
          // 处理 Ollama 格式: {"message": {"content": "...", "reasoning": "...", "thinking": "..."}, "done": false}
          // reasoning 与 thinking 均为思考过程，合并展示
          if (jsonData.message) {
            const msg = jsonData.message
            let chunkContent = ''
            if (msg.content !== undefined) {
              chunkContent = filterThinkTags(msg.content)
              fullContent += chunkContent
            }
            const chunkReason = (msg.reasoning || '') + (msg.thinking || '')
            if (chunkReason) {
              fullReasoning += chunkReason
            }
            if (msg.content !== undefined || msg.reasoning !== undefined || msg.thinking !== undefined) {
              if (onChunk) {
                onChunk({
                  content: filterThinkTags(fullContent),
                  reasoning: fullReasoning,
                  chunk: chunkContent,
                  done: jsonData.done || false,
                  session: jsonData.session || sessionId,
                  filesinfos: filesinfos.length ? filesinfos : undefined
                })
              }
            }
          }
          // 处理 vLLM/OpenAI 格式: {"choices": [{"delta": {"content": "...", "reasoning": "...", "thinking": "..."}}]}
          // reasoning 与 thinking 均为思考过程，合并展示
          else if (jsonData.choices && Array.isArray(jsonData.choices)) {
            for (const choice of jsonData.choices) {
              const delta = choice.delta || choice.message || {}
              let chunkContent = delta.content !== undefined ? (delta.content || '') : ''
              const chunkReasoning = (delta.reasoning || '') + (delta.thinking || '')
              if (choice.message && choice.message.content !== undefined && chunkContent === '') {
                chunkContent = choice.message.content || ''
              }
              if (chunkContent) {
                chunkContent = filterThinkTags(chunkContent)
                fullContent += chunkContent
              }
              if (chunkReasoning) {
                fullReasoning += chunkReasoning
              }
              if (chunkContent || chunkReasoning) {
                if (onChunk) {
                  onChunk({
                    content: filterThinkTags(fullContent),
                    reasoning: fullReasoning,
                    chunk: chunkContent,
                    done: choice.finish_reason !== null || false,
                    session: sessionId,
                    filesinfos: filesinfos.length ? filesinfos : undefined
                  })
                }
              }
            }
          }

          // 如果完成，更新 sessionId (Ollama 格式)
          if (jsonData.done) {
            if (jsonData.session) {
              sessionId = jsonData.session
            }
          }
        } catch (e) {
          // 忽略解析失败的行（可能是部分数据）
          // console.warn('跳过无效JSON行:', trimmedLine, e)
        }
      }
      
      // 处理 buffer 中剩余的完整 JSON（如果存在）
      if (buffer.trim()) {
        try {
          let jsonStr = buffer.trim()
          // 处理 vLLM/OpenAI 格式: "data: {...}"
          if (jsonStr.startsWith('data: ')) {
            jsonStr = jsonStr.slice(6)
            if (jsonStr === '[DONE]') {
              buffer = ''
              continue
            }
          }
          
          const jsonData = JSON.parse(jsonStr)

          if ((jsonData.filesinfos && Array.isArray(jsonData.filesinfos)) || jsonData.stats) {
            if (jsonData.filesinfos && Array.isArray(jsonData.filesinfos)) {
              filesinfos = dedupeArray(jsonData.filesinfos)
            }
            if (jsonData.stats) {
              stats = jsonData.stats
            }
            if (onChunk) {
              onChunk({
                content: filterThinkTags(fullContent),
                reasoning: fullReasoning,
                chunk: '',
                filesinfos,
                stats
              })
            }
          }
          // 处理 Ollama 格式（reasoning + thinking 合并为思考过程）
          else if (jsonData.message) {
            const msg = jsonData.message
            let chunkContent = ''
            if (msg.content !== undefined) {
              chunkContent = filterThinkTags(msg.content)
              fullContent += chunkContent
            }
            const chunkReason = (msg.reasoning || '') + (msg.thinking || '')
            if (chunkReason) {
              fullReasoning += chunkReason
            }
            if (msg.content !== undefined || msg.reasoning !== undefined || msg.thinking !== undefined) {
              if (onChunk) {
                onChunk({
                  content: filterThinkTags(fullContent),
                  reasoning: fullReasoning,
                  chunk: chunkContent,
                  done: jsonData.done || false,
                  session: jsonData.session || sessionId,
                  filesinfos: filesinfos.length ? filesinfos : undefined
                })
              }
            }
          }
          // 处理 vLLM/OpenAI 格式（reasoning + thinking 合并为思考过程）
          else if (jsonData.choices && Array.isArray(jsonData.choices)) {
            for (const choice of jsonData.choices) {
              const delta = choice.delta || choice.message || {}
              let chunkContent = delta.content !== undefined ? (delta.content || '') : ''
              const chunkReasoning = (delta.reasoning || '') + (delta.thinking || '')
              if (choice.message && choice.message.content !== undefined && chunkContent === '') {
                chunkContent = choice.message.content || ''
              }
              if (chunkContent) {
                chunkContent = filterThinkTags(chunkContent)
                fullContent += chunkContent
              }
              if (chunkReasoning) {
                fullReasoning += chunkReasoning
              }
              if (chunkContent || chunkReasoning) {
                if (onChunk) {
                  onChunk({
                    content: filterThinkTags(fullContent),
                    reasoning: fullReasoning,
                    chunk: chunkContent,
                    done: choice.finish_reason !== null || false,
                    session: sessionId,
                    filesinfos: filesinfos.length ? filesinfos : undefined
                  })
                }
              }
            }
          }
          
          if (jsonData.done && jsonData.session) {
            sessionId = jsonData.session
          }
          buffer = '' // 清空 buffer
        } catch (e) {
          // buffer 中的 JSON 可能不完整，保留在 buffer 中等待下次数据
        }
      }
    }
  } catch (error) {
    if (onError) {
      onError(error)
    } else {
      throw error
    }
  }
}

// 历史会话列表
export function getHistoryList() {
  return request({
    url: '/ai/chat/history/list',
    method: 'post'
  })
}

// 历史会话详情
export function getSessionDetail(data) {
  return request({
    url: '/ai/chat/history/detail',
    method: 'post',
    data
  })
}

// 删除历史会话
export function deleteSession(data) {
  return request({
    url: '/ai/chat/history/delete',
    method: 'post',
    data
  })
}

// 批量删除历史会话
export function batchDeleteSession(data) {
  return request({
    url: '/ai/chat/history/batch-delete',
    method: 'post',
    data
  })
}

// AI 对话上传临时文档（txt/docx/pdf），识别内容并写入对话缓存，发下一条消息时作为参考
// config 可选，如 { onUploadProgress: (e) => {} } 用于上传进度
export function uploadChatDocument(formData, config = {}) {
  return request({
    url: '/ai/chat/upload-document',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    ...config
  })
}

// AI 对话从缓存中移除已上传的文档（按文件名）
export function removeChatDocument(data) {
  return request({
    url: '/ai/chat/upload-document/remove',
    method: 'post',
    data
  })
}

// 获取 AI 问候语
export function getAIGreet() {
  return request({
    url: '/conf/ai/name/问候词',
    method: 'get'
  })
}

