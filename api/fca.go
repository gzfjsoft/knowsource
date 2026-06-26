package main

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"knowsource/api/executor"
	"knowsource/api/internal/bootstrap"
	"knowsource/api/internal/config"
	"knowsource/api/internal/handler"
	"knowsource/api/internal/logic"
	"knowsource/api/internal/logic/k8s"
	knowsourceLogic "knowsource/api/internal/logic/knowsource"
	"knowsource/api/internal/logic/page"
	"knowsource/api/internal/middleware"
	"knowsource/api/internal/superadmin"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/telnetchat"
	"knowsource/api/internal/utils"
	"knowsource/api/internal/utils/mytail"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/consts"
	"knowsource/model"

	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/router"
	"gopkg.in/yaml.v3"
	// "github.com/zs5460/art"
	// "github.com/gotoeasy/glang/cmn"
)

//go:embed etc/*
var etcFs embed.FS

var TerminalMode bool

var c config.Config
var globalTailer *mytail.LogTailer
var e config.EmailConfig

type ObjGetRequest struct {
	File string `path:"file"`
}

func GetFiles(req *ObjGetRequest, w http.ResponseWriter, r *http.Request) {

	logx.Infof("GetObj %s", req.File)
	fullpath := path.Join(c.FilesRoot, req.File)

	http.ServeFile(w, r, fullpath)

}

func GetStatic(file string, writer http.ResponseWriter) error {

	logx.Infof("GetStatic %s", file)
	body, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	n, err := writer.Write(body)
	if err != nil {
		return err
	}

	if n < len(body) {
		return io.ErrClosedPipe
	}

	return nil
}

func GetUpload(file string, writer http.ResponseWriter) error {

	logx.Infof("GetUpload %s", file)
	body, err := os.ReadFile(path.Join(c.UploadPath, file))
	if err != nil {
		return err
	}

	n, err := writer.Write(body)
	if err != nil {
		return err
	}

	if n < len(body) {
		return io.ErrClosedPipe
	}

	return nil
}

// allowedMdStaticExt 允许 /api/v1/md/ 静态服务返回的扩展名：仅图片和 md 文档
var allowedMdStaticExt = map[string]bool{
	".md": true, ".mdx": true,
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".webp": true, ".svg": true, ".bmp": true, ".ico": true,
}

// getStaticFullPath 解析 /api/v1/static/ 后的路径，禁止路径穿越，返回 static 根下的绝对路径
func getStaticFullPath(relativePath string) (fullPath string, err error) {
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", fmt.Errorf("invalid path")
	}
	root := "static"
	fullPath = filepath.Join(root, clean)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	sep := string(filepath.Separator)
	if absFull != absRoot && !strings.HasPrefix(absFull, absRoot+sep) {
		return "", fmt.Errorf("invalid path")
	}
	return fullPath, nil
}

// getMdStaticFullPath 解析 /api/v1/md/ 后的路径，校验仅允许图片和 md，返回根目录下的绝对路径；禁止路径穿越
func getMdStaticFullPath(relativePath string) (fullPath string, err error) {
	clean := filepath.Clean(relativePath)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("invalid path")
	}
	ext := strings.ToLower(filepath.Ext(clean))
	if !allowedMdStaticExt[ext] {
		return "", fmt.Errorf("forbidden file type: %s", ext)
	}
	root := c.Knowdata.DocumentPath
	if root == "" {
		return "", fmt.Errorf("document path not configured")
	}
	fullPath = filepath.Join(root, filepath.FromSlash(clean))
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", err
	}
	sep := string(filepath.Separator)
	if absFull != absRoot && !strings.HasPrefix(absFull, absRoot+sep) {
		return "", fmt.Errorf("invalid path")
	}
	return fullPath, nil
}

