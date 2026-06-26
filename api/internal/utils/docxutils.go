package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/zeromicro/go-zero/core/logx"
)

// ConvertXlsxToMd 将 xlsx 文件转换为 md 文件
// 每一行内容转为文字，每一列换行，每一行内容之间用 "<AI分隔符>" 分割
func ConvertXlsxToMd(xlsxPath, mdPath string) error {
	// 读取 Excel 文件
	f, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %v", err)
	}
	defer f.Close()

	// 获取第一个工作表
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return fmt.Errorf("Excel文件中没有工作表")
	}

	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return fmt.Errorf("读取Excel工作表数据失败: %v", err)
	}

	// 构建 Markdown 内容
	var content strings.Builder

	// 第一行作为标题
	if len(rows) == 0 {
		return fmt.Errorf("Excel文件中没有数据")
	}

	headers := rows[0]
	// 清理标题，去除空标题
	var validHeaders []string
	var headerIndices []int
	for i, header := range headers {
		headerText := strings.TrimSpace(header)
		if headerText != "" {
			validHeaders = append(validHeaders, headerText)
			headerIndices = append(headerIndices, i)
		}
	}

	if len(validHeaders) == 0 {
		return fmt.Errorf("Excel文件中没有有效的标题行")
	}

	// 处理数据行（从第二行开始）
	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		if rowIdx > 1 {
			// 在行之间添加分隔符（第一行数据之前不添加）
			content.WriteString("<AI分隔符>\n")
		}

		row := rows[rowIdx]

		// 处理每一列，使用标题
		for i, headerIdx := range headerIndices {
			header := validHeaders[i]

			// 获取对应的单元格值
			var cellValue string
			if headerIdx < len(row) {
				cellValue = strings.TrimSpace(row[headerIdx])
			}
			if cellValue == "" {
				cellValue = " " // 空单元格用空格表示
			}

			// 格式：<标题>：<值>
			content.WriteString(header)
			content.WriteString("：")
			content.WriteString(cellValue)
			content.WriteString("\n")
		}
	}

	// 确保目标目录存在
	dir := filepath.Dir(mdPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 保存 Markdown 文件
	if err := os.WriteFile(mdPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("保存Markdown文件失败: %v", err)
	}

	return nil
}

//1 . file.doc 就先 转为 file.doc.file/file.docx
//   然后进入   file.doc.file ，pandoc file.docx -o file.md --extract-media=. --to=gfm
//2. file.docx 就直接进入 file.docx.file , pandoc ../file.docx -o file.md --extract-media=. --to=gfm

