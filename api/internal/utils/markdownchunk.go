package utils

import (
	"bytes"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ChunkMode 分块模式
type ChunkMode string

const (
	ChunkModeNone         ChunkMode = "none"          // 不分块
	ChunkModeSimple       ChunkMode = "simple"        // 简单分块
	ChunkModeSmart        ChunkMode = "smart"         // 智能分块
	ChunkModeSuperSmart   ChunkMode = "super_smart"   // 基于 AST 的智能分块
	ChunkModeAISeparator  ChunkMode = "ai_separator"  // 按 <AI分隔符> 分块
)

// MarkDownChunk Markdown 文档分块工具类
type MarkDownChunk struct{}

// NewMarkDownChunk 创建新的 MarkDownChunk 实例
func NewMarkDownChunk() *MarkDownChunk {
	return &MarkDownChunk{}
}

// ChunkNone 不分块模式：直接返回整个文件内容
func (m *MarkDownChunk) ChunkNone(fileContent string) []string {
	if fileContent == "" {
		return []string{}
	}
	return []string{fileContent}
}

// ChunkSimple 简单分块模式：按长度直接剪断，带重叠部分
// chunkSize: 分块大小（字符数）
// chunkOverlap: 重叠大小（字符数）
func (m *MarkDownChunk) ChunkSimple(fileContent string, chunkSize int, chunkOverlap int) []string {
	if fileContent == "" {
		return []string{}
	}

	chunks := []string{}
	start := 0
	contentLength := len(fileContent)

	for start < contentLength {
		end := start + chunkSize
		if end > contentLength {
			end = contentLength
		}

		chunk := fileContent[start:end]
		chunks = append(chunks, chunk)

		if end >= contentLength {
			break
		}

		start = end - chunkOverlap
		if start < 0 {
			start = 0
		}
	}

	return chunks
}

// ChunkSmart 智能分块模式：先按长度分块，分块后 trim \n 空格，然后从后边向前查，按优先级查找分割点：
// 1. \n# (markdown标题)
// 2. \n\n (双换行)
// 3. 中文句号。
// 4. 英文句号.
func (m *MarkDownChunk) ChunkSmart(fileContent string, chunkSize int) []string {
	if fileContent == "" {
		return []string{}
	}

	chunks := []string{}
	start := 0
	contentLength := len(fileContent)

	for start < contentLength {
		end := start + chunkSize
		if end > contentLength {
			end = contentLength
		}

		if end >= contentLength {
			// 到达文件末尾，直接添加剩余内容并 trim
			chunk := strings.Trim(fileContent[start:end], "\n ")
			if chunk != "" {
				chunks = append(chunks, chunk)
			}
			break
		}

		// 从后边向前查找分割点，按优先级查找
		splitPos := -1
		searchStart := end - 1

		// 优先级1: 查找 \n# (markdown标题)
		for searchStart >= start {
			if searchStart > start && searchStart > 0 {
				if fileContent[searchStart-1] == '\n' && fileContent[searchStart] == '#' {
					splitPos = searchStart - 1 // 找到 \n 的位置
					break
				}
			}
			searchStart--
		}

		// 优先级2: 如果没有找到 \n#，查找 \n\n (双换行)
		if splitPos < 0 {
			searchStart = end - 1
			for searchStart >= start {
				if searchStart > start && searchStart > 0 {
					if fileContent[searchStart-1] == '\n' && fileContent[searchStart] == '\n' {
						splitPos = searchStart + 1 // 在第二个 \n 之后分割
						break
					}
				}
				searchStart--
			}
		}

		// 优先级3: 如果没有找到 \n\n，查找中文句号。
		if splitPos < 0 {
			searchStart = end - 1
			for searchStart >= start {
				if rune(fileContent[searchStart]) == '。' {
					splitPos = searchStart + 1 // 在句号之后分割
					break
				}
				searchStart--
			}
		}

		// 优先级4: 如果没有找到中文句号，查找英文句号.
		if splitPos < 0 {
			searchStart = end - 1
			for searchStart >= start {
				if fileContent[searchStart] == '.' {
					// 确保不是小数点或URL中的点（简单判断：前后不是数字）
					prevIsDigit := searchStart > start && unicode.IsDigit(rune(fileContent[searchStart-1]))
					nextIsDigit := searchStart < end-1 && unicode.IsDigit(rune(fileContent[searchStart+1]))
					if !prevIsDigit && !nextIsDigit {
						splitPos = searchStart + 1 // 在句号之后分割
						break
					}
				}
				searchStart--
			}
		}

		if splitPos >= 0 && splitPos > start {
			// 如果找到了分割点，在它之前切分（确保 splitPos > start 避免死循环）
			actualChunk := strings.Trim(fileContent[start:splitPos], "\n ")
			if actualChunk != "" {
				chunks = append(chunks, actualChunk)
			}
			// 下一个块的起始位置从分割点开始
			start = splitPos
		} else {
			// 如果没有找到任何分割点，或者 splitPos <= start（避免死循环），这个块就这样了（trim 后添加）
			chunkText := strings.Trim(fileContent[start:end], "\n ")
			if chunkText != "" {
				chunks = append(chunks, chunkText)
			}
			start = end
		}
	}

	return chunks
}

// blockNodeExtent 获取块节点的原文范围，返回 start, stop（含子树的完整范围）
func blockNodeExtent(n ast.Node, source []byte) (start, stop int, ok bool) {
	lines := n.Lines()
	if lines != nil && lines.Len() > 0 {
		first := lines.At(0)
		last := lines.At(lines.Len() - 1)
		return first.Start, last.Stop, true
	}
	// 无 Lines 的容器节点：从子树聚合
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		s, e, childOk := blockNodeExtent(c, source)
		if !childOk {
			continue
		}
		if !ok {
			start, stop, ok = s, e, true
			continue
		}
		if s < start {
			start = s
		}
		if e > stop {
			stop = e
		}
	}
	return start, stop, ok
}

