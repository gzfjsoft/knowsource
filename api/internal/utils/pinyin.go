package utils

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// ConvertToPinyinCode 将字符串转换为拼音代码
// 中文字符转换为拼音，英文和数字保持原样
// 例如: "测试ABC123" -> "ceshiABC123"
func ConvertToPinyinCode(text string) string {
	if text == "" {
		return ""
	}

	var result strings.Builder

	runes := []rune(text)
	for _, r := range runes {
		// 判断是否为中文字符
		if unicode.Is(unicode.Han, r) {
			// 转换为拼音（不带声调，小写）
			pinyinSlice := pinyin.LazyPinyin(string(r), pinyin.NewArgs())
			if len(pinyinSlice) > 0 {
				result.WriteString(pinyinSlice[0])
			}
		} else {
			// 英文、数字和其他字符保持原样
			result.WriteRune(r)
		}
	}

	return result.String()
}

