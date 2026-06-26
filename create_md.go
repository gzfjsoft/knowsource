package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

// readCSVFile 读取CSV文件
func readCSVFile(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var reader io.Reader = file

	// 创建CSV读取器
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.FieldsPerRecord = -1 // 允许不同长度的行

	// 读取所有行
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取CSV文件失败: %v", err)
	}

	return records, nil
}

func CsvToMarkdownWithRoot(inputFile, outputDir, rootDir string) error {
	// 读取CSV文件
	records, err := readCSVFile(inputFile)
	if err != nil {
		return fmt.Errorf("读取CSV文件失败: %v", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV文件为空")
	}

	// 获取输入文件名（去掉路径和扩展名）
	baseName := filepath.Base(inputFile)
	fileNameWithoutExt := strings.TrimSuffix(baseName, ".csv")
	if fileNameWithoutExt == baseName {
		fileNameWithoutExt = baseName + "_输出"
	}

	// 创建输出目录：保持源文件的目录结构
	// 计算相对于根目录的路径
	relPath, err := filepath.Rel(rootDir, inputFile)
	if err != nil {
		relPath = baseName
	}

	// 获取目录部分（去掉文件名）
	relDir := filepath.Dir(relPath)
	if relDir == "." {
		relDir = ""
	}

	// 创建完整的输出目录路径（不包含文件名子文件夹）
	var finalOutputDir string
	if relDir != "" {
		finalOutputDir = filepath.Join(outputDir, relDir)
	} else {
		finalOutputDir = outputDir
	}

	if err := os.MkdirAll(finalOutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 获取表头
	headers := records[0]
	if len(headers) == 0 {
		return fmt.Errorf("CSV文件没有表头")
	}

	fmt.Printf("正在处理CSV文件: %s\n", inputFile)

	// 处理每一行数据（跳过表头）
	for i, record := range records[1:] {
		if len(record) == 0 {
			continue
		}

		// 创建行数据
		rowData := RowData{
			RowIndex: i + 2, // 行号从2开始（跳过表头）
			Data:     make(map[string]interface{}),
		}

		// 填充数据
		for j, cell := range record {
			if j < len(headers) {
				rowData.Data[headers[j]] = cell
			}
		}

		// 检查是否有数据
		if len(rowData.Data) == 0 {
			continue
		}

		// 生成markdown文件名：文件名--<行号>.csv
		fileName := fmt.Sprintf("%s--%d.csv.md", fileNameWithoutExt, i+2)
		filePath := filepath.Join(finalOutputDir, fileName)

		// 生成markdown内容
		markdownContent := generateMarkdown("CSV", rowData, headers, inputFile, outputDir, fileName)

		// 写入文件
		if err := os.WriteFile(filePath, []byte(markdownContent), 0644); err != nil {
			logx.Infof("写入文件 %s 失败: %v", filePath, err)
			continue
		}

		fmt.Printf("已生成: %s\n", filePath)
	}

	return nil
}

func main() {
	csvFile := "./files/knowledge/风机诊断/东气风电/2.0MV/风机诊断_东气风电_2.0MV.csv"
	// outPath := "./files/markdown"
	rootPath := "./files/knowledge"

	path := filepath.Dir(csvFile)
	fmt.Println(path)

	path, err := filepath.Rel(rootPath, path)
	if err != nil {
		fmt.Errorf("转换失败: %v", err)
	}

	fmt.Println(path)

	// err := CsvToMarkdownWithRoot(csvFile, outPath, rootPath)
	// if err != nil {
	// 	fmt.Errorf("转换失败: %v", err)
	// }
}

// RowData 表示一行数据
type RowData struct {
	RowIndex int                    `json:"row_index"`
	Data     map[string]interface{} `json:"data"`
}

// generateMarkdown 生成markdown内容
func generateMarkdown(sheetName string, rowData RowData, headers []string, inputFile string, outputDir string, fileName string) string {
	var content strings.Builder

	// 添加文件路径信息

	// 计算相对路径，去掉文件名，只显示目录路径
	relPath, err := filepath.Rel(outputDir, inputFile)
	if err != nil {
		// 如果无法计算相对路径，使用绝对路径
		relPath = inputFile
	}

	// 去掉文件名，只保留目录路径
	relDir := filepath.Dir(relPath)
	if relDir == "." {
		relDir = "当前目录"
	}

	type DocInfo struct {
		SourceDir   string
		CurrentFile string
		SourceFile  string
		RowIndex    int
		Attachment  []string `json:"attachment,omitempty"`
	}

	var docInfo DocInfo
	// content.WriteString(fmt.Sprintf("**工作表名称:** `%s`\n\n", sheetName))
	// content.WriteString(fmt.Sprintf("**数据行号:** `%d`\n\n", rowData.RowIndex))

	// 数据详情 - 使用简洁的字段名: 字段值格式
	// content.WriteString("## 数据详情\n\n")

	// 遍历所有数据，包括空字段名的值
	for key, value := range rowData.Data {
		if value == nil {
			continue
		}

		// 跳过空值
		if fmt.Sprintf("%v", value) == "" {
			continue
		}

		// 处理字段名
		fieldName := key
		fieldName = strings.ReplaceAll(fieldName, "\n", "")
		if strings.TrimSpace(fieldName) == "" {
			fieldName = "内容"
		}

		if HasString(fieldName, []string{"附件"}) { ////, "图片", "照片", "视频", "音频", "文件", "图纸"
			valueStr, ok := value.(string)
			if ok {
				// valueStr = strings.ReplaceAll(valueStr, "、", ";")
				// valueStr = strings.ReplaceAll(valueStr, "，", ";")
				// valueStr = strings.ReplaceAll(valueStr, ",", ";")
				valueStr = strings.ReplaceAll(valueStr, "\\", ";")
				valueStr = strings.ReplaceAll(valueStr, "\n", ";")
			}

			Values := strings.Split(valueStr, ";")

			// 去重处理，确保非空串和TrimSpace
			for _, val := range Values {
				cleanedVal := cleanAttachmentString(val)
				// 确保非空字符串且不重复
				if isValidAttachment(cleanedVal) && !contains(docInfo.Attachment, cleanedVal) {
					docInfo.Attachment = append(docInfo.Attachment, cleanedVal)
				}
			}

		}

		// 使用粗体字段名，后面跟冒号和值
		content.WriteString(fmt.Sprintf("### %s:\n%v\n\n", fieldName, value))

	}

	// 打印DocInfo
	inputFileWithoutPath := filepath.Base(inputFile)

	docInfo.SourceDir = relDir
	docInfo.CurrentFile = fileName
	docInfo.SourceFile = inputFileWithoutPath
	docInfo.RowIndex = rowData.RowIndex

	JsonStr, _ := json.Marshal(docInfo)
	content.WriteString("### DocInfo:\n")
	content.WriteString(fmt.Sprintf("%s\n\n", string(JsonStr)))

	// JSON格式数据
	if false {
		content.WriteString("## JSON格式\n\n")
		content.WriteString("```json\n")
		jsonData, _ := json.MarshalIndent(rowData, "", "  ")
		content.WriteString(string(jsonData))
		content.WriteString("\n```\n")
	}
	return content.String()
}

func HasString(str string, substr []string) bool {
	for _, s := range substr {
		if strings.Contains(strings.ToLower(str), strings.ToLower(s)) {
			return true
		}
	}
	return false
}

// contains 检查切片中是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// cleanAttachmentString 清理附件字符串，确保非空且已TrimSpace
func cleanAttachmentString(s string) string {
	return strings.TrimSpace(s)
}

// isValidAttachment 检查附件字符串是否有效（非空且已清理）
func isValidAttachment(s string) bool {
	cleaned := cleanAttachmentString(s)
	return cleaned != ""
}
