import { defineStore } from 'pinia'
import { login, refreshToken } from '@/api/knowsource'

/** 规范化登录返回的 roles（去空串、trim），避免 [""] 导致无法匹配 user/admin */
function normalizeRoleList(roles) {
  if (!Array.isArray(roles)) return []
  return [...new Set(roles.map((r) => (typeof r === 'string' ? r.trim() : '')).filter(Boolean))]
}

/** 展示用主角色：已知角色按权限优先级，其余取 roles 中第一个（如 demo、自定义角色码） */
const displayRolePriority = ['superadmin', 'admin', 'user', 'demo']

function pickDisplayRole(roles) {
  const list = normalizeRoleList(roles)
  for (const p of displayRolePriority) {
    if (list.includes(p)) return p
  }
  return list[0] || ''
}

/** 仅解码 JWT payload（不验签），用于开发时核对 claims 是否含 clientId */
function decodeJwtPayload(token) {
  if (!token || typeof token !== 'string') return null
  const parts = token.split('.')
  if (parts.length < 2) return null
  try {
    const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const pad = (4 - (base64.length % 4)) % 4
    const padded = base64 + '='.repeat(pad)
    const json = atob(padded)
    return JSON.parse(json)
  } catch {
    return null
  }
}

function logClientIdVsJwt(context, token, userInfo, extra = {}) {
  if (!import.meta.env.DEV) return
  const payload = decodeJwtPayload(token)
  const fromJwt = payload?.clientId ?? payload?.ClientId ?? ''
  const fromUserInfo = (userInfo?.clientId || '').trim()
  const fromLs = (localStorage.getItem('clientId') || '').trim()
  console.log(`[knowsource:${context}] clientId 核对`, {
    userInfoClientId: fromUserInfo || '(无)',
    localStorageClientId: fromLs || '(无)',
    jwtPayloadClientId: fromJwt || '(JWT 无 clientId 字段)',
    jwtMatchesUserInfo:
      !fromJwt && !fromUserInfo ? '(二者皆空)' : fromJwt === fromUserInfo,
    ...extra,
  })
  if (payload && import.meta.env.DEV) {
    console.log(`[knowsource:${context}] JWT payload（未验签）`, payload)
  }
}

export const useUserStore = defineStore('user', {
  state: () => ({
    token: localStorage.getItem('token') || '',
    userInfo: JSON.parse(localStorage.getItem('userInfo') || 'null')
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
    userName: (state) => state.userInfo?.empName || '',
    empCode: (state) => state.userInfo?.empCode || '',
    role: (state) => {
      // 兼容旧版本：如果存在 role 字段，直接返回
      if (state.userInfo?.role) {
        return state.userInfo.role
      }
      // 新版本：roles 数组按优先级取展示角色（含 demo、自定义码兜底为首个）
      return pickDisplayRole(state.userInfo?.roles)
    },
    roles: (state) => normalizeRoleList(state.userInfo?.roles),
    isSuperAdmin: (state) => {
      const roles = normalizeRoleList(state.userInfo?.roles)
      return roles.includes('superadmin')
    },
    /** 平台运维入口：仅 admin 租户且登录账号 empCode 为 superadmin（不走 fr_permissions） */
    isPlatformSuperUser: (state) => {
      const id = (
        state.userInfo?.clientId ||
        (typeof localStorage !== 'undefined'
          ? localStorage.getItem('clientId')
          : '') ||
        ''
      ).trim()
      if (!id) return false
      const emp = (state.userInfo?.empCode || '').trim().toLowerCase()
      return id === 'admin' && emp === 'superadmin'
    },
    /** 当前登录租户 clientId（后端 userInfo 优先，兼容 OA 等未写入 userInfo 时用 localStorage） */
    clientId: (state) => {
      const fromInfo = (state.userInfo?.clientId || '').trim()
      if (fromInfo) return fromInfo
      return (localStorage.getItem('clientId') || '').trim()
    },
    isAdmin: (state) => {
      const roles = normalizeRoleList(state.userInfo?.roles)
      return roles.includes('superadmin') || roles.includes('admin')
    },
    /**
     * 是否显示左侧菜单栏：仅当 fr 权限「有且只有」菜单-知识库问答 时隐藏侧栏（与 role 是否为 user 无关）。
     * 其余情况（含多权限、仅有其它单权限、无权限列表）均显示侧栏。
     */
    showMenu: (state) => {
      const perms = state.userInfo?.empPermissions || []
      const uniq = [...new Set(perms.filter((p) => typeof p === 'string' && p.trim()))]
      const onlyKnowledgeChat =
        uniq.length === 1 && uniq[0] === '菜单-知识库问答'
      return !onlyKnowledgeChat
    },
    empPermissions: (state) => state.userInfo?.empPermissions || []
  },

  actions: {
    async login(loginData) {
      try {
        const res = await login(loginData)
        if (res.code === 200 && res.data) {
          this.token = res.data.token
          this.userInfo = res.data.userInfo
          localStorage.setItem('token', res.data.token)
          localStorage.setItem('userInfo', JSON.stringify(res.data.userInfo))
          logClientIdVsJwt('login', res.data.token, res.data.userInfo, {
            loginFormClientId: (loginData.clientId || '').trim() || '(无)',
          })
          // 触发自定义事件，通知 App.vue 启动 token 刷新定时器
          window.dispatchEvent(new CustomEvent('user-logged-in'))
          return { success: true }
        } else if (res.code === 458) {
          // 邮箱未验证，跳转到邮箱验证页面
          localStorage.setItem('clientId', loginData.clientId)
          window.location.href = '/email/verify'
          return {
            success: false,
            message: res.message || '邮箱未验证'
          }
        }
        // 登录失败，返回错误信息
        return {
          success: false,
          message: res.message || '登录失败，请检查账号密码'
        }
      } catch (error) {
        console.error('Login error:', error)
        // 从错误对象中提取消息
        const errorMessage = error.response?.data?.message || error.message || '登录失败，请稍后重试'
        const errorCode = error.response?.data?.code
        
        if (errorCode === 458) {
          // 邮箱未验证，跳转到邮箱验证页面
          localStorage.setItem('clientId', loginData.clientId)
          window.location.href = '/email/verify'
          return {
            success: false,
            message: errorMessage
          }
        }
        
        return {
          success: false,
          message: errorMessage
        }
      }
    },

    logout() {
      this.token = ''
      this.userInfo = null
      localStorage.removeItem('token')
      localStorage.removeItem('userInfo')
      // 重新登录后进入 ai-chat 需再次弹出知识库选择
      localStorage.removeItem('ai-chat-restore-session')
    },

    async refreshToken() {
      try {
        const res = await refreshToken()
        if (res.code === 200 && res.data) {
          this.token = res.data.token
          this.userInfo = res.data.userInfo
          localStorage.setItem('token', res.data.token)
          localStorage.setItem('userInfo', JSON.stringify(res.data.userInfo))
          logClientIdVsJwt('refreshToken', res.data.token, res.data.userInfo)
          return { success: true }
        }
        return {
          success: false,
          message: res.message || '刷新token失败'
        }
      } catch (error) {
        console.error('Refresh token error:', error)
        return {
          success: false,
          message: error.response?.data?.message || error.message || '刷新token失败'
        }
      }
    }
  }
})

