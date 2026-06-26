// Code scaffolded by goctl. Safe to edit.
// AI对话上传临时文档（txt/docx/pdf），识别内容并写入对话缓存

package ai

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxUploadSize = 20 << 20 // 20 MB

type AIChatUploadDocumentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewAIChatUploadDocumentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AIChatUploadDocumentLogic {
	return &AIChatUploadDocumentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AIChatUploadDocumentLogic) SetRequest(r *http.Request) {
	l.r = r
}

func (l *AIChatUploadDocumentLogic) AIChatUploadDocument(req *types.AIChatUploadDocumentRequest) (resp *types.AIChatUploadDocumentResponse, err error) {
	if l.r == nil {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.InvalidRequestParamCodeInHandler,
				Message: "请求无效",
				Info:    "缺少请求体",
			},
		}, nil
	}

	empCode, _ := l.ctx.Value("empCode").(string)
	if empCode == "" {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "未登录",
			},
		}, nil
	}

	if err := l.r.ParseMultipartForm(maxUploadSize); err != nil {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.InvalidRequestParamCodeInHandler,
				Message: "解析上传失败",
				Info:    err.Error(),
			},
		}, nil
	}

	file, header, err := l.r.FormFile("file")
	if err != nil {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.InvalidRequestParamCodeInHandler,
				Message: "请选择文件",
				Info:    err.Error(),
			},
		}, nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !utils.AllowedUploadExt(ext) {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.InvalidRequestParamCodeInHandler,
				Message: "仅支持 .txt、.pdf、.docx 文件",
				Info:    ext,
			},
		}, nil
	}

	// 保存到临时文件
	tmpDir := l.svcCtx.Config.UploadPath
	if tmpDir == "" {
		tmpDir = os.TempDir()
	}
	tmpDir = filepath.Join(tmpDir, "ai_chat_upload")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		l.Errorf("创建临时目录失败: %v", err)
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "服务器错误",
				Info:    err.Error(),
			},
		}, nil
	}

	tmpPath := filepath.Join(tmpDir, time.Now().Format("20060102150405")+"_"+header.Filename)
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "创建临时文件失败",
				Info:    err.Error(),
			},
		}, nil
	}
	_, err = io.Copy(tmpFile, file)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "保存文件失败",
				Info:    err.Error(),
			},
		}, nil
	}
	defer os.Remove(tmpPath)

	opts := &utils.ExtractOptions{
		MinerUURL: l.svcCtx.Config.MinerU.URL,
		FilesRoot: l.svcCtx.Config.FilesRoot,
	}
	if opts.FilesRoot == "" && l.svcCtx.Config.Knowdata.DocumentPath != "" {
		opts.FilesRoot = l.svcCtx.Config.Knowdata.DocumentPath
	}

	content, err := utils.ExtractTextFromFileWithOptions(l.ctx, tmpPath, opts)
	if err != nil {
		return &types.AIChatUploadDocumentResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "识别文档内容失败",
				Info:    err.Error(),
			},
		}, nil
	}

	item := utils.AIChatCachedDocItem{
		Filename: header.Filename,
		Content:  content,
	}
	if l.svcCtx.RedisClient != nil {
		if err := utils.AIChatDocCacheSet(l.svcCtx.RedisClient, empCode, item); err != nil {
			l.Errorf("写入 Redis 缓存失败: %v", err)
			return &types.AIChatUploadDocumentResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "缓存失败",
					Info:    err.Error(),
				},
			}, nil
		}
	}

	return &types.AIChatUploadDocumentResponse{
		Response: types.Response{
			Code:    200,
			Message: "success",
		},
		Data: &types.AIChatUploadDocumentData{
			Filename: header.Filename,
			Message:  "已识别并缓存，发送下一条消息时将作为参考",
		},
	}, nil
}
