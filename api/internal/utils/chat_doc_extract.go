package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	// AIChatDocCacheKeyPrefix Redis key prefix for AI chat temporary document cache (value = empCode)
	AIChatDocCacheKeyPrefix = "ai_chat_doc_cache:"
	// AIChatDocCacheTTLSeconds TTL for the cache (1 hour)
	AIChatDocCacheTTLSeconds = 3600
)

// ExtractOptions 文档识别可选配置，与 document/convert/to/zip 一致：PDF 用 MinerU，DOCX 用 pandoc
type ExtractOptions struct {
	MinerUURL  string // MinerU API 地址，如 http://127.0.0.1:8100
	FilesRoot  string // 文件根目录，用于路径解析（可为空）
}

// AIChatCachedDocItem 单条缓存的文档：文件名 + 识别出的文本
type AIChatCachedDocItem struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
}

// ExtractTextFromFile 从本地文件提取文本，支持 .txt、.pdf、.docx
// filePath 为已保存到磁盘的完整路径。
// 若 opts 不为空：PDF 使用 MinerU，DOCX 使用 pandoc（与 document/convert/to/zip 一致）；否则使用 go-fitz 兜底。
func ExtractTextFromFile(filePath string) (text string, err error) {
	return ExtractTextFromFileWithOptions(context.TODO(), filePath, nil)
}

// ExtractTextFromFileWithOptions 使用可选配置提取文本（MinerU / pandoc）
func ExtractTextFromFileWithOptions(ctx context.Context, filePath string, opts *ExtractOptions) (text string, err error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".txt":
		return extractTxt(filePath)
	case ".pdf":
		if opts != nil && opts.MinerUURL != "" {
			return extractPdfWithMinerU(ctx, filePath, opts.MinerUURL, opts.FilesRoot)
		}
		return extractPdfFitz(filePath)
	case ".docx":
		if opts != nil {
			return extractDocxWithPandoc(filePath)
		}
		return extractDocxFitz(filePath)
	default:
		return "", fmt.Errorf("不支持的文件类型: %s，仅支持 .txt、.pdf、.docx", ext)
	}
}

func extractTxt(filePath string) (string, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取 txt 失败: %w", err)
	}
	// 尝试 UTF-8
	return string(raw), nil
}

func extractPdfFitz(filePath string) (string, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return "", fmt.Errorf("打开 PDF 失败: %w", err)
	}
	defer doc.Close()

	var sb strings.Builder
	n := doc.NumPage()
	for i := 0; i < n; i++ {
		t, err := doc.Text(i)
		if err != nil {
			logx.WithContext(context.TODO()).Infof("提取 PDF 第 %d 页文本失败: %v", i+1, err)
			continue
		}
		if t != "" {
			sb.WriteString(t)
			if i < n-1 {
				sb.WriteString("\n\n")
			}
		}
	}
	return sb.String(), nil
}

// extractPdfWithMinerU 使用 MinerU 解析 PDF，与 document/convert/to/zip 一致
func extractPdfWithMinerU(ctx context.Context, filePath string, minerUURL string, filesRoot string) (string, error) {
	client := NewMinerUClient(strings.TrimSuffix(minerUURL, "/"))
	options := DefaultParseOptions()
	options.ReturnMD = true
	options.ReturnMiddleJSON = false
	options.ReturnModelOutput = false
	options.ReturnContentList = false
	options.ReturnImages = false
	options.ResponseFormatZip = false

	result, err := client.ParseDocument(filePath, options)
	if err != nil {
		return "", fmt.Errorf("MinerU 解析 PDF 失败: %w", err)
	}

	if format, ok := result["response_format"].(string); ok && format == "zip" {
		zipData, err := client.ExtractZipData(result)
		if err != nil {
			return "", fmt.Errorf("MinerU 提取 ZIP 失败: %v", err)
		}
		baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		return client.ExtractMarkdownFromZip(zipData, baseName)
	}
	return client.ExtractMarkdown(result)
}

