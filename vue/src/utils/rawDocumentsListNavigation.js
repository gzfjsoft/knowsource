/**
 * 文档管理列表（RawDocuments）与详情/预览页之间传递分页、筛选状态。
 * 列表页跳转时带 query.list，返回时用 navigateBackToRawDocumentsList 还原。
 */

export function listQueryFromState (s) {
  const q = {
    page: String(s.page ?? 1),
    pageSize: String(s.pageSize ?? 10)
  }
  if (s.documentCode) q.documentCode = String(s.documentCode)
  if (s.fileName) q.fileName = String(s.fileName)
  if (s.tag) q.tag = String(s.tag)
  if (s.isAudit !== undefined && s.isAudit !== null && s.isAudit !== '') {
    q.isAudit = String(s.isAudit)
  }
  return q
}

/** 将列表快照编码为路由 query.list */
export function encodeListStateForRoute (snapshot) {
  return encodeURIComponent(JSON.stringify(snapshot))
}

/**
 * 从预览/内容页返回文档列表：若有 list 则带 query 回到原页；否则 history.back
 */
export function navigateBackToRawDocumentsList (router, route) {
  const raw = route.query.list
  if (raw) {
    const str = Array.isArray(raw) ? raw[0] : raw
    try {
      let json = str
      try {
        json = decodeURIComponent(str)
      } catch (e) {
        json = str
      }
      const s = JSON.parse(json)
      if (s && typeof s === 'object') {
        router.push({
          name: 'RawDocuments',
          query: listQueryFromState(s)
        })
        return
      }
    } catch (e) {
      // fall through
    }
  }
  router.back()
}
