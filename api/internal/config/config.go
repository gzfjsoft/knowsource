package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/rest"
)

type EmailConfig struct {
	VerifyMailTitle           string
	VerifyMailContent         string
	LoginMailTitle            string
	LoginMailContent          string
	ForgetPasswordMailTitle   string
	ForgetPasswordMailContent string
	StopInstanceMailTitle     string
	StopInstanceMailContent   string
}

type Config struct {
	rest.RestConf

	MySQL struct {
		DataSource string
	}

	CacheRedis cache.CacheConf

	SMS struct {
		Supplier string //alibaba, volc, baishan
		// TemplateCode
		LoginTemplateCode         string // 登录
		RegTemplateCode           string // 注册账号
		ApproveRegTemplateCode    string // 注册通过审核
		DisapproveRegTemplateCode string // 注册未通过审核
		CompanyName               string // 公司名称
		Volc struct {
			AccessKey  string
			SecretKey  string
			SmsAccount string
			Sign       string
			TemplateID string
		}
		Baishan struct {
			Token    string
			Template string
		}
	}

	Boot struct {
		Url string
	}

	Fca struct {
		Url string
	}

	FilesRoot string

	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	Salt string

	AliPay struct {
		AppId      string
		PrivateKey string
		PublicKey  string
		NotifyUrl  string
		ReturnUrl  string
	}

	WeixinPay struct {
		AppId          string
		MerchantId     string
		MerchantKey    string
		SerialNumber   string
		NotifyUrl      string
		PrivateKeyPath string
	}

	WeixinFWH struct {
		AppId          string
		AppSecret      string
		Token          string
		EncodingAESKey string
		CallbackUrl    string
	}

	WeixinOpen struct {
		AppId          string
		AppSecret      string
		Token          string
		EncodingAESKey string
	}

	WeixinMiniApp struct {
		AppId          string
		AppSecret      string
		Token          string
		EncodingAESKey string
	}

	Mail struct {
		MailAccount string
		MailHost    string
		MailPass    string
		MailPort    int
		MailSSL     bool // 是否使用 SSL（一般 465 需要）
	}
	UploadPath   string
	DownloadPath string
	BucketPath   string

	Google struct {
		Key    string
		Client string
	}

	X struct {
		ClientId     string
		ClientSecret string
	}

	AdminJWT string

	Aliyun struct {
		AccessKeyId     string
		AccessKeySecret string
	}

	Qdrant struct {
		Host             string
		Port             int
		CollectionPrefix string
	}

	// Document 分块参数，未配置时默认 ChunkSize=5000, ChunkOverlap=100
	Document struct {
		ChunkSize    int // 默认 5000
		ChunkOverlap int //  默认 100
	}

	IsAllFile int64

	IsDebug int

	Knowdata struct {
		Host                   string
		KnowledgeFilePath      string
		OfflinePackageFilePath string
		TempFilePath           string
		MarkdownPath           string
		WatermarkFile          string
		DocumentPath           string
	}

	MinerU struct {
		URL string // MinerU API 地址，如 http://127.0.0.1:8100
	}

	// Rag RAG 服务配置，支持按 Type 切换 vllm/llama.cpp 等实现
	Rag struct {
		RerankerUrl    string // 重排服务地址，如 http://106.55.186.253:6243
		RerankerType   string // vllm 或 llama.cpp
		EmbeddingsUrl  string
		EmbeddingsType string // vllm 或 ollama
	}
	Llm struct {
		CompletionUrl  string
		CompletionType string // vllm 或 ollama
	}
	RAGURL             string
	SensitiveWordsFile string

	DefaultUserRoleId int64

	Telnet struct {
		Enable bool
		Addr   string
	}
}