// extractDocxWithPandoc 使用 pandoc（ConvertDocxToMd）解析 docx，与 document/convert/to/zip 一致
func extractDocxWithPandoc(filePath string) (string, error) {
	mdDir := filepath.Join(os.TempDir(), "ai_chat_docx_"+filepath.Base(filePath)+".tmp")
	defer os.RemoveAll(mdDir)

	mdPath, err := ConvertDocxToMd(filePath, mdDir)
	if err != nil {
		return "", fmt.Errorf("pandoc 转换 DOCX 失败: %w", err)
	}
	raw, err := os.ReadFile(mdPath)
	if err != nil {
		return "", fmt.Errorf("读取转换后的 MD 失败: %w", err)
	}
	return string(raw), nil
}

func extractDocxFitz(filePath string) (string, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return "", fmt.Errorf("打开 DOCX 失败: %w", err)
	}
	defer doc.Close()

	var sb strings.Builder
	n := doc.NumPage()
	for i := 0; i < n; i++ {
		t, err := doc.Text(i)
		if err != nil {
			logx.WithContext(context.TODO()).Infof("提取 DOCX 第 %d 页文本失败: %v", i+1, err)
			continue
		}
		if t != "" {
			sb.WriteString(t)
			if i < n-1 {
				sb.WriteString("\n\n")
			}
		}
	}
	return sb.String(), nil
}

// AllowedUploadExt 允许的上传扩展名
func AllowedUploadExt(ext string) bool {
	ext = strings.ToLower(ext)
	return ext == ".txt" || ext == ".pdf" || ext == ".docx"
}

// AIChatDocCacheGet 读取当前用户的 AI 对话文档缓存，返回 nil 表示无缓存
func AIChatDocCacheGet(rc *redis.Redis, empCode string) ([]AIChatCachedDocItem, error) {
	if rc == nil || empCode == "" {
		return nil, nil
	}
	key := AIChatDocCacheKeyPrefix + empCode
	val, err := rc.Get(key)
	if err != nil || val == "" {
		return nil, nil
	}
	var list []AIChatCachedDocItem
	if err := json.Unmarshal([]byte(val), &list); err != nil {
		return nil, err
	}
	return list, nil
}

// AIChatDocCacheSet 追加一条文档到缓存，并设置 TTL
func AIChatDocCacheSet(rc *redis.Redis, empCode string, item AIChatCachedDocItem) error {
	if rc == nil || empCode == "" {
		return fmt.Errorf("redis 或 empCode 为空")
	}
	key := AIChatDocCacheKeyPrefix + empCode
	list, _ := AIChatDocCacheGet(rc, empCode)
	if list == nil {
		list = []AIChatCachedDocItem{}
	}
	list = append(list, item)
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	return rc.Setex(key, string(data), AIChatDocCacheTTLSeconds)
}

// AIChatDocCacheClear 清除当前用户的 AI 对话文档缓存（发消息后调用）
func AIChatDocCacheClear(rc *redis.Redis, empCode string) error {
	if rc == nil || empCode == "" {
		return nil
	}
	key := AIChatDocCacheKeyPrefix + empCode
	_, err := rc.Del(key)
	return err
}

// AIChatDocCacheRemove 从缓存中移除指定文件名的文档
func AIChatDocCacheRemove(rc *redis.Redis, empCode string, filename string) error {
	if rc == nil || empCode == "" || filename == "" {
		return nil
	}
	key := AIChatDocCacheKeyPrefix + empCode
	list, err := AIChatDocCacheGet(rc, empCode)
	if err != nil || len(list) == 0 {
		return nil
	}
	var newList []AIChatCachedDocItem
	for _, item := range list {
		if item.Filename != filename {
			newList = append(newList, item)
		}
	}
	if len(newList) == 0 {
		_, err = rc.Del(key)
		return err
	}
	data, err := json.Marshal(newList)
	if err != nil {
		return err
	}
	return rc.Setex(key, string(data), AIChatDocCacheTTLSeconds)
}

// BuildContextFromCachedDocs 将缓存文档列表拼成发给 AI 的参考上下文
func BuildContextFromCachedDocs(list []AIChatCachedDocItem) string {
	if len(list) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("【以下为用户上传的文档内容，供参考】\n\n")
	for _, item := range list {
		sb.WriteString("--- 文档: ")
		sb.WriteString(item.Filename)
		sb.WriteString(" ---\n\n")
		sb.WriteString(item.Content)
		sb.WriteString("\n\n")
	}
	sb.WriteString("【用户问题】\n\n")
	return sb.String()
}
