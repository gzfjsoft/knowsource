// Dify Chat API — configure via Vite env:
// - VITE_DIFY_BASE_URL
// - VITE_DIFY_API_KEY
const DIFY_BASE_URL = import.meta.env.VITE_DIFY_BASE_URL || ''
const DIFY_API_KEY = import.meta.env.VITE_DIFY_API_KEY || ''

function requireDifyConfig(baseURL, apiKey) {
  const url = (baseURL || DIFY_BASE_URL || '').replace(/\/$/, '')
  const key = apiKey || DIFY_API_KEY
  if (!url) {
    throw new Error('Dify base URL 未配置，请设置 VITE_DIFY_BASE_URL')
  }
  if (!key) {
    throw new Error('Dify API key 未配置，请设置 VITE_DIFY_API_KEY')
  }
  return { url, key }
}

// Dify 流式聊天
export async function difyChat(data, onChunk, onComplete, onError, baseURL, apiKey) {
  let url
  let key
  try {
    ({ url, key } = requireDifyConfig(baseURL, apiKey))
  } catch (error) {
    if (onError) {
      onError(error)
      return
    }
    throw error
  }

  try {
    const response = await fetch(`${url}/v1/chat-messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${key}`
      },
      body: JSON.stringify({
        inputs: data.inputs || {},
        query: data.query,
        response_mode: 'streaming',
        conversation_id: data.conversation_id || '',
        user: data.user || 'user-123',
        files: data.files || [],
        model_config: data.model_config || {}
      })
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({ message: '请求失败' }))
      throw new Error(errorData.message || `HTTP error! status: ${response.status}`)
    }

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let buffer = ''
    let fullContent = ''
    let conversationId = ''
    let messageId = ''
    let metadata = null
    let files = null

    while (true) {
      const { done, value } = await reader.read()

      if (done) {
        if (onComplete) {
          onComplete({
            content: fullContent,
            conversation_id: conversationId,
            message_id: messageId,
            metadata,
            files
          })
        }
        break
      }

      const chunk = decoder.decode(value, { stream: true })
      buffer += chunk

      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmedLine = line.trim()
        if (!trimmedLine || !trimmedLine.startsWith('data: ')) continue

        try {
          const jsonStr = trimmedLine.substring(6)
          const jsonData = JSON.parse(jsonStr)

          if (jsonData.event === 'message') {
            if (jsonData.answer !== undefined) {
              fullContent += jsonData.answer
              if (onChunk) {
                onChunk({
                  content: fullContent,
                  chunk: jsonData.answer,
                  conversation_id: jsonData.conversation_id,
                  message_id: jsonData.message_id
                })
              }
            }

            if (jsonData.conversation_id) {
              conversationId = jsonData.conversation_id
            }
            if (jsonData.message_id) {
              messageId = jsonData.message_id
            }
          }

          if (jsonData.event === 'message_end') {
            if (jsonData.metadata) {
              metadata = jsonData.metadata
            }
            if (jsonData.files !== undefined) {
              files = jsonData.files
            }
            if (jsonData.conversation_id) {
              conversationId = jsonData.conversation_id
            }
            if (jsonData.message_id) {
              messageId = jsonData.message_id
            }
          }
        } catch (e) {
          console.warn('跳过无效JSON行:', trimmedLine, e)
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
