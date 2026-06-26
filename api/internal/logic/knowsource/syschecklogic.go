// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package knowsource

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysCheckLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 系统依赖检查：Vllmchat、Vllmembedding、Vllmreranker、Qdrant、Redis、Mysql
func NewSysCheckLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysCheckLogic {
	return &SysCheckLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysCheckLogic) SysCheck() (resp *types.SysCheckResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	chatURL, _, chatApiKey := utils.ResolveCompletionRuntime(&l.svcCtx.Config, clientId)
	embedURL, _, embedApiKey, _ := utils.ResolveEmbeddingRuntime(&l.svcCtx.Config, clientId)
	rerankURL, _, rerankApiKey, _ := utils.ResolveRerankRuntime(&l.svcCtx.Config, clientId)

	data := &types.SysCheckData{
		Vllmchat:      l.checkVllmWithTypeVersion(chatURL, chatApiKey),
		Vllmembedding: l.checkVllmWithTypeVersion(embedURL, embedApiKey),
		Vllmreranker:  l.checkVllmWithTypeVersion(rerankURL, rerankApiKey),
		Qdrant:        l.checkQdrant(),
		Redis:         l.checkRedis(),
		Mysql:         l.checkMysql(),
		Mineru:        l.checkMinerU(),
		Mail:          l.checkMail(),
	}
	// 每次 /sys/check 都更新内存中的模型 ID，供 chat/embedding/rerank 解析使用
	utils.LLMModelStore.Update(data.Vllmchat.ModelIds, data.Vllmembedding.ModelIds, data.Vllmreranker.ModelIds)
	return &types.SysCheckResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: data,
	}, nil
}

