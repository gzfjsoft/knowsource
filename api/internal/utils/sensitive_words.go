package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
)

// SensitiveWordFilter 敏感词过滤器
type SensitiveWordFilter struct {
	words   []string         // 敏感词列表
	regexps []*regexp.Regexp // 编译后的正则表达式
	trie    *TrieNode        // Trie树用于快速匹配
	mutex   sync.RWMutex     // 读写锁
}

// TrieNode Trie树节点
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

// NewSensitiveWordFilter 创建新的敏感词过滤器
func NewSensitiveWordFilter() *SensitiveWordFilter {
	return &SensitiveWordFilter{
		words:   make([]string, 0),
		regexps: make([]*regexp.Regexp, 0),
		trie:    &TrieNode{children: make(map[rune]*TrieNode)},
	}
}
func GetSensitiveWordsFilePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)
	confpath := filepath.Join(execDir, "sensitive-words.txt")

	logx.Infof("GetSensitiveWordsFilePath confpath: %s", confpath)

	return confpath, nil
}

// DefaultSensitiveWords 默认敏感词列表
var DefaultSensitiveWords = []string{
	"政治敏感词",
	"涉黄",
	"涉赌",
	"涉毒",
	"暴力",
	"恐怖主义",
	"邪教",
	"反动",
	"分裂",
	"诈骗",
	"违法",
	"犯罪",
	"TMD",
	"法轮功",
	"傻逼",
	"你妈逼",
	"你爸逼",
	"你大爷逼",
}

// DefaultSensitiveWordFilter 默认的全局敏感词过滤器
var (
	defaultFilter *SensitiveWordFilter
	once          sync.Once
)

// GetDefaultFilter 获取默认的敏感词过滤器（单例模式）
func GetDefaultFilter() *SensitiveWordFilter {
	once.Do(func() {
		defaultFilter = NewSensitiveWordFilter()
		defaultFilter.AddWords(DefaultSensitiveWords...)
	})
	return defaultFilter
}

// AddWords 添加敏感词
func (f *SensitiveWordFilter) AddWords(words ...string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for _, word := range words {
		if word == "" {
			continue
		}

		// 添加到词列表
		f.words = append(f.words, word)

		// 编译正则表达式（用于模糊匹配）
		pattern := regexp.QuoteMeta(word)
		if regex, err := regexp.Compile("(?i)" + pattern); err == nil {
			f.regexps = append(f.regexps, regex)
		}

		// 添加到Trie树
		f.addToTrie(word)
	}
}

// addToTrie 将词添加到Trie树
func (f *SensitiveWordFilter) addToTrie(word string) {
	node := f.trie
	for _, char := range word {
		if node.children[char] == nil {
			node.children[char] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[char]
	}
	node.isEnd = true
}

// RemoveWords 移除敏感词
func (f *SensitiveWordFilter) RemoveWords(words ...string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// 重建词列表和正则表达式
	newWords := make([]string, 0)
	newRegexps := make([]*regexp.Regexp, 0)

	for i, word := range f.words {
		found := false
		for _, removeWord := range words {
			if word == removeWord {
				found = true
				break
			}
		}
		if !found {
			newWords = append(newWords, word)
			if i < len(f.regexps) {
				newRegexps = append(newRegexps, f.regexps[i])
			}
		}
	}

	f.words = newWords
	f.regexps = newRegexps

	// 重建Trie树
	f.rebuildTrie()
}

// rebuildTrie 重建Trie树
func (f *SensitiveWordFilter) rebuildTrie() {
	f.trie = &TrieNode{children: make(map[rune]*TrieNode)}
	for _, word := range f.words {
		f.addToTrie(word)
	}
}

// ContainsSensitiveWord 检查文本是否包含敏感词
func (f *SensitiveWordFilter) ContainsSensitiveWord(text string) bool {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// 使用Trie树进行快速检测
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if f.searchInTrie(runes, i) {
			return true
		}
	}

	return false
}

