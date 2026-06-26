import request from '@/utils/request'
import axios from 'axios'
import { useUserStore } from '@/stores/user'

// ==================== 系统版本 ====================
export function getSysVersion() {
  return request({
    url: '/sys/version',
    method: 'get'
  })
}

// ==================== 文档类型相关 ====================
export function listDocumentsType(data) {
  return request({
    url: '/documents/type/list',
    method: 'post',
    data
  })
}

export function createDocumentsType(data) {
  return request({
    url: '/documents/type/create',
    method: 'post',
    data
  })
}

export function updateDocumentsType(data) {
  return request({
    url: '/documents/type/update',
    method: 'post',
    data
  })
}

export function deleteDocumentsType(data) {
  return request({
    url: '/documents/type/delete',
    method: 'post',
    data
  })
}

export function getDocumentsType(data) {
  return request({
    url: '/documents/type/get',
    method: 'post',
    data
  })
}

// ==================== 原始文档相关 ====================
export function uploadRawDocuments(formData, onUploadProgress) {
  return request({
    url: '/raw-documents/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    onUploadProgress: onUploadProgress
  })
}

export function listRawDocuments(data) {
  return request({
    url: '/raw-documents/list',
    method: 'post',
    data
  })
}

export function searchRawDocuments(data) {
  return request({
    url: '/raw-documents/search',
    method: 'post',
    data
  })
}

export function deleteRawDocuments(data) {
  return request({
    url: '/raw-documents/delete',
    method: 'post',
    data
  })
}

export function changeFileDocumentType(data) {
  return request({
    url: '/knowsource/admin/change/file/document-type',
    method: 'post',
    data
  })
}

export function getRawDocuments(data) {
  return request({
    url: '/raw-documents/get',
    method: 'post',
    data
  })
}

// 根据文件名获取原始文档（AI 对话参考资料预览用）
export function getRawDocumentsByFilename(data) {
  return request({
    url: '/raw-documents/get-by-filename',
    method: 'post',
    data
  })
}

// 下载原始文档源文件（返回 axios response，data 为 Blob）
export function downloadRawDocumentsFile(id) {
  const userStore = useUserStore()
  const headers = {}
  if (userStore?.token) {
    headers.Authorization = `Bearer ${userStore.token}`
  }

  return axios({
    url: `/api/raw-documents/download/${id}`,
    method: 'get',
    responseType: 'blob',
    headers
  })
}

export function getDistinctTags(data) {
  return request({
    url: '/raw-documents/tags/distinct',
    method: 'post',
    data
  })
}

export function changeRawDocumentsTag(data) {
  return request({
    url: '/raw-documents/tag/change',
    method: 'post',
    data
  })
}

export function auditRawDocuments(data) {
  return request({
    url: '/raw-documents/audit',
    method: 'post',
    data
  })
}

export function cancelAuditRawDocuments(data) {
  return request({
    url: '/raw-documents/audit/cancel',
    method: 'post',
    data
  })
}

/** 已审核文档在 Qdrant 中的分块列表（主集合 + 全文概要集合） */
export function getRawDocumentQdrantChunks(data) {
  return request({
    url: '/raw-documents/qdrant/chunks',
    method: 'post',
    data
  })
}

/** 审核入库时抽取的问答队列 */
export function listRawDocumentQaPairs(data) {
  return request({
    url: '/raw-documents/qa/list',
    method: 'post',
    data
  })
}

export function updateRawDocumentsContent(data) {
  return request({
    url: '/raw-documents/content/update',
    method: 'post',
    data
  })
}

/** LLM 规范化 Markdown 预览（返回原文与格式化结果，确认后调 updateRawDocumentsContent 保存） */
export function previewRawDocumentsMarkdownNormalize(data) {
  return request({
    url: '/raw-documents/content/markdown-normalize/preview',
    method: 'post',
    data
  })
}

// ==================== 获取原始文档内容差异 ====================
export function getRawDocumentsContentDiff(data) {
  return request({
    url: '/raw-documents/content/diff',
    method: 'post',
    data
  })
}

// ==================== 检查所有原始文档文件是否存在于硬盘 ====================
export function checkRawDocumentsFileExists() {
  return request({
    url: '/raw-documents/check-file-exists',
    method: 'post'
  })
}

// ==================== AI 配置相关 ====================
export function listAIConfig(params) {
  return request({
    url: '/conf/ai/list',
    method: 'get',
    params
  })
}

export function createAIConfig(data) {
  return request({
    url: '/conf/ai',
    method: 'post',
    data
  })
}

export function updateAIConfig(id, data) {
  return request({
    url: `/conf/ai/${id}`,
    method: 'post',
    data
  })
}

export function deleteAIConfig(id) {
  return request({
    url: `/conf/ai/${id}`,
    method: 'delete'
  })
}

export function getAIConfigByName(name) {
  return request({
    url: `/conf/ai/name/${encodeURIComponent(name)}`,
    method: 'get'
  })
}

// ==================== LLM 设置相关 ====================
export function loadLLMSetting() {
  return request({
    url: '/knowsource/llm/setting/load',
    method: 'post'
  })
}

export function saveLLMSetting(data) {
  return request({
    url: '/knowsource/llm/setting/save',
    method: 'post',
    data
  })
}

/** 获取 LLM 系统默认配置（knowsource.yaml），供界面恢复默认，不写入文件 */
export function loadLLMSettingDefaults() {
  return request({
    url: '/knowsource/llm/setting/defaults',
    method: 'post'
  })
}

/** 获取对话模型列表（与监控检查中 Chat 使用同一接口 /v1/models 或 Ollama /api/tags），排除名称含 embedding 的模型 */
export function getLLMChatModels() {
  return request({
    url: '/knowsource/llm/chat/models',
    method: 'post'
  })
}

/** 获取 Embedding 模型列表（从 Rag.EmbeddingsUrl 拉取），仅返回名称含 embedding 的模型 */
export function getLLMEmbeddingModels() {
  return request({
    url: '/knowsource/llm/embedding/models',
    method: 'post'
  })
}

export function callLLMOneShot(data) {
  return request({
    url: '/ai/llm/oneshot',
    method: 'post',
    data
  })
}

export function testLLMServices(data = {}) {
  return request({
    url: '/knowsource/llm/service/test',
    method: 'post',
    data
  })
}