// vllmModelsResponse /v1/models 响应结构
type vllmModelsResponse struct {
	Data []struct {
		ID      string `json:"id"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

// checkVllm 检查 vLLM 服务是否可访问，并拉取 /v1/models 显示 model id（owned_by 为 vllm 的）
func (l *SysCheckLogic) checkVllm(url, apiKey, name string) types.SysCheckItem {
	if url == "" {
		return types.SysCheckItem{Ok: false, Value: "", Message: "未配置"}
	}
	base := strings.TrimSuffix(url, "/")
	if base == "" {
		base = url
	}
	// 优先请求 /v1/models，既可判断可用性又能拿到 model 列表
	u := base + "/v1/models"
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(l.ctx, http.MethodGet, u, nil)
	if err != nil {
		return types.SysCheckItem{Ok: false, Value: url, Message: err.Error()}
	}
	if strings.TrimSpace(apiKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	}
	res, err := client.Do(req)
	if err != nil {
		// 回退到 /health 或 /
		for _, path := range []string{"/health", "/"} {
			u2 := base + path
			req2, _ := http.NewRequestWithContext(l.ctx, http.MethodGet, u2, nil)
			if strings.TrimSpace(apiKey) != "" {
				req2.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
			}
			res2, err2 := client.Do(req2)
			if err2 != nil {
				continue
			}
			res2.Body.Close()
			if res2.StatusCode < 500 {
				return types.SysCheckItem{Ok: true, Value: url, Message: "可访问"}
			}
		}
		return types.SysCheckItem{Ok: false, Value: url, Message: "请求失败或超时"}
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return types.SysCheckItem{Ok: false, Value: url, Message: fmt.Sprintf("HTTP %d", res.StatusCode)}
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return types.SysCheckItem{Ok: true, Value: url, Message: "可访问"}
	}
	var vllmResp vllmModelsResponse
	if err := json.Unmarshal(body, &vllmResp); err != nil {
		return types.SysCheckItem{Ok: true, Value: url, Message: "可访问"}
	}
	var modelIds []string
	for _, m := range vllmResp.Data {
		if m.ID != "" {
			modelIds = append(modelIds, m.ID)
		}
	}
	return types.SysCheckItem{
		Ok:       true,
		Value:    url,
		Message:  "可访问",
		ModelIds: modelIds,
	}
}

// checkVllmWithTypeVersion 在 checkVllm 基础上，对可访问的服务探测 type（ollama/vllm）和 version
func (l *SysCheckLogic) checkVllmWithTypeVersion(url, apiKey string) types.SysCheckItem {
	item := l.checkVllm(url, apiKey, "")
	if !item.Ok || url == "" {
		return item
	}
	base := strings.TrimSuffix(url, "/")
	if ctype, version, err := utils.GetCompletionTypeAndVersion(base); err == nil {
		item.Type = ctype
		item.Version = version
	}
	return item
}

func (l *SysCheckLogic) checkMail() types.SysCheckItem {
	m := l.svcCtx.Config.Mail
	host := strings.TrimSpace(m.MailHost)
	acc := strings.TrimSpace(m.MailAccount)
	port := m.MailPort
	pass := m.MailPass
	useSSL := m.MailSSL || port == 465

	if host == "" || acc == "" || port <= 0 {
		return types.SysCheckItem{Ok: false, Value: "", Message: "未配置"}
	}
	if pass == "" {
		// 不强制，但提示
		return types.SysCheckItem{Ok: false, Value: fmt.Sprintf("%s:%d", host, port), Message: "未配置 MailPass"}
	}

	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	d := &net.Dialer{Timeout: 3 * time.Second}

	if useSSL {
		conn, err := tls.DialWithDialer(d, "tcp", addr, &tls.Config{ServerName: host})
		if err != nil {
			return types.SysCheckItem{Ok: false, Value: addr, Message: err.Error()}
		}
		_ = conn.Close()
		return types.SysCheckItem{Ok: true, Value: addr, Message: "可连接(SSL)"}
	}

	conn, err := d.Dial("tcp", addr)
	if err != nil {
		return types.SysCheckItem{Ok: false, Value: addr, Message: err.Error()}
	}
	_ = conn.Close()
	return types.SysCheckItem{Ok: true, Value: addr, Message: "可连接"}
}

func (l *SysCheckLogic) checkQdrant() types.SysCheckItem {
	cfg := l.svcCtx.Config.Qdrant
	if cfg.Host == "" || cfg.Port == 0 {
		return types.SysCheckItem{Ok: false, Value: "", Message: "未配置"}
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	if l.svcCtx.QdrantClient == nil {
		return types.SysCheckItem{Ok: false, Value: addr, Message: "客户端未初始化"}
	}
	_, err := l.svcCtx.QdrantClient.HealthCheck(l.ctx)
	if err != nil {
		return types.SysCheckItem{Ok: false, Value: addr, Message: err.Error()}
	}
	return types.SysCheckItem{Ok: true, Value: addr, Message: "可访问"}
}

func (l *SysCheckLogic) checkRedis() types.SysCheckItem {
	if l.svcCtx.RedisClient == nil {
		return types.SysCheckItem{Ok: false, Value: "", Message: "未配置"}
	}
	addr := ""
	if len(l.svcCtx.Config.CacheRedis) > 0 {
		addr = l.svcCtx.Config.CacheRedis[0].Host
	}
	ok := l.svcCtx.RedisClient.Ping()
	if ok {
		return types.SysCheckItem{Ok: true, Value: addr, Message: "可访问"}
	}
	return types.SysCheckItem{Ok: false, Value: addr, Message: "Ping 失败"}
}

// mysqlDSNRe 从 DataSource 解析 tcp(host:port)/dbname
var mysqlDSNRe = regexp.MustCompile(`tcp\(([^)]+)\)/([^?&\s]+)`)

func (l *SysCheckLogic) checkMysql() types.SysCheckItem {
	if l.svcCtx.Mysql == nil {
		return types.SysCheckItem{Ok: false, Value: "", Message: "未配置"}
	}
	ds := l.svcCtx.Config.MySQL.DataSource
	value := ds
	if matches := mysqlDSNRe.FindStringSubmatch(ds); len(matches) >= 3 {
		hostPort := matches[1]
		db := matches[2]
		// hostPort 可能是 "host:port" 或 "host"
		value = fmt.Sprintf("host=%s db=%s", hostPort, db)
		if i := strings.LastIndex(hostPort, ":"); i >= 0 {
			value = fmt.Sprintf("host=%s port=%s db=%s", hostPort[:i], hostPort[i+1:], db)
		}
	}
	_, err := l.svcCtx.Mysql.ExecCtx(l.ctx, "SELECT 1")
	if err != nil {
		return types.SysCheckItem{Ok: false, Value: value, Message: err.Error()}
	}
	return types.SysCheckItem{Ok: true, Value: value, Message: "可访问"}
}

func (l *SysCheckLogic) checkMinerU() types.SysCheckItem {
	url := l.svcCtx.Config.MinerU.URL
	if url == "" {
		return types.SysCheckItem{Ok: false, Value: "", Message: "未配置"}
	}
	base := strings.TrimSuffix(url, "/")
	for _, path := range []string{"/health", "/"} {
		u := base + path
		client := &http.Client{Timeout: 3 * time.Second}
		req, err := http.NewRequestWithContext(l.ctx, http.MethodGet, u, nil)
		if err != nil {
			continue
		}
		res, err := client.Do(req)
		if err != nil {
			continue
		}
		res.Body.Close()
		if res.StatusCode < 500 {
			return types.SysCheckItem{Ok: true, Value: url, Message: "可访问"}
		}
	}
	return types.SysCheckItem{Ok: false, Value: url, Message: "请求失败或超时"}
}