// searchInTrie 在Trie树中搜索
func (f *SensitiveWordFilter) searchInTrie(text []rune, start int) bool {
	node := f.trie
	for i := start; i < len(text); i++ {
		char := text[i]
		if node.children[char] == nil {
			break
		}
		node = node.children[char]
		if node.isEnd {
			return true
		}
	}
	return false
}

// FindSensitiveWords 查找文本中的所有敏感词
func (f *SensitiveWordFilter) FindSensitiveWords(text string) []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	foundWords := make([]string, 0)
	foundMap := make(map[string]bool) // 用于去重

	// 使用Trie树查找
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if word := f.findWordInTrie(runes, i); word != "" {
			if !foundMap[word] {
				foundWords = append(foundWords, word)
				foundMap[word] = true
			}
		}
	}

	return foundWords
}

// findWordInTrie 在Trie树中查找完整的敏感词
func (f *SensitiveWordFilter) findWordInTrie(text []rune, start int) string {
	node := f.trie
	var word strings.Builder

	for i := start; i < len(text); i++ {
		char := text[i]
		if node.children[char] == nil {
			break
		}
		node = node.children[char]
		word.WriteRune(char)
		if node.isEnd {
			return word.String()
		}
	}
	return ""
}

// FilterSensitiveWords 过滤敏感词，用指定字符替换
func (f *SensitiveWordFilter) FilterSensitiveWords(text string, replacement string) string {
	if replacement == "" {
		replacement = "*"
	}

	f.mutex.RLock()
	defer f.mutex.RUnlock()

	result := text

	// 使用正则表达式进行替换
	for _, regex := range f.regexps {
		result = regex.ReplaceAllStringFunc(result, func(match string) string {
			return strings.Repeat(replacement, len([]rune(match)))
		})
	}

	return result
}

// FilterSensitiveWordsAdvanced 高级过滤，支持自定义替换规则
func (f *SensitiveWordFilter) FilterSensitiveWordsAdvanced(text string, replaceFunc func(string) string) string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	result := text

	for _, regex := range f.regexps {
		result = regex.ReplaceAllStringFunc(result, replaceFunc)
	}

	return result
}

// GetWordCount 获取敏感词数量
func (f *SensitiveWordFilter) GetWordCount() int {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return len(f.words)
}

// GetWords 获取所有敏感词
func (f *SensitiveWordFilter) GetWords() []string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	result := make([]string, len(f.words))
	copy(result, f.words)
	return result
}

// Clear 清空所有敏感词
func (f *SensitiveWordFilter) Clear() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	f.words = make([]string, 0)
	f.regexps = make([]*regexp.Regexp, 0)
	f.trie = &TrieNode{children: make(map[rune]*TrieNode)}
}

// 便捷函数，使用默认过滤器

// ContainsSensitiveWord 检查文本是否包含敏感词（使用默认过滤器）
func ContainsSensitiveWord(text string) bool {
	return GetDefaultFilter().ContainsSensitiveWord(text)
}

// FindSensitiveWords 查找文本中的敏感词（使用默认过滤器）
func FindSensitiveWords(text string) []string {
	return GetDefaultFilter().FindSensitiveWords(text)
}

// FilterSensitiveWords 过滤敏感词（使用默认过滤器）
func FilterSensitiveWords(text string, replacement ...string) string {
	rep := "*"
	if len(replacement) > 0 && replacement[0] != "" {
		rep = replacement[0]
	}
	return GetDefaultFilter().FilterSensitiveWords(text, rep)
}

// AddSensitiveWords 添加敏感词到默认过滤器
func AddSensitiveWords(words ...string) {
	GetDefaultFilter().AddWords(words...)
}

// RenewSensitiveWords 重新加载敏感词
func RenewSensitiveWords(words ...string) {
	GetDefaultFilter().Clear()
	GetDefaultFilter().AddWords(words...)
}

// RemoveSensitiveWords 从默认过滤器移除敏感词
func RemoveSensitiveWords(words ...string) {
	GetDefaultFilter().RemoveWords(words...)
}