func LoadFromYamlBytes(file []byte, c *config.Config) {

	err := yaml.Unmarshal(file, c)
	if err != nil {
		fmt.Println("Error unmarshalling config file:", err)
		return
	}

}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func main() {

	k8s.DEMO()

	enableFeature := flag.Bool("t", false, "Run in terminal logging mode")
	configFile := flag.String("c", "", "Path to config file (default: knowsource.yaml in current directory)")

	flag.Parse()

	TerminalMode = *enableFeature

	// Get current directory for default config path
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	var mainConfigPath string
	// Check if config file is specified via command line
	if *configFile != "" {
		mainConfigPath = *configFile
		fmt.Println("Using config file from command line:", *configFile)
		file, err := os.ReadFile(*configFile)
		if err != nil {
			fmt.Println("Error reading config file:", err)
			return
		}
		conf.LoadFromYamlBytes(file, &c)
	} else {
		// Determine config file based on dev.ini existence
		devIniPath := filepath.Join(currentDir, "dev.ini")
		var yamlPath string

		if _, err = os.Stat(devIniPath); err == nil {
			// dev.ini exists, use knowsource.dev.yaml
			yamlPath = filepath.Join(currentDir, "knowsource.dev.yaml")
			fmt.Println("Found dev.ini, using knowsource.dev.yaml")
		} else {
			// dev.ini does not exist, use knowsource.yaml
			yamlPath = filepath.Join(currentDir, "knowsource.yaml")
			fmt.Println("dev.ini not found, using knowsource.yaml")
		}

		// Check if the determined config file exists
		if _, err = os.Stat(yamlPath); err == nil {
			mainConfigPath = yamlPath
			fmt.Println("Found config file in:", yamlPath)
			file, err := os.ReadFile(yamlPath)
			if err != nil {
				fmt.Println("Error reading config file:", err)
				return
			}
			conf.LoadFromYamlBytes(file, &c)
		} else {
			//exit
			fmt.Println("can not find knowsource.yaml or knowsource.dev.yaml")
			os.Exit(1)

		}
	}

	// Check if fca-email.yaml exists in current directory
	emailYamlPath := filepath.Join(currentDir, "fca-emails.yaml")
	if _, err = os.Stat(emailYamlPath); err == nil {
		fmt.Println("Found fca-emails.yaml in:", emailYamlPath)

		file, _ := os.ReadFile(emailYamlPath)
		// println(string(file))

		yaml.Unmarshal(file, &e)

		fmt.Println("= Email Config =")
		fmt.Println("")
		yamlData, err := yaml.Marshal(e)
		if err != nil {
			fmt.Println("转换为YAML格式失败:", err)
			return
		}
		fmt.Println(string(yamlData))

		fmt.Printf("[%s][%s]", e.VerifyMailContent, e.ForgetPasswordMailContent)
	}

	if true { //敏感词处理
		wordsFilePath, err := utils.GetSensitiveWordsFilePath()
		if err != nil {
			logx.Errorf("获取敏感词文件路径失败: %v", err)
			return
		}
		if c.SensitiveWordsFile == "" {
			c.SensitiveWordsFile = wordsFilePath
		}
		readFile, err := os.ReadFile(c.SensitiveWordsFile)
		if err != nil {

			logx.Errorf("读取敏感词文件失败: %v", err)

		} else {
			if string(readFile) == "" {
				readFile = []byte(strings.Join(utils.DefaultSensitiveWords, "\n"))
			}
			logx.Infof("读取敏感词文件成功: %s", c.SensitiveWordsFile)
			sensitiveWords := strings.Split(string(readFile), "\n")

			utils.RenewSensitiveWords(sensitiveWords...)
		}
	}

	apifile, _ := etcFs.ReadFile("etc/fca.json")
	page.API_FILE = string(apifile)
	logic.API_FILE = string(apifile)

	fmt.Println("")
	fmt.Println("= Config =")
	fmt.Println("")
	yamlData, err := yaml.Marshal(c)
	if err != nil {
		fmt.Println("转换为YAML格式失败:", err)
		return
	}

	if c.IsDebug == 1 {
		// 打印YAML格式的数据
		fmt.Println(string(yamlData))
	}

	// Configure logx based on the -t flag
	logic.BOOTURL = c.Boot.Url

	logic.TerminalMode = TerminalMode
	if TerminalMode {

		logx.DisableStat()
		logx.SetUp(logx.LogConf{
			Mode:     "console",
			Level:    "info",
			Encoding: "plain",
			Path:     "", // No file path for console logging
		})

		fmt.Println("########### Running in terminal logging mode ###########")
	} else {
		logx.MustSetup(c.Log)
		// 配置文件中已配置 file 模式
		// 再加一路 stdout → systemd journal 自动收集
		logx.AddWriter(logx.NewWriter(os.Stdout))
	}

	// 根据 CompletionUrl 探测 /api/version（ollama）或 /version（vllm）自动确定 CompletionType
	utils.EnsureCompletionType(&c)

	logx.Infof("svc.NewServiceContext")

	svcCtx, svcErr := svc.NewServiceContext(c, e)
	if svcErr != nil {
		bootstrap.ServerStartedInInitMode = true
		logx.Errorf("核心依赖未就绪，进入初始化模式（仅开放配置向导 API）: %v", svcErr)
	} else {
		bootstrap.ServerStartedInInitMode = false
	}

	r := router.NewRouter()
	r.SetNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//if r.URL.Path match api/v1/object/get/

		logx.Infof("SetNotFoundHandler r.URL.Path: %s", r.URL.Path)

		// 处理mytail路由
		if globalTailer != nil {
			// if r.URL.Path == "/api/v1/mytail/logs" && r.Method == "GET" {
			// 	globalTailer.HandleLogPage(w, r)
			// 	return
			// }
			if r.URL.Path == "/api/v1/mytail/ws" && r.Method == "GET" {
				globalTailer.HandleWebSocket(w, r)
				return
			}
			// 处理mytail路由 - 需要认证
			if r.URL.Path == "/api/v1/mytail/logs" && r.Method == "GET" {
				if svcCtx == nil {
					httpx.WriteJson(w, http.StatusServiceUnavailable, response.Error("服务正在初始化，请稍后重试"))
					return
				}
				// 应用认证中间件
				authHandler := svcCtx.Auth(func(w http.ResponseWriter, r *http.Request) {
					// 认证通过后的处理逻辑
					if globalTailer != nil {
						globalTailer.HandleLogPage(w, r)
					} else {
						// 服务未启用
						httpx.WriteJson(w, http.StatusServiceUnavailable, response.Error("日志查看服务未启用"))
					}
				})
				authHandler(w, r)
				return
			}

		}

		if ((strings.HasPrefix(r.URL.Path, "/static/")) || r.URL.Path == "/static") && r.Method == "GET" {

			if r.URL.Path == "/static" {
				http.Redirect(w, r, "/static/index.html", http.StatusSeeOther)
				return
			}

			parts := strings.Split(r.URL.Path[1:], "/")
			if len(parts) >= 2 {
				target := strings.Join(parts[1:], "/") // 去掉前缀 "static"，得到 relative path
				fullpath, err := getStaticFullPath(target)
				if err != nil {
					httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
					return
				}
				if _, err := os.Stat(fullpath); os.IsNotExist(err) {
					httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "File not found", fullpath))
					return
				}
				http.ServeFile(w, r, fullpath)
				return
			}
		}

		if (strings.HasPrefix(r.URL.Path, "/api/v1/static/")) && r.Method == "GET" {

			urlpath := r.URL.Path
			parts := strings.Split(urlpath[1:], "/")
			if len(parts) < 4 {
				httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, "Invalid request param"))
				return
			}
			target := strings.TrimPrefix("/"+strings.Join(parts[3:], "/"), "/")
			fullpath, err := getStaticFullPath(target)
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
				return
			}
			if _, err := os.Stat(fullpath); os.IsNotExist(err) {
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "File not found", fullpath))
				return
			}
			logx.Infof("Get static %s", fullpath)
			http.ServeFile(w, r, fullpath)
			return

		}

		if (strings.HasPrefix(r.URL.Path, "/api/v1/md/")) && r.Method == "GET" {

			urlpath := r.URL.Path
			parts := strings.Split(urlpath[1:], "/")
			if len(parts) < 4 {
				httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, "Invalid request param"))
				return
			}
			target := strings.TrimPrefix("/"+strings.Join(parts[3:], "/"), "/")
			fullpath, err := getMdStaticFullPath(target)
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
				return
			}
			if _, err := os.Stat(fullpath); os.IsNotExist(err) {
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "File not found", fullpath))
				return
			}
			http.ServeFile(w, r, fullpath)
			return
		}

		if (strings.HasPrefix(r.URL.Path, "/api/v1/upload/")) && r.Method == "GET" {

			// authHeader := r.Header.Get("Authorization")
			// if authHeader == "" {
			// 	logx.Error("Missing Authorization header")
			// 	http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			// 	return
			// }
			// parts := strings.SplitN(authHeader, " ", 2)
			// if !(len(parts) == 2 && parts[0] == "Bearer") {
			// 	logx.Error("Invalid Authorization header format")
			// 	http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			// 	return
			// }
			// token := parts[1]
			// logx.Infof("token: %s", token)
			// l := logic.NewGetObjLogic(ctx2, ctx, w)

			path := r.URL.Path
			parts := strings.Split(path[1:], "/")
			if len(parts) >= 4 {
				target := "/" + strings.Join(parts[3:], "/")
				fmt.Printf("target: %s\n", target)

				err = GetUpload(target, w)

				if err != nil {
					httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.InvalidRequestParamCodeInHandler, "Error while handle call", err.Error()))
				}

				return
			}
		}

		if (strings.HasPrefix(r.URL.Path, "/api/v1/filesget/")) && r.Method == "GET" {

			// authHeader := r.Header.Get("Authorization")
			// if authHeader == "" {
			// 	logx.Error("Missing Authorization header")
			// 	http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			// 	return
			// }

			// parts := strings.SplitN(authHeader, " ", 2)
			// if !(len(parts) == 2 && parts[0] == "Bearer") {
			// 	logx.Error("Invalid Authorization header format")
			// 	http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			// 	return
			// }

			// token := parts[1]
			// logx.Infof("token: %s", token)

			urlpath := r.URL.Path
			logx.Infof("path: %s", urlpath)
			parts := strings.Split(urlpath[1:], "/")

			if len(parts) < 4 {
				httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, "Invalid request param"))
				return
			}

			// id := parts[3]
			target := "/" + strings.Join(parts[3:], "/")
			logx.Infof("target: %s", target)

			code := r.URL.Query().Get("code")
			logx.Infof("code: %s", code)
			var logpath string

			if true {
				if code == "" {
					httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, "Invalid request param"))
					return
				}

				if svcCtx == nil {
					httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InternalServerErrorDuringProcessingCode, "服务正在初始化"))
					return
				}
				logpath, err = svcCtx.RedisClient.Get("FA-" + code)
				if err != nil {
					httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, err.Error()))
					return
				}

				if logpath == "" {
					httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, "not allowed access"))
					return
				}
				logpath = "/" + logpath
			}
			// } else {
			// 	logpath = target
			// }

			if logpath != target {
				httpx.OkJsonCtx(r.Context(), w, response.Fail(response.InvalidRequestParamCodeInHandler, "Not the same file"))
				return
			}
			//not the same

			logx.Infof("GetObj %s", target)
			fullpath := path.Join(c.FilesRoot, target)

			http.ServeFile(w, r, fullpath)

			return

		}

		w.WriteHeader(http.StatusNotFound)
		// w.Write([]byte(r.URL.Path))
	}))
	logx.Infof("rest.MustNewServer")
	server := rest.MustNewServer(c.RestConf,
		rest.WithUnauthorizedCallback(func(w http.ResponseWriter, r *http.Request, err error) {
			res := response.Fail(response.UnauthorizedCode, "未登录")
			bts, _ := json.Marshal(res)
			w.Write(bts)
		}),
		rest.WithCors(),
		rest.WithRouter(r),
	)
	defer server.Stop()

	if false {
		// 测试时关闭权限检查
		server.Use(PermissionCheck(svcCtx))
	}

	bootstrap.RegisterRoutes(server, bootstrap.Options{
		MainPath:  mainConfigPath,
		EmailPath: filepath.Join(currentDir, "fca-emails.yaml"),
	})

	if svcErr == nil {
		handler.RegisterHandlers(server, svcCtx)
	} else {
		logx.Infof("已跳过全量业务路由；修复 MySQL/Redis 并保存配置后请重启服务")
	}

	// 初始化日志监控器
	if svcErr == nil {
		logFile := filepath.Join(c.Log.Path, "access.log")

		// 使用默认日志路径
		if _, err = os.Stat(c.Log.Path); os.IsNotExist(err) {
			os.MkdirAll(c.Log.Path, 0755)
		}
		tailer, err := mytail.NewLogTailer(logFile)
		if err != nil {
			logx.Errorf("创建日志监控器失败: %v", err)
		} else {
			// 启动日志监控
			logx.Infof("启动日志监控")
			go func() {
				if err = tailer.Start(context.Background()); err != nil {
					logx.Errorf("启动日志监控失败: %v", err)
				}
			}()

			// 注册mytail路由到NotFoundHandler中
			logx.Infof("日志查看器已启动，访问地址: http://%s:%d/api/v1/mytail/logs", c.Host, c.Port)

			// 保存tailer引用以便在NotFoundHandler中使用
			globalTailer = tailer
		}
		logx.Info("清理 dept_document_type：删除 document_type 中不存在的 client_id + doc_type_code 关联")
		if n, orphanErr := svcCtx.DeptDocumentTypeModel.DeleteOrphansNotInDocumentType(context.Background()); orphanErr != nil {
			logx.Errorf("清理 dept_document_type 孤儿行失败: %v", orphanErr)
		} else if n > 0 {
			logx.Infof("已删除 %d 条 dept_document_type 孤儿记录（document_type 无对应类型）", n)
		}

		logx.Info("确保所有租户的 superadmin")
		// 获取所有 client_id
		var clientIds []string
		qErr := svcCtx.Mysql.QueryRowsCtx(context.Background(), &clientIds, "SELECT client_id FROM client")
		if qErr == nil {
			for _, cid := range clientIds {
				_, err = superadmin.EnsureSuperadmin(context.Background(), svcCtx, cid)
				if err != nil {
					logx.Errorf("[%s] 确保 superadmin 失败: %v", cid, err)
				}
			}
		} else {
			logx.Errorf("查询所有 client 失败: %v", qErr)
			// 至少尝试一下 demo 租户（作为降级方案，或者如果连 client 表都还没查到）
			_, _ = superadmin.EnsureSuperadmin(context.Background(), svcCtx, "demo")
		}
	}
	//server.PrintRoutes()
	// 确保管理者有全部权限
	// isKnowdata := true
	// if isKnowdata {
	// 	err = ctx.RoleModel.InitWithAdminRole(context.Background())
	// 	if err != nil {
	// 		logx.Errorf("初始化超级管理员角色失败: %v", err)
	// 	}
	// }

	// 保存路由到数据库
	if false {
		var permissions []string
		for _, route := range server.Routes() {
			path := strings.Replace(route.Path, "/api/v1/", "", -1)
			path = strings.Replace(path, "/", ".", -1)
			// fmt.Printf("insert router(path) value(\"%s\") \n", path)
			if path == "." {
				continue
			}
			permissions = append(permissions, path)

		}

		sort.Slice(permissions, func(i, j int) bool {
			return permissions[i] < permissions[j]
		})

		for _, permission := range permissions {
			_, err = svcCtx.PermissionsModel.Insert(context.Background(), &model.Permissions{
				PermissionName: permission,
				Description:    sql.NullString{String: permission, Valid: true},
			})
			if err != nil {
				logx.Errorf("Failed to insert permission: %v", err)
			}
		}
	}

	if c.IsDebug != 1 && false {
		// 使用费用计算
		// fc := NewFeeCalculator(ctx)
		// StartFeeTimer(fc)

		// 或使用内置限流器
		periodLimit := middleware.NewPeriodLimit(svcCtx.RedisClient, 100, 60)
		server.Use(periodLimit.Handle)
	}

	parameterCheck := middleware.NewParameterCheck()
	server.Use(parameterCheck.Handle)

	// Add CORS handler
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		return http.StatusOK, err.Error()
	})

	// glc 配置
	// cmn.SetGlcClient(cmn.NewGlcClient(&cmn.GlcOptions{
	// 	ApiUrl:           "http://glc.edge.baishancloud.com/", // 日志中心的添加日志接口地址，默认取环境变量GLC_API_URL
	// 	Enable:           "true",                              // 是否开启发送到日志中心(true/false)，默认取环境变量GLC_ENABLE，默认false
	// 	System:           "fca-peta",                          // 系统名（对应日志中心检索页面的分类栏），默认取环境变量GLC_SYSTEM，默认default
	// 	ApiKey:           "X-GLC-AUTH:glogcenter",             // 日志中心的ApiKey，默认取环境变量GLC_API_KEY，默认X-GLC-AUTH:glogcenter
	// 	EnableConsoleLog: "true",                              // 是否禁止打印控制台日志(true/false)，默认取环境变量GLC_ENABLE_CONSOLE_LOG，默认true
	// 	LogLevel:         "DEBUG",                             // 能输出的日志级别（DEBUG/INFO/WARN/ERROR），默认取环境变量GLC_LOG_LEVEL，默认DEBUG
	// 	PrintSrcLine:     true,                                // 是否添加打印调用的文件行号，默认false
	// }))

	// startInfo := fmt.Sprintf("Starting fca api server at %s:%d...\n", c.Host, c.Port)
	// cmn.Info(startInfo)

	logx.Infof("Starting coolpeople server at %s:%d...", c.Host, c.Port)

	// Telnet chat server (same process). Calls existing HTTP API with AdminJWT => superadmin.
	if c.Telnet.Enable {
		addr := strings.TrimSpace(c.Telnet.Addr)
		if addr == "" {
			addr = ":2323"
		}
		host := strings.TrimSpace(c.Host)
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "127.0.0.1"
		}
		chatURL := fmt.Sprintf("http://%s:%d/api/ai/session/chat", host, c.Port)
		routerURL := utils.GetCompletionURL(&c)
		go func() {
			if err := telnetchat.Start(context.Background(), telnetchat.Options{
				ListenAddr: addr,
				ChatURL:    chatURL,
				RouterURL:  routerURL,
				AdminJWT:   c.AdminJWT,
				ClientID:   consts.ONLY_ADMIN,
				FilesRoot:  firstNonEmpty(c.Knowdata.DocumentPath, c.FilesRoot),
			}); err != nil {
				logx.Errorf("telnet chat server stopped: %v", err)
			}
		}()
		logx.Infof("telnet chat enabled at %s (api=%s)", addr, chatURL)
	}

	if svcErr == nil && len(c.CacheRedis) > 0 {
		utils.InitCaptchaStore(c.CacheRedis[0].Host, c.CacheRedis[0].Pass)

		// 预热数据
		preheatData(svcCtx)

		// 启动时预拉取 vLLM 模型列表到内存，供 chat/embedding/rerank 解析使用
		if l := knowsourceLogic.NewSysCheckLogic(context.Background(), svcCtx); l != nil {
			if resp, err := l.SysCheck(); err == nil && resp != nil && resp.Data != nil {
				utils.LLMModelStore.Update(resp.Data.Vllmchat.ModelIds, resp.Data.Vllmembedding.ModelIds, resp.Data.Vllmreranker.ModelIds)
				logx.Info("LLM model store updated from sys check at startup")
			}
		}

		// 启动异步任务处理器
		initAsyncTaskProcessor(svcCtx)

		go utils.GetMessageFromQueue(svcCtx.RedisClient, "sync_to_qdrant", svcCtx.RawDocumentsModel)
	}

	// 最后一条消息为专家发出，且超过24h无消息交互，则系统自动变为已解决，评分默认5分

	server.Start()
}

