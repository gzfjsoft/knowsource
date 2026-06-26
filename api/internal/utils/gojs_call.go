package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dop251/goja"
)

// 从外部文件读取JavaScript代码
// scriptContent, err := ioutil.ReadFile("script.js")
// if err != nil {
// 	fmt.Printf("读取脚本文件错误: %v\n", err)
// 	os.Exit(1)
// }

func GojsCall(scriptContent string) map[string]interface{} {
	// 创建一个新的JavaScript虚拟机
	vm := goja.New()

	// 注册console对象到JS环境
	console := vm.NewObject()
	// 添加console.log方法
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	// 添加console.error方法
	console.Set("error", func(call goja.FunctionCall) goja.Value {
		var args []interface{}
		for _, arg := range call.Arguments {
			args = append(args, arg.String())
		}
		fmt.Fprintln(os.Stderr, args...)
		return goja.Undefined()
	})
	vm.Set("console", console)

	// 注册HTTP GET请求函数到JS环境
	vm.Set("httpGet", func(url string) (string, error) {
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(body), nil
	})

	// 注册HTTP POST请求函数到JS环境
	vm.Set("httpPost", func(url string, data string) (string, error) {
		resp, err := http.Post(url, "application/json", bytes.NewBuffer([]byte(data)))
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(body), nil
	})

	// 执行脚本并获取返回值
	result, err := vm.RunString(string(scriptContent))
	if err != nil {
		fmt.Printf("脚本执行错误: %v\n", err)
		return map[string]interface{}{"success": false, "error": err.Error(), "message": "脚本执行错误"}
	}

	// 将JavaScript返回值转换为Go映射
	var returnData map[string]interface{}
	if err := vm.ExportTo(result, &returnData); err == nil {
		// 打印从JavaScript返回的数据
		fmt.Println("\n从JavaScript返回的结果:")
		if success, ok := returnData["success"].(bool); ok {
			if success {
				fmt.Printf("请求状态: 成功\n")
				if length, ok := returnData["length"].(float64); ok {
					fmt.Printf("响应长度: %.0f\n", length)
				}
				if content, ok := returnData["content"].(string); ok {
					fmt.Printf("响应内容预览: %s\n", content)
				}
			} else {
				fmt.Printf("请求状态: 失败\n")
				if errMsg, ok := returnData["error"].(string); ok {
					fmt.Printf("错误信息: %s\n", errMsg)
				}
			}
		}
	} else {
		// 如果转换失败，尝试直接打印结果
		fmt.Printf("脚本返回结果: %v\n", result)
	}

	return returnData
}
