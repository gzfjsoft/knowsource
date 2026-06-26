package knowsource

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/asynctasksignal"
	"knowsource/common/constants"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConvertDocumentToMDLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// convert document to markdown
func NewConvertDocumentToMDLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConvertDocumentToMDLogic {
	return &ConvertDocumentToMDLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ConvertDocumentToMDLogic) ConvertDocumentToMD(req *types.ConvertDocumentToMDRequest) (resp *types.ConvertDocumentToMDResponse, err error) {
	// 异步提交转换任务（识别文字）；实际转换由 async_task worker 完成。
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.UnauthorizedCode,
				Message: "clientId不能为空，请重新登录",
			},
		}, nil
	}
	if req.Id <= 0 {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ParameterErrorCode,
				Message: "id不能为空",
			},
		}, nil
	}

	rawDoc, findErr := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if findErr != nil {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询文档失败",
				Info:    findErr.Error(),
			},
		}, nil
	}
	currentStatus := strings.TrimSpace(rawDoc.Status)
	allowReExtract := currentStatus == constants.RawDocumentsStatusExtractedNotInDB || currentStatus == ""
	if rawDoc.IsToMd == 1 && !allowReExtract {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "文档已识别",
			},
			Data: &types.ConvertDocumentToMDData{
				Id:       rawDoc.Id,
				FileName: rawDoc.FileName,
				Content:  rawDoc.Content,
			},
		}, nil
	}

	taskModel := model.NewAsyncTaskModel(l.svcCtx.Mysql)
	active, aErr := taskModel.FindActiveByTaskTypeAndSourceId(l.ctx, clientId, constants.AsyncTaskTypeRawDocumentsConvertMD, rawDoc.Id)
	if aErr != nil {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询识别任务失败",
				Info:    aErr.Error(),
			},
		}, nil
	}
	if active != nil {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "正在提取文字...",
			},
			Data: &types.ConvertDocumentToMDData{
				Id:       rawDoc.Id,
				FileName: rawDoc.FileName,
				Content:  "",
			},
		}, nil
	}

	_, _ = UpdateRawDocumentStatus(l.ctx, l.svcCtx, clientId, rawDoc.Id, constants.RawDocumentsStatusExtracting, "")

	_, cErr := taskModel.CreateWithClientId(l.ctx, clientId, constants.AsyncTaskTypeRawDocumentsConvertMD, fmt.Sprintf("识别文字:%s", rawDoc.FileName), rawDoc.Id, rawDoc.FileName)
	if cErr != nil {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "创建识别任务失败",
				Info:    cErr.Error(),
			},
		}, nil
	}
	_ = asynctasksignal.NotifyPending(l.ctx, l.svcCtx.RedisClient, clientId)

	return &types.ConvertDocumentToMDResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "已提交识别任务，后台处理中",
		},
		Data: &types.ConvertDocumentToMDData{
			Id:       rawDoc.Id,
			FileName: rawDoc.FileName,
			Content:  "",
		},
	}, nil
}

