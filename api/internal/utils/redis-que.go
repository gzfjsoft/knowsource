package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"knowsource/model"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/zeromicro/go-zero/core/logx"
)

// type BusinessMessage struct {
// 	Id        int    `json:"id"`        // 对应 Python 端的 id
// 	Content   string `json:"content"`   // 对应 Python 端的 content
// 	Timestamp string `json:"timestamp"` // 对应 Python 端的 timestamp
// }

// {'Authorization': 'RAG_SERVER_TOKEN', 'rawDocumentFileNames': [{'fileName': 'RD22 大部件更换模版.xlsx', 'documentCode': '..'}]}
type RawDocumentFileName struct {
	ClientId     string `json:"clientId"`
	FileName     string `json:"fileName"`
	DocumentCode string `json:"documentCode"`
}

type BusinessMessage struct {
	Authorization        string                `json:"Authorization"`
	RawDocumentFileNames []RawDocumentFileName `json:"rawDocumentFileNames"`
}

func GetMessageFromQueue(redisClient *redis.Redis, queueKey string, RawDocumentsModel model.RawDocumentsModel) {

	logx.Info("go-zero 消费者启动，监听队列：", queueKey)
	ctx := context.Background()
	for {
		// BRPOP 阻塞消费（超时时间 5 秒，0 表示永久阻塞）
		result, err := redisClient.RpopCtx(ctx, queueKey)
		if err != nil {
			if err.Error() == "redis: nil" {
				time.Sleep(5 * time.Second)
				continue
			}
			logx.Errorf("消费消息失败：%v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// 解析消息（BRPOP 返回 [key, msg]）
		if len(result) < 2 {
			continue
		}
		msg := result
		if msg == "" {
			continue
		}

		// 5. 解析 JSON 消息（和 Python 端的格式对应）
		var businessMsg BusinessMessage
		err = json.Unmarshal([]byte(msg), &businessMsg)
		if err != nil {
			logx.Errorf("解析消息失败：%v，原始消息：%s", err, msg)
			continue
		}

		info := fmt.Sprintf("消费到消息：\n  Authorization：%s\n  RawDocumentFileNames：%s\n", businessMsg.Authorization, businessMsg.RawDocumentFileNames)
		for _, rawDocumentFileName := range businessMsg.RawDocumentFileNames {
			info += fmt.Sprintf("    FileName：%s\n    DocumentCode：%s\n", rawDocumentFileName.FileName, rawDocumentFileName.DocumentCode)
		}
		logx.Infof("%s", info)

		// 更新 KnowledgeDataFileModel 的 is_on_rag = 1 where edit_status = 3 and 文件名 = 消息的文件名
		for _, rawDocumentFileName := range businessMsg.RawDocumentFileNames {
			fileName := rawDocumentFileName.FileName
			documentCode := rawDocumentFileName.DocumentCode
			if fileName != "" {
				clientId := strings.TrimSpace(rawDocumentFileName.ClientId)
				err := RawDocumentsModel.UpdateIsToAi(ctx, clientId, fileName, documentCode, 1)
				if err != nil {
					logx.Errorf("更新文件 %s 的 is_to_ai 状态失败：%v", fileName, err)
					continue
				}

			}
		}

		logx.Infof("%s", msg)

		logx.Infof("------------------------")
	}

}
