package utils

import (
	"fmt"
	"strings"
	"testing"
)

/*
go test -v -run ^TestKnowdatanErrcode$ todoplus/api/internal/utils
*/

func TestKnowdatanErrcode(t *testing.T) {
	// 使用简单版本
	// fmt.Println("=== 简单版本演示 ===")
	// simpleDemo()

	fmt.Println("\n=== 高级版本演示 ===")
	demonstrateErrorCodeFinder()
}

// func findErrorCodes(text string) []string {
// 	// 正则表达式匹配: 连续的英文字母、下划线、数字组成，且必须包含至少一个数字
// 	// 包括纯数字的情况
// 	re := regexp.MustCompile(`\b(?:[A-Za-z_]*\d[A-Za-z_\d]*|\d+)\b`)

// 	matches := re.FindAllString(text, -1)

// 	// 过滤结果，确保匹配的是真正的错误码格式
// 	var errorCodes []string
// 	for _, match := range matches {
// 		if isValidErrorCode(match) {
// 			errorCodes = append(errorCodes, match)
// 		}
// 	}

// 	return errorCodes
// }

// func isValidErrorCode(code string) bool {
// 	// 检查是否包含数字
// 	hasDigit := regexp.MustCompile(`\d`).MatchString(code)
// 	if !hasDigit {
// 		return false
// 	}

// 	// 检查是否只包含字母、下划线和数字
// 	validChars := regexp.MustCompile(`^[A-Za-z_\d]+$`).MatchString(code)
// 	if !validChars {
// 		return false
// 	}

// 	// 纯数字也是有效的错误码（根据需求：错误码肯定有数字，但不一定有英文或下划线）
// 	return true
// }

// 演示函数
func demonstrateErrorCodeFinder() {
	finder := NewErrorCodeFinder()

	testCases := []string{
		"Error occurred: ERR_404 please retry",
		"System failure: DB_TIMEOUT_500 and AUTH_FAILED_401",
		"Multiple errors: E123, NETWORK_ERROR_503, FILE_NOT_FOUND_404",
		"Status: 200 SUCCESS_200 FAIL_500",
		"Invalid: 123 (pure number)",
		"Valid codes: ERROR123ABC, _TIMEOUT_30, CODE_001",
		"Edge cases: A1, B2C, _500_, CODE__123__",
		"Mixed: 系统错误 ERR_NETWORK_500 和 DB_CONNECTION_FAILED_1001 请联系管理员",
		"11003 故障码",
		"15007",
	}

	fmt.Println("错误码识别演示")
	fmt.Println(strings.Repeat("=", 50))

	for i, testCase := range testCases {
		fmt.Printf("\n[测试 %d] %s\n", i+1, testCase)

		codes := finder.FindErrorCodes(testCase)
		if len(codes) == 0 {
			fmt.Println("结果: 未发现错误码")
		} else {
			fmt.Printf("结果: 发现 %d 个错误码 -> %v\n", len(codes), codes)
		}
	}
}

// func simpleDemo() {
// 	testStrings := []string{
// 		"发生错误 ERR_404 请重试",
// 		"系统异常: DB_CONNECTION_TIMEOUT_500 和 AUTH_FAILED_401",
// 		"错误代码: E123, NETWORK_ERROR_503, 和 FILE_NOT_FOUND_404",
// 		"状态码: 200 SUCCESS_200 FAIL_500",
// 		"ERROR123ABC 和 _TIMEOUT_30 都是有效的",
// 	}

// 	for i, text := range testStrings {
// 		fmt.Printf("测试 %d: %s\n", i+1, text)
// 		errorCodes := findErrorCodes(text)

// 		if len(errorCodes) == 0 {
// 			fmt.Println("  未找到错误码")
// 		} else {
// 			fmt.Printf("  找到错误码: %v\n", errorCodes)
// 		}
// 		fmt.Println()
// 	}
// }
