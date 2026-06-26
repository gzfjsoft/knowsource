package main

import (
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/pmezard/go-difflib/difflib"
)

const expected = `abc
...`

const actual = `hello
abc`

func abcd() {

	// if a, e := strings.TrimSpace(actual), strings.TrimSpace(expected); a != e {
	// 	fmt.Println(diff.LineDiff(e, a))
	// 	fmt.Println()
	// 	fmt.Println(diff.LineDiff(a, e))
	// }

	// 定义两个需要比较的字符串
	text1 := "Hello, world!"
	text2 := "Hello, Go world!"
	text1 = expected
	text2 = actual

	// 创建一个 diffmatchpatch 实例
	dmp := diffmatchpatch.New()

	// 计算两个字符串的差异
	diffs := dmp.DiffMain(text1, text2, false)

	// 输出差异结果
	fmt.Println(dmp.DiffPrettyText(diffs))

	fmt.Println(diffs)

	//====================
	mainx()

}

func mainxx() {
	// 定义要比较的两个字符串
	text1 := `第一行
第二行
第三行`
	text2 := `第一行
第二行修改
第三行`

	text1 = expected
	text2 = actual

	// 创建一个统一的差异比较器
	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(text1),
		B:        difflib.SplitLines(text2),
		FromFile: "原始文本",
		ToFile:   "修改后文本",
		Context:  3,
	}

	// 生成差异结果
	result, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		fmt.Println("生成差异结果时出错:", err)
		return
	}

	// 输出差异结果
	fmt.Println(result)
}
