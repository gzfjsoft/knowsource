package utils

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// MinerUClient MinerU API 客户端
type MinerUClient struct {
	BaseURL string
	Client  *http.Client
}

// ParseOptions MinerU 解析选项
type ParseOptions struct {
	OutputDir         string
	LangList          []string // 语言列表，如 ["ch"]
	Backend           string
	ParseMethod       string
	FormulaEnable     bool
	TableEnable       bool
	ServerURL         string
	ReturnMD          bool
	ReturnMiddleJSON  bool
	ReturnModelOutput bool
	ReturnContentList bool
	ReturnImages      bool
	ResponseFormatZip bool
	StartPageID       int
	EndPageID         int
}

// DefaultParseOptions 返回默认的解析选项
func DefaultParseOptions() *ParseOptions {
	return &ParseOptions{
		OutputDir:         "./output",
		LangList:          []string{"ch"},
		Backend:           "pipeline",
		ParseMethod:       "auto",
		FormulaEnable:     true,
		TableEnable:       true,
		ServerURL:         "",
		ReturnMD:          true,
		ReturnMiddleJSON:  false,
		ReturnModelOutput: false,
		ReturnContentList: false,
		ReturnImages:      false,
		ResponseFormatZip: false,
		StartPageID:       0,
		EndPageID:         99999,
	}
}

// NewMinerUClient 创建新的 MinerU 客户端
func NewMinerUClient(baseURL string) *MinerUClient {
	return &MinerUClient{
		BaseURL: baseURL,
		Client: &http.Client{
			Timeout: 600 * time.Second, // 10分钟超时，PDF解析可能需要较长时间
		},
	}
}

