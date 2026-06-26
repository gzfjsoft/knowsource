package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

const RawDocQaLLMLogDir = "rawdoc_qa_llm_logs"

var rawDocQaLogFileNameRE = regexp.MustCompile(`[<>:"/\\|?*\s]+`)

func sanitizeRawDocQaLogName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}
	s = rawDocQaLogFileNameRE.ReplaceAllString(s, "_")
	return strings.Trim(strings.TrimSuffix(s, "."), "_")
}

func BuildRawDocQaLLMLogPath(clientId string, rawDocumentId int64, fileName string) string {
	clientId = sanitizeRawDocQaLogName(clientId)
	if clientId == "" {
		clientId = "unknown-client"
	}
	fileName = sanitizeRawDocQaLogName(fileName)
	if fileName == "" {
		fileName = "unknown-file"
	}
	clientDir := filepath.Join(RawDocQaLLMLogDir, clientId)
	_ = os.MkdirAll(clientDir, 0755)
	logName := "rawdoc_" + strconv.FormatInt(rawDocumentId, 10) + "_" + fileName + "_" + time.Now().Format("20060102_150405") + ".log.txt"
	return filepath.Join(clientDir, logName)
}

func AppendRawDocQaLLMLog(filePath, stage, content string) {
	if strings.TrimSpace(filePath) == "" {
		return
	}
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logx.Errorf("打开问答抽取日志文件失败 path=%s err=%v", filePath, err)
		return
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	block := fmt.Sprintf("\n[%s] [%s]\n%s\n", ts, strings.TrimSpace(stage), strings.TrimSpace(content))
	if _, err = f.WriteString(block); err != nil {
		logx.Errorf("写入问答抽取日志失败 path=%s err=%v", filePath, err)
	}
}
