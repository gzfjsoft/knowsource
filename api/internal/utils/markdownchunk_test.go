package utils

import (
	"fmt"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func dumpAST(n ast.Node, source []byte, depth int) {
	indent := strings.Repeat("  ", depth)
	kind := n.Kind().String()
	preview := ""
	if n.Type() == ast.TypeBlock && n.Lines() != nil && n.Lines().Len() > 0 {
		var buf []byte
		for i := 0; i < n.Lines().Len(); i++ {
			seg := n.Lines().At(i)
			buf = append(buf, source[seg.Start:seg.Stop]...)
		}
		s := string(buf)
		s = strings.ReplaceAll(s, "\n", " ")
		if len([]rune(s)) > 40 {
			s = string([]rune(s)[:37]) + "..."
		}
		preview = " \"" + s + "\""
	}
	fmt.Printf("%s%s%s\n", indent, kind, preview)
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		dumpAST(c, source, depth+1)
	}
}

func TestChunkSuperSmart(t *testing.T) {
	md := `# 第一章 概述

这是第一段的介绍文字，描述项目背景和目标。

## 1.1 功能特性

- 支持 Markdown 解析
- 基于 goldmark AST 分块
- 按块级节点顺序组织

## 1.2 使用说明

调用 ChunkSuperSmart(content, 100) 即可分块。

<table>
<tr><th>列1</th><th>列2</th></tr>
<tr><td>A</td><td>B</td></tr>
</table>

以上是使用说明。
`

	source := []byte(md)
	doc := goldmark.New().Parser().Parse(text.NewReader(source))

	fmt.Printf("=== Markdown 语法树 (AST) 结构 ===\n")
	dumpAST(doc, source, 0)

	fmt.Printf("\n=== ChunkSuperSmart 分块结果 (chunkSize=80) ===\n")
	m := NewMarkDownChunk()
	chunks := m.ChunkSuperSmart(md, 80)
	fmt.Printf("共 %d 块\n\n", len(chunks))
	for i, c := range chunks {
		fmt.Printf("--- 块 %d (len=%d) ---\n%s\n\n", i+1, len(c), c)
	}
}