// ParseDocument 解析文档（调用 MinerU 的 /file_parse 接口）
func (mc *MinerUClient) ParseDocument(filePath string, options *ParseOptions) (map[string]interface{}, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	// 创建multipart writer
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// 添加文件（参数名是 "files"）
	fileField, err := writer.CreateFormFile("files", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("创建文件字段失败: %v", err)
	}

	_, err = io.Copy(fileField, file)
	if err != nil {
		return nil, fmt.Errorf("复制文件内容失败: %v", err)
	}

	// 添加可选参数
	if options != nil {
		if options.OutputDir != "" {
			writer.WriteField("output_dir", options.OutputDir)
		}
		if len(options.LangList) > 0 {
			for _, lang := range options.LangList {
				writer.WriteField("lang_list", lang)
			}
		}
		if options.Backend != "" {
			writer.WriteField("backend", options.Backend)
		}
		if options.ParseMethod != "" {
			writer.WriteField("parse_method", options.ParseMethod)
		}
		writer.WriteField("formula_enable", strconv.FormatBool(options.FormulaEnable))
		writer.WriteField("table_enable", strconv.FormatBool(options.TableEnable))
		if options.ServerURL != "" {
			writer.WriteField("server_url", options.ServerURL)
		}
		writer.WriteField("return_md", strconv.FormatBool(options.ReturnMD))
		writer.WriteField("return_middle_json", strconv.FormatBool(options.ReturnMiddleJSON))
		writer.WriteField("return_model_output", strconv.FormatBool(options.ReturnModelOutput))
		writer.WriteField("return_content_list", strconv.FormatBool(options.ReturnContentList))
		writer.WriteField("return_images", strconv.FormatBool(options.ReturnImages))
		writer.WriteField("response_format_zip", strconv.FormatBool(options.ResponseFormatZip))
		writer.WriteField("start_page_id", strconv.Itoa(options.StartPageID))
		writer.WriteField("end_page_id", strconv.Itoa(options.EndPageID))
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("关闭multipart writer失败: %v", err)
	}

	// 创建请求
	parseURL := strings.TrimSuffix(mc.BaseURL, "/") + "/file_parse"
	req, err := http.NewRequest("POST", parseURL, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := mc.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("解析失败，状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
	}

	// 检查响应类型
	contentType := resp.Header.Get("Content-Type")

	// 如果是 zip 文件响应
	if contentType == "application/zip" || strings.Contains(contentType, "zip") {
		return map[string]interface{}{
			"response_format": "zip",
			"zip_data":        respBody,
			"content_type":    contentType,
		}, nil
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// 如果响应不是JSON，直接返回原始内容
		return map[string]interface{}{
			"raw_response": string(respBody),
		}, nil
	}

	return result, nil
}

// ExtractMarkdown 从解析结果中提取 Markdown 内容
// MinerU API 返回格式：
//
//	{
//	  "backend": "...",
//	  "version": "...",
//	  "results": {
//	    "pdf_file_name": {
//	      "md_content": "...",
//	      ...
//	    }
//	  }
//	}
func (mc *MinerUClient) ExtractMarkdown(result map[string]interface{}) (string, error) {
	// 首先尝试从 results 中提取
	if results, ok := result["results"].(map[string]interface{}); ok {
		// 遍历 results 中的每个文件
		for fileName, fileData := range results {
			if fileDataMap, ok := fileData.(map[string]interface{}); ok {
				// 尝试提取 md_content
				if mdContent, ok := fileDataMap["md_content"].(string); ok && mdContent != "" {
					return mdContent, nil
				}
				// 如果没有 md_content，尝试其他可能的字段
				if md, ok := fileDataMap["md"].(string); ok && md != "" {
					return md, nil
				}
				if md, ok := fileDataMap["markdown"].(string); ok && md != "" {
					return md, nil
				}
				if md, ok := fileDataMap["content"].(string); ok && md != "" {
					return md, nil
				}
				// 如果找到了文件数据但都没有 markdown 内容，记录文件名
				return "", fmt.Errorf("文件 %s 的解析结果中没有 Markdown 内容", fileName)
			}
		}
		return "", fmt.Errorf("results 中没有找到有效的文件数据")
	}

	// 向后兼容：尝试直接从顶层提取
	if md, ok := result["md"].(string); ok && md != "" {
		return md, nil
	}
	if md, ok := result["md_content"].(string); ok && md != "" {
		return md, nil
	}
	if md, ok := result["markdown"].(string); ok && md != "" {
		return md, nil
	}
	if md, ok := result["content"].(string); ok && md != "" {
		return md, nil
	}

	// 如果都没有，返回错误
	return "", fmt.Errorf("无法从解析结果中提取 Markdown 内容，结果结构: %+v", result)
}

// ExtractZipData 从解析结果中提取 ZIP 文件数据
func (mc *MinerUClient) ExtractZipData(result map[string]interface{}) ([]byte, error) {
	if zipData, ok := result["zip_data"].([]byte); ok {
		return zipData, nil
	}
	return nil, fmt.Errorf("解析结果中没有 ZIP 数据")
}

// ExtractMarkdownFromZip 从 ZIP 文件中提取 Markdown 内容
// zipData: ZIP 文件的字节数据
// fileName: 要查找的文件名（不含扩展名），如 "世界是什么"
func (mc *MinerUClient) ExtractMarkdownFromZip(zipData []byte, fileName string) (string, error) {
	// 创建 ZIP 读取器
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return "", fmt.Errorf("创建 ZIP 读取器失败: %v", err)
	}

	// 查找 Markdown 文件
	mdFileName := fileName + ".md"
	for _, file := range zipReader.File {
		// 检查文件名（可能在子目录中，如 "世界是什么/世界是什么.md"）
		if strings.HasSuffix(file.Name, mdFileName) || strings.Contains(file.Name, mdFileName) {
			// 打开文件
			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("打开 ZIP 中的文件失败: %v", err)
			}
			defer rc.Close()

			// 读取内容
			content, err := io.ReadAll(rc)
			if err != nil {
				return "", fmt.Errorf("读取 ZIP 中的文件内容失败: %v", err)
			}

			return string(content), nil
		}
	}

	return "", fmt.Errorf("在 ZIP 文件中未找到 %s", mdFileName)
}
