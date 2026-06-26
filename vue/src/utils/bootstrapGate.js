import { getBootstrapStatus } from '@/api/bootstrap'

let cache = { at: 0, value: null }

const TTL_MS = 4000

/**
 * 返回服务端 bootstrap 状态；短时缓存避免路由抖动。
 * networkError: 请求失败；error: 业务非 200。
 */
export async function getBootstrapGate() {
  const now = Date.now()
  if (cache.value && now - cache.at < TTL_MS) {
    return cache.value
  }
  let value
  try {
    const res = await getBootstrapStatus()
    if (res.code !== 200) {
      value = { error: true, message: res.message }
    } else {
      value = { error: false, networkError: false, ...(res.data || {}) }
    }
  } catch (e) {
    value = { error: true, networkError: true }
  }
  cache = { at: now, value }
  return value
}

export function clearBootstrapGateCache() {
  cache = { at: 0, value: null }
}

const SKIP_INIT_REDIRECT_KEY = 'knowsource_skip_init_redirect'

/** 用户曾拒绝进入初始化向导，本会话内不再自动跳转 */
export function isInitRedirectDeclined() {
  return sessionStorage.getItem(SKIP_INIT_REDIRECT_KEY) === '1'
}

export function setInitRedirectDeclined() {
  sessionStorage.setItem(SKIP_INIT_REDIRECT_KEY, '1')
}

export function clearInitRedirectDeclined() {
  sessionStorage.removeItem(SKIP_INIT_REDIRECT_KEY)
}