func (l *ConvertDocumentToMDLogic) ConvertDocumentToMDSync(req *types.ConvertDocumentToMDRequest, asyncTaskId int64) (resp *types.ConvertDocumentToMDResponse, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)

	// 检查配置
	if l.svcCtx.Config.MinerU.URL == "" {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "MinerU 服务未配置",
			},
		}, nil
	}

	// 获取文件根目录
	filesRoot := l.svcCtx.Config.Knowdata.DocumentPath
	if filesRoot == "" {
		filesRoot = l.svcCtx.Config.FilesRoot
	}

	// 查询文档
	var rawDoc *model.RawDocuments
	if clientId != "" {
		rawDoc, err = l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	} else {
		rawDoc, err = l.svcCtx.RawDocumentsModel.FindOne(l.ctx, req.Id)
	}
	if err != nil {
		l.Errorf("查询文档失败: %v", err)
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "查询文档失败",
				Info:    err.Error(),
			},
		}, nil
	}

	if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
		return nil, cancelErr
	}

	// 提取文字开始
	if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
		return nil, cancelErr
	}
	_ = MarkRawDocumentExtracting(l.ctx, l.svcCtx, clientId, rawDoc.Id)

	// 状态为「已提取文字未审核入库」或空状态（列表会推断为该状态）时允许重新识别。
	currentStatus := strings.TrimSpace(rawDoc.Status)
	allowReExtract := currentStatus == constants.RawDocumentsStatusExtractedNotInDB || currentStatus == ""
	if rawDoc.IsToMd == 1 && !allowReExtract {
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.SuccessCode,
				Message: "文档已转换",
			},
			Data: &types.ConvertDocumentToMDData{
				Id:       rawDoc.Id,
				FileName: rawDoc.FileName,
				Content:  rawDoc.Content,
			},
		}, nil
	}

	// 创建文档转换服务
	converter := utils.NewDocumentConverter(l.svcCtx.Config.MinerU.URL, filesRoot)

	// 确定原文件的完整路径
	var originalFilePath string
	originalFilePath = filepath.Clean(rawDoc.FilePath)
	if _, err = os.Stat(originalFilePath); err != nil {
		// 文件不存在，尝试拼接 FilesRoot
		originalFilePath = filepath.Clean(filepath.Join(filesRoot, rawDoc.FilePath))
		if _, err = os.Stat(originalFilePath); err != nil {
			// 两种方式都失败，返回错误
			l.Errorf("原文件不存在: 尝试路径1=%s, 尝试路径2=%s", filepath.Clean(rawDoc.FilePath), originalFilePath)
			return &types.ConvertDocumentToMDResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "原文件不存在",
					Info:    err.Error(),
				},
			}, nil
		}
	}

	// 获取原文件所在目录
	originalFileDir := filepath.Dir(originalFilePath)

	// 转换文档
	err = converter.ConvertDocumentToMD(l.ctx, rawDoc)
	if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
		return nil, cancelErr
	}
	if err != nil {
		l.Errorf("转换文档失败: %v", err)
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "转换文档失败",
				Info:    err.Error(),
			},
		}, nil
	}

	if cancelErr := l.checkConvertTaskCanceled(asyncTaskId, rawDoc.Id); cancelErr != nil {
		return nil, cancelErr
	}

	// 同时更新 content 和 content_org（首次转换时两者相同）
	rawDoc.ContentOrg = rawDoc.Content
	rawDoc.Status = constants.RawDocumentsStatusExtractedNotInDB

	// 更新数据库
	err = l.svcCtx.RawDocumentsModel.Update(l.ctx, rawDoc)
	if err != nil {
		l.Errorf("更新文档失败: %v", err)
		return &types.ConvertDocumentToMDResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "更新文档失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 将 MD 文件保存到对应目录（类似 ZIP 转换的方式）
	// 例如：测试xls转MD.xlsx -> 测试xls转MD.xlsx.file/测试xls转MD.md
	baseName := filepath.Base(rawDoc.FileName)
	ext := filepath.Ext(baseName)
	baseNameWithoutExt := baseName
	if ext != "" {
		baseNameWithoutExt = baseName[:len(baseName)-len(ext)]
	}

	// 创建 .file 目录（例如：测试xls转MD.xlsx.file）
	var mdDirName string
	if ext != "" && len(ext) > 1 {
		mdDirName = baseNameWithoutExt + ext + ".file"
	} else {
		mdDirName = baseNameWithoutExt + ".file"
	}
	mdDir := filepath.Join(originalFileDir, mdDirName)

	// 确保目录存在
	if err := os.MkdirAll(mdDir, 0755); err != nil {
		l.Errorf("创建 MD 文件目录失败: %v", err)
		// 不返回错误，只记录日志，因为数据库更新已经成功
	} else {
		// 生成 MD 文件名（例如：测试xls转MD.md）
		mdFileName := baseNameWithoutExt + ".md"
		mdFilePath := filepath.Join(mdDir, mdFileName)

		// 保存 MD 文件
		if err := os.WriteFile(mdFilePath, []byte(rawDoc.Content), 0644); err != nil {
			l.Errorf("保存 MD 文件失败: %v", err)
			// 不返回错误，只记录日志，因为数据库更新已经成功
		} else {
			l.Infof("成功保存 MD 文件: %s", mdFilePath)
		}
	}

	return &types.ConvertDocumentToMDResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.ConvertDocumentToMDData{
			Id:       rawDoc.Id,
			FileName: rawDoc.FileName,
			Content:  rawDoc.Content,
		},
	}, nil
}
