package utils

// CopyFile 复制文件

import (
	"archive/zip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

func CopyFile(src, dst string) error {
	// 打开源文件
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// 创建目标文件
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 复制文件内容
	_, err = io.Copy(destination, source)
	return err
}

// CopyFileWithBuffer 复制文件，使用缓冲区
func CopyFileWithBuffer(src, dst string) error {
	// 打开源文件
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// 创建目标文件
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	// 使用缓冲区复制文件内容
	_, err = io.CopyBuffer(destination, source, make([]byte, 1024))
	return err
}

// 图片加水印
func AddWatermarkToImage(srcImage image.Image, dstPath, watermarkPath string) error {
	// 2. 创建水印图片（这里创建一个简单的文字水印）
	watermark, err := imaging.Open(watermarkPath)
	if err != nil {
		return err
	}

	// 3. 平铺水印（45度）
	result := tileWatermarkAt45Degrees(srcImage, watermark)

	// 4. 保存结果
	err = imaging.Save(result, dstPath)
	if err != nil {
		return err
	}

	return nil
}

func AddWatermark(srcPath, dstPath, watermarkPath string) error {
	// 1. 加载原始图片
	src, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}

	// 2. 创建水印图片（这里创建一个简单的文字水印）
	watermark, err := imaging.Open(watermarkPath)
	if err != nil {
		return err
	}

	// 3. 平铺水印（45度）
	result := tileWatermarkAt45Degrees(src, watermark)

	// 4. 保存结果
	err = imaging.Save(result, dstPath)
	if err != nil {
		return err
	}

	return nil
}

// 45度平铺水印
func tileWatermarkAt45Degrees(dst, watermark image.Image) image.Image {
	// 获取水印尺寸
	wmBounds := watermark.Bounds()
	wmWidth := wmBounds.Dx()
	wmHeight := wmBounds.Dy()

	// 计算水印对角线长度（用于确定平铺间距）
	diagonal := math.Sqrt(float64(wmWidth*wmWidth + wmHeight*wmHeight))
	spacing := int(diagonal * 1.2) // 间距稍大于对角线长度

	// 创建目标图像的副本
	result := imaging.Clone(dst)
	bounds := result.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// 计算需要多少个水印才能覆盖整个图像
	countX := int(float64(width)/float64(spacing)) + 2
	countY := int(float64(height)/float64(spacing)) + 2

	// 旋转水印45度
	rotatedWM := imaging.Rotate(watermark, 45, color.Transparent)

	// 平铺水印
	for i := -1; i < countX; i++ {
		for j := -1; j < countY; j++ {
			x := i*spacing - j*spacing/2
			y := j * spacing

			// 将水印绘制到目标图像上
			result = imaging.Overlay(result, rotatedWM, image.Pt(x, y), 0.5)
		}
	}

	return result
}

func FindExcelAndCsvFiles(rootPath string) ([]string, error) {
	var files []string

	// 检查目录是否存在
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", rootPath)
	}

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if (ext == ".xlsx" || ext == ".csv") && !strings.Contains(path, "__MACOSX/.") {
				files = append(files, path)
			}
		}

		return nil
	})
	fmt.Println(files)
	return files, err
}

// 解压ZIP文件到目标目录
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// 检查根目录是否只有一个文件夹
	var rootDirs []string
	var rootFiles []string

	for _, f := range r.File {
		// 清理路径，移除前导和尾随的斜杠
		cleanPath := strings.Trim(f.Name, "/")
		if cleanPath == "" {
			continue // 跳过空路径
		}

		parts := strings.Split(cleanPath, "/")
		if len(parts) > 0 && parts[0] != "" {
			if f.FileInfo().IsDir() || len(parts) > 1 {
				// 这是一个目录或者目录下的文件
				if !StringContains(rootDirs, parts[0]) && parts[0] != "__MACOSX" {
					rootDirs = append(rootDirs, parts[0])
				}
			} else {
				// 这是根目录下的文件
				if !StringContains(rootFiles, parts[0]) {
					rootFiles = append(rootFiles, parts[0])
				}
			}
		}
	}

	// 如果根目录下只有一个文件夹且没有其他文件，则提取该文件夹内的内容
	shouldFlatten := len(rootDirs) == 1 && len(rootFiles) == 0

	for _, f := range r.File {
		// 获取相对路径
		relPath := strings.Trim(f.Name, "/")
		if relPath == "" {
			continue // 跳过空路径
		}

		if shouldFlatten {
			// 移除第一层目录
			parts := strings.SplitN(relPath, "/", 2)
			if len(parts) > 1 {
				relPath = parts[1]
			} else {
				// 如果是根目录下的文件夹本身，跳过
				continue
			}
		}

		// 构建完整路径
		fpath := filepath.Join(dest, filepath.FromSlash(relPath))

		// 防止路径遍历攻击
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return fmt.Errorf("创建目录失败: %v", err)
			}
			continue
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return fmt.Errorf("创建目录失败: %v", err)
		}

		// 创建目标文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("创建文件失败: %v", err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("打开ZIP内容失败: %v", err)
		}

		_, err = io.Copy(outFile, rc)

		// 关闭文件句柄
		outFile.Close()
		rc.Close()

		if err != nil {
			return fmt.Errorf("写入文件失败: %v", err)
		}

		// 设置文件修改时间
		if !f.Modified.IsZero() {
			os.Chtimes(fpath, f.Modified, f.Modified)
		}
	}

	if shouldFlatten {
		os.RemoveAll(filepath.Join(dest, rootDirs[0])) // 删除被搬空的目录
	}
	return nil
}

// StringContains 检查字符串切片中是否包含指定字符串
func StringContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ListZipFiles 列出 ZIP 文件中的所有文件（不包括目录）
func ListZipFiles(zipPath string) ([]string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var files []string
	for _, f := range r.File {
		// 跳过目录
		if f.FileInfo().IsDir() {
			continue
		}
		// 跳过 __MACOSX 目录下的文件
		if strings.Contains(f.Name, "__MACOSX") {
			continue
		}
		// 清理路径，移除前导和尾随的斜杠
		cleanPath := strings.Trim(f.Name, "/")
		if cleanPath != "" {
			files = append(files, cleanPath)
		}
	}

	return files, nil
}

func HideStringInfo(str string, from, to int) string {
	if from < 0 || to > len(str) || from > to {
		return str
	}
	if from == 0 {
		return fmt.Sprintf("%s%s", str[:to], strings.Repeat("*", len(str)-to))
	}
	return fmt.Sprintf("%s%s%s", str[:from], strings.Repeat("*", to-from), str[to:])
}

// 对称加密
func EncryptStringAes(key, text string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptStringAes(key, text string) (string, error) {
	ciphertext, _ := base64.URLEncoding.DecodeString(text)
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}