func getUIDFromJWT(tokenString string) (int64, string, error) {
	// 密钥要和生成时使用的一致
	secretKey := []byte(c.Auth.AccessSecret)
	// 解析JWT令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return 0, "", err
	}
	// 验证解析后的Token是否有效并且是使用期望的签名算法

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		yamlData, _ := yaml.Marshal(claims)
		fmt.Println(string(yamlData))

		fmt.Println(reflect.TypeOf(claims["uid"]))
		fmt.Println(reflect.TypeOf(claims["roles"]))

		roles := claims["roles"].(string)
		fuid, _ := claims["uid"].(float64)
		uid := int64(fuid)

		fmt.Println("uid: ", uid)
		if ok {
			return uid, roles, nil
		}
		//return 0, fmt.Errorf("uid not found in claims")
	}
	return 0, "", errors.New("invalid token")
}

func PermissionCheck(ctx *svc.ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Skip permission check for certain paths
			if strings.HasPrefix(r.URL.Path, "/api/v1/upload/") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/object/") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/md/") {
				next.ServeHTTP(w, r)
				return
			}

			var uid int64
			uid = 0
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && parts[0] == "Bearer") {
				next.ServeHTTP(w, r)
				return
			}

			// Extract UID from token
			uid, roles, err := getUIDFromJWT(parts[1])
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, response.Fail(response.UnauthorizedCode, "Invalid token"))
				return
			}

			if roles == consts.ONLY_ADMIN || roles == consts.SUPER_ADMIN {
				next.ServeHTTP(w, r)
				return
			}

			// Convert URL path to permission format
			path := strings.Replace(r.URL.Path, "/api/v1/", "", -1)
			path = strings.Replace(path, "/", ".", -1)
			if path == "." {
				next.ServeHTTP(w, r)
				return
			}
			logx.Infof("path: %s %d", path, uid)

			// 检查用户是否有权限
			hasPermission, err := ctx.PermissionsModel.HasPermission(context.Background(), uid, path)
			if err != nil {
				httpx.OkJsonCtx(r.Context(), w, response.FailWithInfo(response.UnauthorizedCode, "Invalid token", err.Error()))
				return
			}
			if !hasPermission {
				httpx.OkJsonCtx(r.Context(), w, response.Fail(response.UnauthorizedCode, "没有权限访问: "+path))
				return
			}

			// TODO: Check if user's role has permission to access this path
			// This part needs to be implemented based on your database schema
			// Typically would involve:
			// 1. Get user's role(s)
			// 2. Check if any of user's roles have the required permission
			// 3. If not, return unauthorized

			next.ServeHTTP(w, r)
		}
	}
}

