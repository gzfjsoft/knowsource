package utils

import (
	"context"
	"fmt"
	"knowsource/model"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// DocumentConverter 文档转换服务
type DocumentConverter struct {
	MinerUClient *MinerUClient
	MinerUURL    string
	FilesRoot    string
}

// NewDocumentConverter 创建新的文档转换服务
func NewDocumentConverter(minerUURL string, filesRoot string) *DocumentConverter {
	return &DocumentConverter{
		MinerUClient: NewMinerUClient(minerUURL),
		MinerUURL:    minerUURL,
		FilesRoot:    filesRoot,
	}
}

// ConvertDocumentToMD 将文档转换为 Markdown
// 读取 rawdoc，调用 minerU 转换为 md，并保存到数据库
func (dc *DocumentConverter) ConvertDocumentToMD(ctx context.Context, rawDoc *model.RawDocuments) error {
	// 确定文件路径
	// FilePath 在数据库中可能存储的是完整路径或相对路径
	// 先尝试直接使用 FilePath，如果文件不存在，再尝试拼接 FilesRoot

	var filePath string
	var err error

	// 先尝试直接使用 FilePath（可能是完整路径）
	filePath = filepath.Clean(rawDoc.FilePath)
	if _, err = os.Stat(filePath); err == nil {
		// 文件存在，使用这个路径
	} else {
		// 文件不存在，尝试拼接 FilesRoot
		filePath = filepath.Clean(filepath.Join(dc.FilesRoot, rawDoc.FilePath))
		if _, err = os.Stat(filePath); err != nil {
			// 两种方式都失败，返回错误
			return fmt.Errorf("文件不存在: 尝试路径1=%s, 尝试路径2=%s (FilesRoot=%s, FilePath=%s)",
				filepath.Clean(rawDoc.FilePath), filePath, dc.FilesRoot, rawDoc.FilePath)
		}
	}

	// 调用 MinerU 转换文档
	options := DefaultParseOptions()
	options.ReturnMD = true
	options.ReturnMiddleJSON = false
	options.ReturnModelOutput = false
	options.ReturnContentList = false
	options.ReturnImages = false
	options.ResponseFormatZip = false // 默认返回 JSON，如果需要 zip 可以设置为 true

	result, err := dc.MinerUClient.ParseDocument(filePath, options)
	if err != nil {
		return fmt.Errorf("调用 MinerU 转换失败: %v", err)
	}

	// 检查响应格式
	var mdContent string
	if format, ok := result["response_format"].(string); ok && format == "zip" {
		// 从 ZIP 文件中提取 Markdown
		zipData, err := dc.MinerUClient.ExtractZipData(result)
		if err != nil {
			return fmt.Errorf("提取 ZIP 数据失败: %v", err)
		}

		// 从文件名中提取基础名称（不含扩展名）
		baseName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		mdContent, err = dc.MinerUClient.ExtractMarkdownFromZip(zipData, baseName)
		if err != nil {
			return fmt.Errorf("从 ZIP 中提取 Markdown 失败: %v", err)
		}
	} else {
		// 从 JSON 响应中提取 Markdown 内容
		mdContent, err = dc.MinerUClient.ExtractMarkdown(result)
		if err != nil {
			return fmt.Errorf("提取 Markdown 内容失败: %v", err)
		}
	}

	// 更新数据库：保存 Markdown 内容并标记为已转换
	// 同时更新 content 和 content_org（首次转换时两者相同）
	rawDoc.Content = mdContent
	rawDoc.ContentOrg = mdContent
	rawDoc.IsToMd = 1

	return nil
}

// BatchConvertDocuments 批量转换文档
func (dc *DocumentConverter) BatchConvertDocuments(ctx context.Context, rawDocs []*model.RawDocuments) (successCount int, failCount int, errors []error) {
	successCount = 0
	failCount = 0
	errors = []error{}

	for _, rawDoc := range rawDocs {
		if rawDoc.IsToMd == 1 {
			// 已经转换过，跳过
			continue
		}

		err := dc.ConvertDocumentToMD(ctx, rawDoc)
		if err != nil {
			failCount++
			errors = append(errors, fmt.Errorf("文档 %s (ID: %d) 转换失败: %v", rawDoc.FileName, rawDoc.Id, err))
			logx.WithContext(ctx).Errorf("文档转换失败: %v", err)
		} else {
			successCount++
			logx.WithContext(ctx).Infof("文档转换成功: %s (ID: %d)", rawDoc.FileName, rawDoc.Id)
		}
	}

	return successCount, failCount, errors
}

// DocumentConvertDownloader 将文档转换为 ZIP 格式并保存到文件
// filePath: 要转换的文档路径
// outputZipPath: 输出 ZIP 文件的保存路径
// 返回保存的 ZIP 文件路径和错误
func (dc *DocumentConverter) DocumentConvertDownloader(ctx context.Context, filePath string, outputZipPath string) (string, error) {
	// 确定文件路径
	// 先尝试直接使用 filePath，如果文件不存在，再尝试拼接 FilesRoot
	var actualFilePath string
	var err error

	// 先尝试直接使用 filePath（可能是完整路径）
	actualFilePath = filepath.Clean(filePath)
	if _, err = os.Stat(actualFilePath); err == nil {
		// 文件存在，使用这个路径
	} else {
		// 文件不存在，尝试拼接 FilesRoot
		actualFilePath = filepath.Clean(filepath.Join(dc.FilesRoot, filePath))
		if _, err = os.Stat(actualFilePath); err != nil {
			// 两种方式都失败，返回错误
			return "", fmt.Errorf("文件不存在: 尝试路径1=%s, 尝试路径2=%s (FilesRoot=%s, FilePath=%s)",
				filepath.Clean(filePath), actualFilePath, dc.FilesRoot, filePath)
		}
	}

	// 调用 MinerU 转换文档，设置 ResponseFormatZip=true
	options := DefaultParseOptions()
	options.ReturnMD = true
	options.ReturnMiddleJSON = false
	options.ReturnModelOutput = false
	options.ReturnContentList = false
	options.ReturnImages = true
	options.ResponseFormatZip = true // 设置为 true 以获取 ZIP 格式响应

	result, err := dc.MinerUClient.ParseDocument(actualFilePath, options)
	if err != nil {
		return "", fmt.Errorf("调用 MinerU 转换失败: %v", err)
	}

	// 检查响应格式，确保是 ZIP 格式
	var zipData []byte
	if format, ok := result["response_format"].(string); ok && format == "zip" {
		// 从结果中提取 ZIP 数据
		zipData, err = dc.MinerUClient.ExtractZipData(result)
		if err != nil {
			return "", fmt.Errorf("提取 ZIP 数据失败: %v", err)
		}
	} else {
		return "", fmt.Errorf("响应格式不是 ZIP，实际格式: %v", result["response_format"])
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputZipPath)
	if outputDir != "" && outputDir != "." {
		if err = os.MkdirAll(outputDir, 0755); err != nil {
			return "", fmt.Errorf("创建输出目录失败: %v", err)
		}
	}

	// 保存 ZIP 文件
	if err = os.WriteFile(outputZipPath, zipData, 0644); err != nil {
		return "", fmt.Errorf("保存 ZIP 文件失败: %v", err)
	}

	return outputZipPath, nil
}

// ConvertDocumentToZIP 将文档转换为 ZIP 格式并保存到文件
// 读取 rawdoc，调用 minerU 转换为 zip，并保存到文件系统
// 返回保存的 ZIP 文件路径
func (dc *DocumentConverter) ConvertDocumentToZIP(ctx context.Context, rawDoc *model.RawDocuments, outputZipPath string) (string, error) {
	// 确定文件路径
	// FilePath 在数据库中可能存储的是完整路径或相对路径
	// 先尝试直接使用 FilePath，如果文件不存在，再尝试拼接 FilesRoot

	var filePath string
	var err error

	// 先尝试直接使用 FilePath（可能是完整路径）
	filePath = filepath.Clean(rawDoc.FilePath)
	if _, err = os.Stat(filePath); err == nil {
		// 文件存在，使用这个路径
	} else {
		// 文件不存在，尝试拼接 FilesRoot
		filePath = filepath.Clean(filepath.Join(dc.FilesRoot, rawDoc.FilePath))
		if _, err = os.Stat(filePath); err != nil {
			// 两种方式都失败，返回错误
			return "", fmt.Errorf("文件不存在: 尝试路径1=%s, 尝试路径2=%s (FilesRoot=%s, FilePath=%s)",
				filepath.Clean(rawDoc.FilePath), filePath, dc.FilesRoot, rawDoc.FilePath)
		}
	}

	// 调用 DocumentConvertDownloader 下载并保存 ZIP 文件
	zipPath, err := dc.DocumentConvertDownloader(ctx, filePath, outputZipPath)
	if err != nil {
		return "", fmt.Errorf("转换并下载 ZIP 文件失败: %v", err)
	}

	return zipPath, nil
}
