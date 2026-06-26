package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func mainx() {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("获取当前目录失败: %v\n", err)
		return
	}

	// 查找所有 handle 文件
	var handleFiles []string
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), "handler.go") {
			handleFiles = append(handleFiles, path)
		}
		return nil
	})
	fmt.Printf("handleFiles: %v\n", handleFiles)

	if err != nil {
		fmt.Printf("遍历目录失败: %v\n", err)
		return
	}

	// 查找所有 api 文件
	var apiFiles []string
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".api") {
			apiFiles = append(apiFiles, path)
		}
		return nil
	})

	// 检查每个 handle 文件是否被引用
	for _, handleFile := range handleFiles {
		// handleName := handleFile

		handleName := filepath.Base(handleFile)
		handleName = strings.TrimSuffix(handleName, ".go")

		isReferenced := false

		// 检查所有 api 文件中是否包含对该 handle 的引用
		for _, apiFile := range apiFiles {
			content, err := ioutil.ReadFile(apiFile)
			if err != nil {
				fmt.Printf("读取文件 %s 失败: %v\n", apiFile, err)
				continue
			}

			if strings.Contains(strings.ToLower(string(content)), handleName) {
				isReferenced = true
				break
			}
		}

		// 如果没有被引用，删除 handle 和对应的 logic 文件
		if !isReferenced {
			fmt.Printf("删除未使用的文件: %s\n", handleFile)
			logicFile := strings.Replace(handleFile, "handler", "logic", 2)
			fmt.Printf("logicFile: %s\n", logicFile)

			os.Remove(handleFile)
			os.Remove(logicFile)

			// // 构造并删除对应的 logic 文件

		} else {
			fmt.Printf("============文件 %s 被引用\n", handleFile)
		}
	}
}
