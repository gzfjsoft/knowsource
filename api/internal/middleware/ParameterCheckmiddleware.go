// middleware/ratelimit.go
package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type ParameterCheck struct {
}

func NewParameterCheck() *ParameterCheck {
	return &ParameterCheck{}
}

// checkValue 递归检查值中是否包含问号或百分号
func (l *ParameterCheck) checkValue(value interface{}) error {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "?") || strings.Contains(v, "%") {
			// 影响到正常业务了，不能在这里判断，否则会影响正常业务
			// logx.Warnf("JSON value should not contain question mark or percent sign: %s", v)
			// return fmt.Errorf("JSON value should not contain question mark or percent sign")
		}
	case []interface{}:
		// 如果是数组，递归检查每个元素
		for _, item := range v {
			if err := l.checkValue(item); err != nil {
				return err
			}
		}
	case map[string]interface{}:
		// 如果是对象，递归检查每个字段
		for _, item := range v {
			if err := l.checkValue(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func (l *ParameterCheck) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// 读取请求体
			body, err := io.ReadAll(r.Body)
			if err != nil {

				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			// 关闭原始请求体
			r.Body.Close()
			// 重新设置请求体，以便后续处理
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// 解析JSON数据
			var data map[string]interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				// 如果解析失败，可能不是JSON数据，继续处理请求
				next(w, r)
				return
			}

			// 检查键名是否包含问号
			for key, value := range data {
				if strings.Contains(key, "name") || strings.Contains(key, "Name") {
					if err := l.checkValue(value); err != nil {
						http.Error(w, err.Error(), http.StatusBadRequest)
						return
					}
				}
			}
		}

		// 获取appPlatform头
		appPlatformHeader := r.Header.Get("App-Platform")
		ctx := r.Context() // 将请求体数据添加到上下文中
		ctx = context.WithValue(ctx, "app-platform", appPlatformHeader)

		next(w, r)
	}
}