// 初始化异步任务处理器
func initAsyncTaskProcessor(svcCtx *svc.ServiceContext) {
	// 创建任务处理器
	processor := executor.NewTaskProcessor(svcCtx)

	// 空闲时仅轮询 Redis pending key（默认 1000ms），有任务由插入时 SET key 唤醒
	processor = processor.WithRedisPollInterval(1000 * time.Millisecond)
	// 单任务最长执行 30 分钟，超时后自动中断并标记 canceled
	processor = processor.WithTaskTimeout(30 * time.Minute)

	// 创建任务执行器映射
	executors := executor.TaskExecutorMap{
		constants.AsyncTaskTypeRawDocumentsConvertZIP: executor.NewRawDocumentsConvertZIPExecutor(svcCtx),
		constants.AsyncTaskTypeRawDocumentsConvertMD:  executor.NewRawDocumentsConvertMDExecutor(svcCtx),
		constants.AsyncTaskTypeRawDocumentsAuditIn:    executor.NewRawDocumentsAuditInExecutor(svcCtx),
	}

	// 启动任务处理器
	go processor.Start(executors)

	logx.Info("Async task processor initialized and started")
}

// 预热数据
func preheatData(svcCtx *svc.ServiceContext) {
	// ctx := context.Background()
	// 清理过期数据
	// expiredTime := time.Now().Add(-time.Hour * 24).Unix() // 获取一天前的数据
	// svcCtx.UserPointsRankingModel.CleanExpiredRanking(ctx, "", expiredTime)
	// 排行榜数据预热
	// svcCtx.UserPointsRankingModel.InitRankData(ctx, "", "")

	// 同步所有租户的权限
	ctx := context.Background()
	clients, err := svcCtx.ClientModel.FindAll(ctx)
	if err != nil {
		logx.Errorf("获取所有租户失败: %v", err)
		return
	}

	logx.Infof("开始同步所有租户权限，共 %d 个租户", len(clients))
	for _, client := range clients {
		if err := superadmin.SyncPermissions(ctx, svcCtx, client.ClientId); err != nil {
			logx.Errorf("同步租户 %s 权限失败: %v", client.ClientId, err)
		} else {
			logx.Infof("同步租户 %s 权限完成", client.ClientId)
		}
	}
	logx.Info("所有租户权限同步完成")
}