// extendToLineBounds 将范围扩展到所在行的完整边界，保留 # 等行首 Markdown 标记
func extendToLineBounds(source []byte, start, stop int) (int, int) {
	for start > 0 && source[start-1] != '\n' {
		start--
	}
	for stop < len(source) && source[stop] != '\n' {
		stop++
	}
	if stop < len(source) && source[stop] == '\n' {
		stop++
	}
	return start, stop
}

// blockNodeSource 获取块节点的完整原文（含 # 等 Markdown 标记）
func blockNodeSource(n ast.Node, source []byte) string {
	start, stop, ok := blockNodeExtent(n, source)
	if !ok {
		return ""
	}
	start, stop = extendToLineBounds(source, start, stop)
	return string(source[start:stop])
}

// astBlock 收集到的 AST 块，带类型标记
type astBlock struct {
	text     string
	isHeading bool
}

// ChunkSuperSmart 基于 Markdown AST 的智能分块：使用 goldmark 解析为 AST，
// 按 Heading 切分为“节”：两个 Heading 之间（含前一个 Heading）成为一个块。
// List、Blockquote、HTMLBlock（含 <table> 等）不可拆分，完整保留。当缓冲超过 2*size 时输出。
func (m *MarkDownChunk) ChunkSuperSmart(fileContent string, chunkSize int) []string {
	if fileContent == "" {
		return []string{}
	}

	source := []byte(fileContent)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))
	maxBufSize := chunkSize * 2

	// 收集块级节点（按 AST 顺序），标记是否为 Heading
	var blocks []astBlock
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if n.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}
		if n.Kind() == ast.KindDocument {
			return ast.WalkContinue, nil
		}
		isHeading := n.Kind() == ast.KindHeading
		var text string
		// List、Blockquote、HTMLBlock 等不可拆分，整块保留（含 <table> 等 HTML 标签）
		if n.Kind() == ast.KindList || n.Kind() == ast.KindBlockquote || n.Kind() == ast.KindHTMLBlock {
			text = blockNodeSource(n, source)
			blocks = append(blocks, astBlock{text: text, isHeading: false})
			return ast.WalkSkipChildren, nil
		}
		// 使用 blockNodeSource 获取完整原文，保留 # 等 Markdown 标记
		text = blockNodeSource(n, source)
		if text == "" {
			return ast.WalkContinue, nil
		}
		blocks = append(blocks, astBlock{text: text, isHeading: isHeading})
		return ast.WalkContinue, nil
	})

	if len(blocks) == 0 {
		return m.ChunkSmart(fileContent, chunkSize)
	}

	// 按 Heading 分组：两个 Heading 之间为一个 section
	var sections []string
	var sectionBuf bytes.Buffer
	for _, b := range blocks {
		if b.isHeading && sectionBuf.Len() > 0 {
			sections = append(sections, strings.Trim(sectionBuf.String(), "\n "))
			sectionBuf.Reset()
		}
		if sectionBuf.Len() > 0 {
			sectionBuf.WriteString("\n\n")
		}
		sectionBuf.WriteString(b.text)
	}
	if sectionBuf.Len() > 0 {
		sections = append(sections, strings.Trim(sectionBuf.String(), "\n "))
	}

	// 按 section 分块，每个 section 不可拆分
	chunks := []string{}
	var buf bytes.Buffer
	for _, sec := range sections {
		if sec == "" {
			continue
		}
		wouldBe := buf.Len()
		if buf.Len() > 0 {
			wouldBe += 2
		}
		wouldBe += len(sec)

		if wouldBe > maxBufSize && buf.Len() > 0 {
			chunk := strings.Trim(buf.String(), "\n ")
			if chunk != "" {
				chunks = append(chunks, chunk)
			}
			buf.Reset()
		}
		if buf.Len() > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(sec)
	}
	if buf.Len() > 0 {
		chunk := strings.Trim(buf.String(), "\n ")
		if chunk != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// AISeparator 按 AI 分割符分块时使用的标记
const AISeparator = "<AI分隔符>"

// ChunkByAISeparator 按 <AI分隔符> 分块：若内容包含该标记则按此分块，否则返回 nil（调用方需改用其他模式）
func (m *MarkDownChunk) ChunkByAISeparator(fileContent string) []string {
	if fileContent == "" {
		return []string{}
	}
	if !strings.Contains(fileContent, AISeparator) {
		return nil
	}
	parts := strings.Split(fileContent, AISeparator)
	chunks := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			chunks = append(chunks, trimmed)
		}
	}
	return chunks
}
