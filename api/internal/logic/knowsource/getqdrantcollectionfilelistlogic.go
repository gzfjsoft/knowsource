package knowsource

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQdrantCollectionFileListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取 qrdant collection list
func NewGetQdrantCollectionFileListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQdrantCollectionFileListLogic {
	return &GetQdrantCollectionFileListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQdrantCollectionFileListLogic) GetQdrantCollectionFileList(req *types.QdrantCollectionFileListRequest) (resp *types.QdrantCollectionFileListResponse, err error) {
	newResp := func(code int64, message, info string) *types.QdrantCollectionFileListResponse {
		return &types.QdrantCollectionFileListResponse{
			Response: types.Response{
				Code:    code,
				Message: message,
				Info:    info,
			},
		}
	}

	if req.Collection == "" {
		return newResp(response.InvalidRequestParamCode, "collection 不能为空", "collection is empty"), nil
	}

	cfg := l.svcCtx.Config.Qdrant
	if cfg.Host == "" || cfg.Port == 0 {
		return newResp(response.ServerErrorCode, "Qdrant 配置缺失", "请在 knowsource.yaml 中配置 Qdrant.host 和 Qdrant.port"), nil
	}

	// 调用 Qdrant Scroll HTTP 接口，读取该 collection 中的 payload 信息
	url := fmt.Sprintf("http://%s:%d/collections/%s/points/scroll", cfg.Host, cfg.Port, req.Collection)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 简单一次性读取较多条记录（假定单个 collection 文件数量不会特别大）
	body := map[string]interface{}{
		"limit":        1000,
		"with_payload": true,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		l.Errorf("序列化 Qdrant Scroll 请求体失败: %v", err)
		return newResp(response.ServerErrorCode, "创建 Qdrant 请求失败", err.Error()), nil
	}

	httpReq, err := http.NewRequestWithContext(l.ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		l.Errorf("创建 Qdrant Scroll 请求失败: %v", err)
		return newResp(response.ServerErrorCode, "创建 Qdrant 请求失败", err.Error()), nil
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := httpClient.Do(httpReq)
	if err != nil {
		l.Errorf("请求 Qdrant Scroll 失败: %v", err)
		return newResp(response.ServerErrorCode, "请求 Qdrant 失败", err.Error()), nil
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		info := fmt.Sprintf("status=%d, body=%s", httpResp.StatusCode, string(respBody))
		l.Errorf("Qdrant Scroll 返回非 200 状态码: %s", info)
		return newResp(response.ServerErrorCode, "Qdrant 返回错误状态码", info), nil
	}
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		l.Errorf("读取 Qdrant Scroll 响应失败: %v", err)
		return newResp(response.ServerErrorCode, "读取 Qdrant 响应失败", err.Error()), nil
	}

	l.Infof("Qdrant 响应: %s", string(respBody))

	// Qdrant Scroll 响应结构
	type point struct {
		Id      interface{}            `json:"id"`
		Payload map[string]interface{} `json:"payload"`
	}
	type scrollResult struct {
		Points []point `json:"points"`
	}
	type qdrantResp struct {
		Result scrollResult `json:"result"`
		Status string       `json:"status"`
		Time   float64      `json:"time"`
	}

	var qResp qdrantResp
	if err := json.Unmarshal(respBody, &qResp); err != nil {
		l.Errorf("解析 Qdrant Scroll 响应失败: %v, body=%s", err, string(respBody))
		return newResp(response.ServerErrorCode, "解析 Qdrant 响应失败", err.Error()), nil
	}

	// 收集 payload 中的详细信息：id、fileName、documentCode、content、metadata
	items := make([]types.QdrantCollectionFileInfo, 0, len(qResp.Result.Points))
	for _, p := range qResp.Result.Points {
		if p.Payload == nil {
			continue
		}

		info := types.QdrantCollectionFileInfo{
			DocumentCode: req.Collection,
		}

		// id 在最外层
		switch v := p.Id.(type) {
		case string:
			info.Id = v
		case float64:
			// 数字 ID，转成无小数的字符串
			info.Id = fmt.Sprintf("%.0f", v)
		}

		// page_content，截取前 50 个「字符」（按 rune 计），避免截断中文
		if v, ok := p.Payload["page_content"]; ok {
			if s, ok2 := v.(string); ok2 {
				runes := []rune(s)
				if len(runes) > 50 {
					info.Content = string(runes[:50]) + "..."
				} else {
					info.Content = s
				}
			}
		}

		// metadata：原样转成 JSON 字符串，并从中提取 page / total_pages / length 等字段
		if v, ok := p.Payload["metadata"]; ok && v != nil {
			if m, ok2 := v.(map[string]interface{}); ok2 {
				if b, err := json.Marshal(m); err == nil {
					info.Metadata = string(b)
					// 从 metadata.path 推导文件名
					if pathVal, ok3 := m["path"]; ok3 {
						if ps, ok4 := pathVal.(string); ok4 && ps != "" {
							info.FileName = ps
						}
					}

					// page
					if pageVal, ok3 := m["page"]; ok3 {
						switch pv := pageVal.(type) {
						case float64:
							info.Page = int64(pv)
						case int64:
							info.Page = pv
						case int:
							info.Page = int64(pv)
						}
					}

					// total_pages
					if tpVal, ok3 := m["total_pages"]; ok3 {
						switch tv := tpVal.(type) {
						case float64:
							info.TotalPages = int64(tv)
						case int64:
							info.TotalPages = tv
						case int:
							info.TotalPages = int64(tv)
						}
					}

					// length
					if lenVal, ok3 := m["length"]; ok3 {
						switch lv := lenVal.(type) {
						case float64:
							info.Length = int64(lv)
						case int64:
							info.Length = lv
						case int:
							info.Length = int64(lv)
						}
					}
				}
			}
		}

		items = append(items, info)
	}

	return &types.QdrantCollectionFileListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.QdrantCollectionFileListData{
			List:  items,
			Total: int64(len(items)),
		},
	}, nil
}
