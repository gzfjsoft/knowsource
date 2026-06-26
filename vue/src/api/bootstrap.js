import axios from 'axios'

/** 不使用带 token 的默认实例，避免初始化阶段 401 跳登录 */
const bootClient = axios.create({
  baseURL: '/api',
  timeout: 60000,
  headers: { 'Content-Type': 'application/json' }
})

export function getBootstrapStatus() {
  return bootClient.get('v1/sys/bootstrap/status').then(r => r.data)
}

export function getBootstrapConfig() {
  return bootClient.get('v1/sys/bootstrap/config').then(r => r.data)
}

export function saveBootstrapConfig(body) {
  return bootClient.post('v1/sys/bootstrap/config', body).then(r => r.data)
}
