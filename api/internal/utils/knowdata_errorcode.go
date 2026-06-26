package utils

import (
	"regexp"
)

// ErrorCodeFinder 错误码查找器
type ErrorCodeFinder struct {
	pattern *regexp.Regexp
}

// NewErrorCodeFinder 创建新的错误码查找器
func NewErrorCodeFinder() *ErrorCodeFinder {
	// 匹配规则：
	// - 必须包含至少一个数字
	// - 可以包含英文字母和下划线
	// - 连续的字符序列，包括纯数字
	pattern := regexp.MustCompile(`\b(?:[A-Za-z_]*\d[A-Za-z_\d]*|\d+)\b`)
	return &ErrorCodeFinder{pattern: pattern}
}

// FindErrorCodes 在给定文本中查找所有错误码
func (f *ErrorCodeFinder) FindErrorCodes(text string) []string {
	matches := f.pattern.FindAllString(text, -1)

	var errorCodes []string
	seen := make(map[string]bool) // 去重

	for _, match := range matches {
		if f.isValidErrorCode(match) && !seen[match] {
			errorCodes = append(errorCodes, match)
			seen[match] = true
		}
	}

	return errorCodes
}

// isValidErrorCode 验证是否为有效的错误码
func (f *ErrorCodeFinder) isValidErrorCode(code string) bool {
	// 必须包含数字
	hasDigit := regexp.MustCompile(`\d`).MatchString(code)
	if !hasDigit {
		return false
	}

	// 只能包含字母、下划线、数字
	validChars := regexp.MustCompile(`^[A-Za-z_\d]+$`).MatchString(code)
	if !validChars {
		return false
	}

	// 纯数字也是有效的错误码（根据需求：错误码肯定有数字，但不一定有英文或下划线）
	// 移除单个字符的限制，因为单个数字也可能是错误码

	return true
}
