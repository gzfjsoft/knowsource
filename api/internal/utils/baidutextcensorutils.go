package utils

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"io/ioutil"
	"net/http"
	url2 "net/url"
	"strings"
)

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ErrorCode   int64  `json:"error_code"`
	ErrorMsg    string `json:"error_msg"`
}

type TextSensorResponse struct {
	LogId          int64                    `json:"log_id"`
	ErrorCode      int64                    `json:"error_code"`
	ErrorMsg       string                   `json:"error_msg"`
	ConclusionType int64                    `json:"conclusionType"` // 1.合规，2.不合规，3.疑似，4.审核失败
	Conclusion     string                   `json:"conclusion"`
	Data           []TextSensorResponseData `json:"data"`
}

type TextSensorResponseData struct {
	ErrorCode int64      `json:"error_code"`
	ErrorMsg  string     `json:"error_msg"`
	Type      int64      `json:"type"`
	SubType   int64      `json:"subType"`
	Msg       string     `json:"msg"`
	Hits      []DataHits `json:"hits"`
}
type DataHits struct {
	DatasetName string   `json:"datasetName"`
	Words       []string `json:"words"`
}

const GetAccessTokenFailed = "获取百度access token失败"
const TextSensorFailed = "敏感词检测失败，请稍后重试 "

func GetAccessToken(clientId, clientSecret string) (string, error) {
	url := fmt.Sprintf("https://aip.baidubce.com/oauth/2.0/token?client_id=%s&client_secret=%s&grant_type=client_credentials", clientId, clientSecret)
	payload := strings.NewReader(``)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		logx.Errorf("%s: %s", GetAccessTokenFailed, err.Error())
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		logx.Errorf("%s: %s", GetAccessTokenFailed, err.Error())
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logx.Errorf("%s: %s", GetAccessTokenFailed, err.Error())
		return "", err
	}
	var resp AccessTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		logx.Errorf("%s: %s", GetAccessTokenFailed, err.Error())
		return "", err
	}
	if resp.ErrorCode != 0 || resp.ErrorMsg != "" {
		return "", errors.Errorf("%s, 错误码: %d, 描述: %s", GetAccessTokenFailed, resp.ErrorCode, resp.ErrorMsg)
	}
	return resp.AccessToken, nil
}

func BaiduTextSensor(text string) string {
	accessToken, err := GetAccessToken("E0NFufdSBXNIEXZDjxsYsA7D", "nCmPf4uqJ9FYG3fxmvIKMLzmhCCrNOAT")
	if err != nil {
		logx.Errorf("%s: %s", GetAccessTokenFailed, err.Error())
		return TextSensorFailed
	}
	v := url2.Values{}
	v.Add("text", text)
	bodyStr := v.Encode()
	url := fmt.Sprintf("https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined?access_token=%s", accessToken)
	payload := strings.NewReader(bodyStr)
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		logx.Errorf("调用百度文本审核失败: %s", err.Error())
		return TextSensorFailed
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		logx.Errorf("调用百度文本审核失败: %s", err.Error())
		return TextSensorFailed
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logx.Errorf("调用百度文本审核失败: %s", err.Error())
		return TextSensorFailed
	}
	var resp TextSensorResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		logx.Errorf("调用百度文本审核失败: %s", err.Error())
		return TextSensorFailed
	}
	if resp.ErrorCode != 0 || resp.ErrorMsg != "" {
		logx.Errorf("调用百度文本审核失败, 错误码: %d, 描述: %s", resp.ErrorCode, resp.ErrorMsg)
		return TextSensorFailed
	}
	logx.Infof("调用百度文本审核成功: %s", string(body))
	if resp.ConclusionType == 2 {
		// 不合规
		sensitiveWordMap := make(map[string]bool)
		for _, data := range resp.Data {
			for _, hit := range data.Hits {
				for _, word := range hit.Words {
					sensitiveWordMap[word] = false
				}
			}
		}
		var sensitiveWordArray []string
		for s := range sensitiveWordMap {
			sensitiveWordArray = append(sensitiveWordArray, s)
		}
		sensitiveWords := strings.Join(sensitiveWordArray, ",")
		logx.Infof("调用百度文本审核结果不合规, 不合规词语: %s", sensitiveWords)
		return sensitiveWords
	}
	return ""
}
