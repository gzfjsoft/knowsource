import { marked } from 'marked'

marked.setOptions({
  gfm: true,
  breaks: false,
})

/**
 * 将 Markdown 转为 HTML（GFM：列表、引用、表格、代码块等）
 */
export function parseMarkdown(md) {
  if (md == null || md === '') return ''
  const html = marked.parse(String(md))
  return typeof html === 'string' ? html : ''
}

/**
 * 重写 HTML 中的相对图片/链接路径（用于文档预览）
 */
export function rewriteHtmlAssetUrls(
  html,
  { resolveImageSrc, resolveLinkHref } = {}
) {
  if (!html) return ''
  let out = html
  if (typeof resolveImageSrc === 'function') {
    out = out.replace(
      /<img([^>]*)\ssrc=["']([^"']+)["']/gi,
      (match, attrs, src) => {
        const resolved = resolveImageSrc(src)
        return `<img${attrs} src="${resolved}"`
      }
    )
  }
  if (typeof resolveLinkHref === 'function') {
    out = out.replace(
      /<a([^>]*)\shref=["']([^"']+)["']/gi,
      (match, attrs, href) => {
        if (
          /^https?:\/\//i.test(href) ||
          href.startsWith('#') ||
          href.startsWith('mailto:')
        ) {
          return match
        }
        const resolved = resolveLinkHref(href)
        if (!resolved || resolved === href) return match
        return `<a${attrs} href="${resolved}"`
      }
    )
  }
  return out
}

export function parseMarkdownWithAssets(md, assetOptions) {
  return rewriteHtmlAssetUrls(parseMarkdown(md), assetOptions)
}
