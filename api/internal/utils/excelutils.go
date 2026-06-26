package utils

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

func ReadExcelDataFromFile(file io.Reader, notMerged bool) ([][]string, error) {
	f, err := excelize.OpenReader(file)
	if err != nil {
		fmt.Printf("打开Excel文件失败: %v\n", err)
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("关闭Excel文件失败: %v\n", err)
		}
	}()
	return GetExcelFirstSheetRows(f, notMerged)
}

func ReadExcelData(filename string, notMerged bool) ([][]string, error) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		fmt.Printf("打开Excel文件失败: %v\n", err)
		return nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("关闭Excel文件失败: %v\n", err)
		}
	}()
	return GetExcelFirstSheetRows(f, notMerged)
}

func GetExcelFirstSheetRows(f *excelize.File, notMerged bool) ([][]string, error) {
	// 获取第一个工作表名称
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, errors.New("Excel文件中没有工作表")
	}
	if notMerged {
		// 1. 先获取合并单元格信息
		mergedCells, err := f.GetMergeCells(sheetName)
		if err != nil {
			return nil, fmt.Errorf("获取合并单元格信息失败: %v", err)
		}

		if len(mergedCells) > 0 {

			arr := make([]string, 0)
			for _, mergedCell := range mergedCells {
				arr = append(arr, fmt.Sprintf("%s:%s", mergedCell.GetStartAxis(), mergedCell.GetEndAxis()))
			}

			return nil, fmt.Errorf("导入Excel数据不允许有合并单元格的格式，请导入前先处理掉合并单元格。合并单元格信息如下:[ %s ]", strings.Join(arr, ","))
		}
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("读取Excel工作表数据失败: %v", err)
	}

	return rows, nil
}

// 删除空行
func RemoveEmptyRows(rows [][]string) [][]string {
	var filteredRows [][]string
	for _, row := range rows {
		if !isEmptyRow(row) {
			filteredRows = append(filteredRows, row)
		}
	}
	return filteredRows
}

func isEmptyRow(row []string) bool {
	if row == nil {
		return true
	}

	columnLen := len(row)
	if columnLen > 4 {
		columnLen = 4
	}
	// 过滤空行
	isEmptyRow := true
	// 如果前4个字段都为空，则认为是空行
	for j := 0; j < columnLen; j++ {
		if row[j] != "" {
			isEmptyRow = false
			break
		}
	}
	return isEmptyRow
}

// 支持的编码格式
var encodings = []struct {
	name string
	enc  encoding.Encoding
}{
	{"UTF-8", nil}, // UTF-8不需要转换
	{"GBK", simplifiedchinese.GBK},
	{"GB18030", simplifiedchinese.GB18030},
	{"Big5", traditionalchinese.Big5},
	{"Shift_JIS", japanese.ShiftJIS},
	{"EUC-JP", japanese.EUCJP},
	{"EUC-KR", korean.EUCKR},
}

// RowData 表示一行数据
type RowData struct {
	RowIndex int            `json:"row_index"`
	Data     map[string]any `json:"data"`
}

// readCSVFile 读取CSV文件
func ReadCsvData(filePath string, autoEncoding bool) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 检测编码
	var reader io.Reader = file
	if autoEncoding {
		detectedEncoding, err := detectEncoding(filePath)
		if err != nil {
			logx.Infof("编码检测失败，使用默认编码: %v", err)
			detectedEncoding = simplifiedchinese.GBK
		}

		if detectedEncoding != nil {
			reader = transform.NewReader(file, detectedEncoding.NewDecoder())
		}
		if detectedEncoding != nil {
			reader = transform.NewReader(file, detectedEncoding.NewDecoder())
			logx.Infof("使用编码: %v", detectedEncoding)
		} else {
			logx.Infof("使用编码: UTF-8")
		}
	}

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

// detectEncoding 检测文件编码
func detectEncoding(filePath string) (encoding.Encoding, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 读取前4KB来检测编码
	buffer := make([]byte, 4096)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	buffer = buffer[:n]

	// 检查是否是UTF-8
	if utf8.Valid(buffer) {
		return nil, nil // UTF-8
	}

	// 尝试其他编码
	for _, enc := range encodings[1:] { // 跳过UTF-8
		reader := transform.NewReader(strings.NewReader(string(buffer)), enc.enc.NewDecoder())
		decoded, err := io.ReadAll(reader)
		if err == nil && utf8.Valid(decoded) {
			return enc.enc, nil
		}
	}

	// 默认返回GBK
	return simplifiedchinese.GBK, nil
}

// 将二维数组转换为CSV文件
func WriteCSVFile(filePath string, data [][]string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range data {
		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil
}

func FormatTime(t time.Time, format string) string {
	emptyTime, _ := time.Parse("2006-01-02 15:04:05", "2000-01-01 00:00:00")

	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	// 如果时间小于等于空时间，则返回空字符串
	if t.Before(emptyTime) {
		return ""
	}

	result := t.Format(format)
	return result
}

// ExportToExcel 导出数据到Excel文件
func ExportToExcel(data [][]interface{}, headers []string, filename string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("关闭Excel文件失败: %v\n", err)
		}
	}()

	// 设置工作表名称
	sheetName := "Sheet1"
	f.SetSheetName("Sheet1", sheetName)

	// 写入表头
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
	}

	// 写入数据
	for i, row := range data {
		rowNum := i + 2 // 从第2行开始（第1行是表头）
		for j, cellData := range row {
			cell := fmt.Sprintf("%s%d", string(rune('A'+j)), rowNum)
			f.SetCellValue(sheetName, cell, cellData)
		}
	}

	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("保存Excel文件失败: %v", err)
	}

	return nil
}
