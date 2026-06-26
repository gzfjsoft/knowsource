package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"testing"
)

func TestExtractMarkdown(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	tests := []struct {
		name     string
		result   map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name: "标准响应格式 - 从 results 中提取",
			result: map[string]interface{}{
				"backend": "pipeline",
				"version": "1.0.0",
				"results": map[string]interface{}{
					"test_document.pdf": map[string]interface{}{
						"md_content": "# Test Document\n\nThis is test content.",
					},
				},
			},
			expected: "# Test Document\n\nThis is test content.",
			wantErr:  false,
		},
		{
			name: "多个文件 - 提取第一个有内容的",
			result: map[string]interface{}{
				"backend": "pipeline",
				"version": "1.0.0",
				"results": map[string]interface{}{
					"file1.pdf": map[string]interface{}{
						"md_content": "",
					},
					"file2.pdf": map[string]interface{}{
						"md_content": "# File 2 Content\n\nContent here.",
					},
				},
			},
			expected: "# File 2 Content\n\nContent here.",
			wantErr:  false,
		},
		{
			name: "使用 md 字段（向后兼容）",
			result: map[string]interface{}{
				"results": map[string]interface{}{
					"test.pdf": map[string]interface{}{
						"md": "# Test\n\nContent.",
					},
				},
			},
			expected: "# Test\n\nContent.",
			wantErr:  false,
		},
		{
			name: "使用 markdown 字段（向后兼容）",
			result: map[string]interface{}{
				"results": map[string]interface{}{
					"test.pdf": map[string]interface{}{
						"markdown": "# Markdown\n\nContent.",
					},
				},
			},
			expected: "# Markdown\n\nContent.",
			wantErr:  false,
		},
		{
			name: "顶层直接包含 md_content（向后兼容）",
			result: map[string]interface{}{
				"md_content": "# Direct Content\n\nTest.",
			},
			expected: "# Direct Content\n\nTest.",
			wantErr:  false,
		},
		{
			name: "没有 Markdown 内容",
			result: map[string]interface{}{
				"backend": "pipeline",
				"version": "1.0.0",
				"results": map[string]interface{}{
					"test.pdf": map[string]interface{}{
						"middle_json": "{}",
					},
				},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "空的 results",
			result: map[string]interface{}{
				"backend": "pipeline",
				"version": "1.0.0",
				"results": map[string]interface{}{},
			},
			expected: "",
			wantErr:  true,
		},
		{
			name: "没有 results 字段",
			result: map[string]interface{}{
				"backend": "pipeline",
				"version": "1.0.0",
			},
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := client.ExtractMarkdown(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("ExtractMarkdown() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultParseOptions(t *testing.T) {
	options := DefaultParseOptions()

	if len(options.LangList) == 0 || options.LangList[0] != "ch" {
		t.Errorf("DefaultParseOptions().LangList = %v, want [\"ch\"]", options.LangList)
	}

	if options.Backend != "pipeline" {
		t.Errorf("DefaultParseOptions().Backend = %v, want %v", options.Backend, "pipeline")
	}

	if options.ParseMethod != "auto" {
		t.Errorf("DefaultParseOptions().ParseMethod = %v, want %v", options.ParseMethod, "auto")
	}

	if !options.ReturnMD {
		t.Errorf("DefaultParseOptions().ReturnMD = %v, want %v", options.ReturnMD, true)
	}

	if options.ReturnMiddleJSON {
		t.Errorf("DefaultParseOptions().ReturnMiddleJSON = %v, want %v", options.ReturnMiddleJSON, false)
	}

	if !options.FormulaEnable {
		t.Errorf("DefaultParseOptions().FormulaEnable = %v, want %v", options.FormulaEnable, true)
	}

	if !options.TableEnable {
		t.Errorf("DefaultParseOptions().TableEnable = %v, want %v", options.TableEnable, true)
	}
}

func TestNewMinerUClient(t *testing.T) {
	baseURL := "http://127.0.0.1:8100"
	client := NewMinerUClient(baseURL)

	if client.BaseURL != baseURL {
		t.Errorf("NewMinerUClient().BaseURL = %v, want %v", client.BaseURL, baseURL)
	}

	if client.Client == nil {
		t.Error("NewMinerUClient().Client should not be nil")
	}
}

// TestExtractMarkdownWithRealJSON 测试真实的 JSON 响应格式
func TestExtractMarkdownWithRealJSON(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	// 模拟真实的 MinerU API 响应
	jsonResponse := `{
		"backend": "pipeline",
		"version": "0.1.0",
		"results": {
			"example_document.pdf": {
				"md_content": "# 示例文档\n\n这是文档内容。\n\n## 章节1\n\n内容1。\n\n## 章节2\n\n内容2。"
			}
		}
	}`

	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonResponse), &result)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	md, err := client.ExtractMarkdown(result)
	if err != nil {
		t.Fatalf("ExtractMarkdown() error = %v", err)
	}

	expected := "# 示例文档\n\n这是文档内容。\n\n## 章节1\n\n内容1。\n\n## 章节2\n\n内容2。"
	if md != expected {
		t.Errorf("ExtractMarkdown() = %v, want %v", md, expected)
	}
}

// TestExtractMarkdownWithMultipleResults 测试多个文件结果的情况
func TestExtractMarkdownWithMultipleResults(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	result := map[string]interface{}{
		"backend": "pipeline",
		"version": "1.0.0",
		"results": map[string]interface{}{
			"document1.pdf": map[string]interface{}{
				"md_content": "# Document 1\n\nContent 1.",
			},
			"document2.pdf": map[string]interface{}{
				"md_content": "# Document 2\n\nContent 2.",
			},
		},
	}

	// 应该提取第一个有内容的文件
	md, err := client.ExtractMarkdown(result)
	if err != nil {
		t.Fatalf("ExtractMarkdown() error = %v", err)
	}

	// 应该返回第一个文件的 markdown 内容
	expected := "# Document 1\n\nContent 1."
	if md != expected {
		t.Errorf("ExtractMarkdown() = %v, want %v", md, expected)
	}
}

// TestExtractMarkdownWithEmptyContent 测试空内容的情况
func TestExtractMarkdownWithEmptyContent(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	result := map[string]interface{}{
		"backend": "pipeline",
		"version": "1.0.0",
		"results": map[string]interface{}{
			"empty.pdf": map[string]interface{}{
				"md_content": "",
			},
		},
	}

	_, err := client.ExtractMarkdown(result)
	if err == nil {
		t.Error("ExtractMarkdown() should return error for empty content")
	}
}

// TestExtractZipData 测试从结果中提取 ZIP 数据
func TestExtractZipData(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	zipData := []byte("fake zip data")
	result := map[string]interface{}{
		"response_format": "zip",
		"zip_data":        zipData,
		"content_type":    "application/zip",
	}

	extracted, err := client.ExtractZipData(result)
	if err != nil {
		t.Fatalf("ExtractZipData() error = %v", err)
	}

	if string(extracted) != string(zipData) {
		t.Errorf("ExtractZipData() = %v, want %v", extracted, zipData)
	}
}

// TestExtractZipDataError 测试提取 ZIP 数据失败的情况
func TestExtractZipDataError(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	result := map[string]interface{}{
		"response_format": "json",
	}

	_, err := client.ExtractZipData(result)
	if err == nil {
		t.Error("ExtractZipData() should return error when zip_data is missing")
	}
}

// TestExtractMarkdownFromZip 测试从 ZIP 文件中提取 Markdown
func TestExtractMarkdownFromZip(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	// 创建一个简单的 ZIP 文件
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// 添加一个 Markdown 文件
	mdFile, err := zipWriter.Create("test_document/test_document.md")
	if err != nil {
		t.Fatalf("创建 ZIP 文件失败: %v", err)
	}

	mdContent := "# Test Document\n\nThis is test content."
	_, err = mdFile.Write([]byte(mdContent))
	if err != nil {
		t.Fatalf("写入 ZIP 文件失败: %v", err)
	}

	err = zipWriter.Close()
	if err != nil {
		t.Fatalf("关闭 ZIP writer 失败: %v", err)
	}

	// 测试提取
	extracted, err := client.ExtractMarkdownFromZip(buf.Bytes(), "test_document")
	if err != nil {
		t.Fatalf("ExtractMarkdownFromZip() error = %v", err)
	}

	if extracted != mdContent {
		t.Errorf("ExtractMarkdownFromZip() = %v, want %v", extracted, mdContent)
	}
}

// TestExtractMarkdownFromZipNotFound 测试 ZIP 中找不到文件的情况
func TestExtractMarkdownFromZipNotFound(t *testing.T) {
	client := NewMinerUClient("http://127.0.0.1:8100")

	// 创建一个空的 ZIP 文件
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	zipWriter.Close()

	_, err := client.ExtractMarkdownFromZip(buf.Bytes(), "nonexistent")
	if err == nil {
		t.Error("ExtractMarkdownFromZip() should return error when file not found")
	}
}