// ConvertDocxToMd 使用 pandoc 将 doc/docx 文件转换为 md 文件
// docPath: 源 doc/docx 文件路径
// mdDir: 输出目录路径（md 文件将保存在此目录中）
// 返回生成的 md 文件路径和错误
func ConvertDocxToMd(docPath, mdDir string) (string, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(mdDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %v", err)
	}

	// 获取文件名（不含扩展名）
	baseName := filepath.Base(docPath)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	fileExt := strings.ToLower(filepath.Ext(docPath))

	// 获取绝对路径
	absDocPath, err := filepath.Abs(docPath)
	if err != nil {
		return "", fmt.Errorf("获取绝对路径失败: %v", err)
	}

	// 如果是 doc 文件，需要先转换为 docx（转换到 mdDir 目录）
	var actualFilePath string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 获取 mdDir 的绝对路径（用于 libreoffice 命令）
	absMdDir, err := filepath.Abs(mdDir)
	if err != nil {
		return "", fmt.Errorf("获取输出目录绝对路径失败: %v", err)
	}

	if fileExt == ".doc" {
		logx.Infof("检测到 doc 文件，先转换为 docx 到目录: %s", absMdDir)
		// 使用 LibreOffice 将 doc 转换到 mdDir 目录
		// 方式1：使用 --outdir 参数（原来的方式）
		// libreoffice --headless --convert-to docx --outdir <output_dir> <input_file>
		cmd := exec.CommandContext(ctx, "libreoffice", "--headless", "--convert-to", "docx", "--outdir", absMdDir, absDocPath)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		logx.Infof("执行命令（方式1）: libreoffice --headless --convert-to docx --outdir %s %s", absMdDir, absDocPath)
		runErr := cmd.Run()
		stdoutStr := stdout.String()
		stderrStr := stderr.String()

		errInfo := stdoutStr + "\n" + stderrStr
		if stdoutStr != "" {
			logx.Infof("LibreOffice 标准输出: %s", stdoutStr)
		}
		if stderrStr != "" {
			logx.Infof("LibreOffice 标准错误: %s", stderrStr)
		}

		libreOfficeSuccess := false
		if runErr != nil {
			// 方式1失败，尝试方式2：使用 CombinedOutput（类似用户提供的示例）
			logx.Infof("LibreOffice 方式1转换失败: %v, 标准输出: %s, 标准错误: %s", runErr, stdoutStr, stderrStr)
			logx.Infof("尝试 LibreOffice 方式2（使用 CombinedOutput）")

			args := []string{
				"--headless",
				"--convert-to",
				"docx",
				"--outdir",
				absMdDir,
				absDocPath,
			}
			cmd2 := exec.CommandContext(ctx, "libreoffice", args...)
			output, err := cmd2.CombinedOutput()
			if err != nil {
				logx.Errorf("LibreOffice 方式2也失败: %v, 输出信息: %s", err, string(output))
				// 如果两种方式都失败，尝试使用 unoconv
				logx.Infof("尝试使用 unoconv 转换")
			} else {
				logx.Infof("LibreOffice 方式2转换成功，输出信息: %s", string(output))
				libreOfficeSuccess = true
			}
		} else {
			libreOfficeSuccess = true
		}

		if !libreOfficeSuccess {
			// 如果 LibreOffice 两种方式都失败，尝试使用 unoconv
			// unoconv 需要指定输出文件路径
			tempDocxPath := filepath.Join(absMdDir, nameWithoutExt+".docx")
			cmd = exec.CommandContext(ctx, "unoconv", "-f", "docx", "-o", tempDocxPath, absDocPath)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("doc 转 docx 失败 (LibreOffice 和 unoconv 都失败): LibreOffice错误=%v,LibreOffice错误信息=%s, unoconv错误=%v, 错误输出: %s", runErr, errInfo, err, stderr.String())
			}
			actualFilePath = tempDocxPath
			logx.Infof("unoconv 转换成功，文件路径: %s", actualFilePath)
		}

		// 如果 LibreOffice 成功（方式1或方式2），查找生成的 docx 文件
		if libreOfficeSuccess {
			// LibreOffice 转换成功，查找生成的 docx 文件
			// LibreOffice 通常会在输出目录生成与输入文件同名的 docx 文件
			expectedDocxPath := filepath.Join(absMdDir, nameWithoutExt+".docx")
			logx.Infof("检查预期的 docx 文件路径: %s", expectedDocxPath)
			if _, err := os.Stat(expectedDocxPath); err == nil {
				actualFilePath = expectedDocxPath
				logx.Infof("找到转换后的 docx 文件: %s", actualFilePath)
			} else {
				// 如果文件名不匹配，查找 mdDir 中所有 docx 文件
				logx.Infof("预期路径不存在，查找目录中的所有 docx 文件: %s", absMdDir)
				files, err := os.ReadDir(absMdDir)
				if err != nil {
					return "", fmt.Errorf("读取输出目录失败: %v", err)
				}
				found := false
				for _, file := range files {
					if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".docx") {
						actualFilePath = filepath.Join(absMdDir, file.Name())
						found = true
						logx.Infof("找到 docx 文件: %s", actualFilePath)
						break
					}
				}
				if !found {
					return "", fmt.Errorf("doc 转 docx 失败，LibreOffice 命令执行成功但找不到转换后的文件在目录: %s, 标准输出: %s, 标准错误: %s", absMdDir, stdoutStr, stderrStr)
				}
			}
		}

		// 验证转换后的文件是否存在
		if _, err := os.Stat(actualFilePath); os.IsNotExist(err) {
			return "", fmt.Errorf("doc 转 docx 失败，找不到转换后的文件: %s", actualFilePath)
		}

		logx.Infof("doc 转 docx 成功，转换后的文件: %s", actualFilePath)
	} else {
		actualFilePath = absDocPath
	}

	// absMdDir 已在上面获取，这里直接使用

	// 按你约定：
	// 1) file.doc：先转到 file.doc.file/file.docx，然后进入 file.doc.file，执行：
	//    pandoc file.docx -o file.md --extract-media=. --to=gfm
	// 2) file.docx：进入 file.docx.file，执行：
	//    pandoc ../file.docx -o file.md --extract-media=. --to=gfm
	//
	// 所以这里统一：进入 mdDir，pandoc 的输入参数用“相对 mdDir 的相对路径”（不带绝对路径）。
	relInputPath, err := filepath.Rel(absMdDir, actualFilePath)
	if err != nil {
		relInputPath = filepath.Base(actualFilePath)
	}
	relInputPath = filepath.ToSlash(relInputPath)

	outputFileName := nameWithoutExt + ".md"
	cmd := exec.CommandContext(ctx, "pandoc", relInputPath, "-o", outputFileName, "--extract-media=.", "--to=gfm")

	// 设置工作目录为 mdDir（media 也落在 mdDir）
	cmd.Dir = absMdDir

	// 捕获标准输出和标准错误
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	logx.Infof("执行命令: pandoc %s -o %s --extract-media=. --to=gfm (工作目录: %s)", relInputPath, outputFileName, absMdDir)

	runErr := cmd.Run()
	if runErr != nil {
		errorMsg := stderr.String()
		if errorMsg == "" {
			errorMsg = runErr.Error()
		}
		return "", fmt.Errorf("pandoc 转换失败: %v, 错误输出: %s", runErr, errorMsg)
	}

	// 验证输出文件是否存在
	finalMdPath := filepath.Join(absMdDir, outputFileName)
	if _, err := os.Stat(finalMdPath); os.IsNotExist(err) {
		return "", fmt.Errorf("转换命令成功但输出文件不存在: %s", finalMdPath)
	}

	logx.Infof("成功转换文件: %s -> %s", docPath, finalMdPath)
	return finalMdPath, nil
}
