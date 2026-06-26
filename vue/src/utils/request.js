import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useUserStore } from '@/stores/user'

const service = axios.create({
  baseURL: '/api',
  timeout: 300000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
service.interceptors.request.use(
  config => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    // AdminJWT 等路径从 Header 取租户；普通 JWT 也会带上，与后端 context 一致
    const cid = (userStore.clientId || '').trim()
    if (cid) {
      config.headers['Client-Id'] = cid
    }
    return config
  },
  error => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
service.interceptors.response.use(
  response => {
    const res = response.data
    
    // 如果响应状态码不是 200，则视为错误
    if (res.code !== 200 && res.code !== 0) {
      // 不在这里显示错误消息，让调用方处理
      const error = new Error(res.message || res.msg || '请求失败')
      error.response = response
      return Promise.reject(error)
    }
    
    return res
  },
  error => {
    console.error('Response error:', error)
    
    if (error.response) {
      const { status, data } = error.response
      
      if (status === 401) {
        ElMessage.error('未授权，请重新登录')
        const userStore = useUserStore()
        userStore.logout()
        window.location.href = '/login'
      } else if (status === 403) {
        ElMessage.error('拒绝访问')
      } else if (status === 404) {
        ElMessage.error('请求的资源不存在')
      } else if (status >= 500) {
        ElMessage.error('服务器错误')
      } else {
        ElMessage.error(data?.message || data?.msg || error.message || '请求失败')
      }
    } else {
      ElMessage.error('网络错误，请检查网络连接')
    }
    
    return Promise.reject(error)
  }
)

export default service

