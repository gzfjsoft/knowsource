package bootstrap

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"knowsource/api/internal/config"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"gopkg.in/yaml.v3"
)

// Payload GET/POST 与前端交互的完整配置体。
type Payload struct {
	MainConfig  config.Config      `json:"mainConfig"`
	EmailConfig config.EmailConfig `json:"emailConfig"`
	MainPath    string             `json:"mainPath"`
	EmailPath   string             `json:"emailPath"`
	EmailExists bool               `json:"emailExists"`
}

// StatusData 引导页状态。
type StatusData struct {
	MysqlOk  bool   `json:"mysqlOk"`
	RedisOk  bool   `json:"redisOk"`
	CoreOk   bool   `json:"coreOk"`
	InitMode bool   `json:"initMode"`
	AppReady bool   `json:"appReady"`
	MysqlMsg string `json:"mysqlMsg,omitempty"`
	RedisMsg string `json:"redisMsg,omitempty"`
	MainPath string `json:"mainPath,omitempty"`
}

// Options 注册路由时的依赖（主配置路径、邮件配置路径）。
type Options struct {
	MainPath  string
	EmailPath string
}

func statusHandler(opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := os.ReadFile(opts.MainPath)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, "读取主配置失败", err.Error()))
			return
		}
		var cfg config.Config
		// 与 fca.go 一致：用 go-zero conf 解析，才能识别 YAML 里常见的 Mysql/CacheRedis 等键名
		if err := conf.LoadFromYamlBytes(body, &cfg); err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ParameterErrorCode, "主配置 YAML 无效", err.Error()))
			return
		}
		mysqlOk, redisOk, mysqlMsg, redisMsg := CheckCore(&cfg)
		coreOk := mysqlOk && redisOk
		initMode := ServerStartedInInitMode
		appReady := coreOk && !initMode
		httpx.OkJsonCtx(r.Context(), w, response.OK(
			StatusData{
				MysqlOk:  mysqlOk,
				RedisOk:  redisOk,
				CoreOk:   coreOk,
				InitMode: initMode,
				AppReady: appReady,
				MysqlMsg: mysqlMsg,
				RedisMsg: redisMsg,
				MainPath: opts.MainPath,
			},
		))
	}
}

func getConfigHandler(opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mainBody, err := os.ReadFile(opts.MainPath)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, "读取主配置失败", err.Error()))
			return
		}
		var mainCfg config.Config
		if err := conf.LoadFromYamlBytes(mainBody, &mainCfg); err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ParameterErrorCode, "主配置 YAML 无效", err.Error()))
			return
		}
		out := Payload{
			MainConfig: mainCfg,
			MainPath:   opts.MainPath,
			EmailPath:  opts.EmailPath,
		}
		if b, err := os.ReadFile(opts.EmailPath); err == nil {
			var em config.EmailConfig
			if yaml.Unmarshal(b, &em) == nil {
				out.EmailConfig = em
				out.EmailExists = true
			}
		}
		httpx.OkJsonCtx(r.Context(), w, response.OK(out))
	}
}

type saveRequest struct {
	MainConfig  config.Config       `json:"mainConfig"`
	EmailConfig *config.EmailConfig `json:"emailConfig,omitempty"`
	SaveEmail   bool                `json:"saveEmail"`
}

func saveConfigHandler(opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.Fail(response.ParameterErrorCode, "读取请求体失败"))
			return
		}
		var req saveRequest
		if err := json.Unmarshal(body, &req); err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ParameterErrorCode, "JSON 无效", err.Error()))
			return
		}
		mainBytes, err := yaml.Marshal(&req.MainConfig)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, "序列化主配置失败", err.Error()))
			return
		}
		if err := writeFileAtomic(opts.MainPath, mainBytes, 0o644); err != nil {
			logx.Errorf("bootstrap save main yaml: %v", err)
			httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, "写入主配置失败", err.Error()))
			return
		}
		if req.SaveEmail && req.EmailConfig != nil {
			emailBytes, err := yaml.Marshal(req.EmailConfig)
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, "序列化邮件配置失败", err.Error()))
				return
			}
			if err := writeFileAtomic(opts.EmailPath, emailBytes, 0o644); err != nil {
				logx.Errorf("bootstrap save email yaml: %v", err)
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.ServerErrorCode, "写入 fca-emails.yaml 失败", err.Error()))
				return
			}
		}
		httpx.OkJsonCtx(r.Context(), w, response.Success("已保存配置文件，若曾处于初始化模式或修改了数据库与 Redis，请重启本服务", map[string]string{
			"mainPath":  opts.MainPath,
			"emailPath": opts.EmailPath,
		}))
	}
}

func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	tmp, err := os.CreateTemp(dir, "."+base+".*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	return os.Rename(tmpPath, path)
}
