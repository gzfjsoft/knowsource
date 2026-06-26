package utils

import (
	"os"
	"testing"
)

// TestParseDocumentIntegration 集成测试示例
// 注意：这个测试需要 MinerU 服务运行在 http://127.0.0.1:8100
// 要运行此测试，请设置环境变量 RUN_INTEGRATION_TESTS=1
func TestParseDocumentIntegration(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("跳过集成测试，设置 RUN_INTEGRATION_TESTS=1 来运行")
	}

	// 创建客户端
	client := NewMinerUClient("http://127.0.0.1:8100")

	// 检查测试文件是否存在
	testFile := "test_document.pdf"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("测试文件 %s 不存在，跳过集成测试", testFile)
	}

	// 配置解析选项
	options := DefaultParseOptions()
	options.ReturnMD = true
	options.ReturnMiddleJSON = false
	options.ReturnModelOutput = false
	options.ReturnContentList = false
	options.ReturnImages = false

	// 调用 API
	result, err := client.ParseDocument(testFile, options)
	if err != nil {
		t.Fatalf("ParseDocument() error = %v", err)
	}

	// 验证结果不为空
	if result == nil {
		t.Fatal("ParseDocument() result should not be nil")
	}

	// 提取 Markdown
	md, err := client.ExtractMarkdown(result)
	if err != nil {
		t.Fatalf("ExtractMarkdown() error = %v", err)
	}

	// 验证 Markdown 内容不为空
	if md == "" {
		t.Error("ExtractMarkdown() should return non-empty markdown content")
	}

	t.Logf("成功提取 Markdown 内容，长度: %d 字符", len(md))
}
