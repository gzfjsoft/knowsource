package knowdata

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ChangeRawDocumentsTagLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更改原始文档标签
func NewChangeRawDocumentsTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChangeRawDocumentsTagLogic {
	return &ChangeRawDocumentsTagLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChangeRawDocumentsTagLogic) ChangeRawDocumentsTag(req *types.ChangeRawDocumentsTagRequest) (resp *types.Response, err error) {
	clientId, _ := l.ctx.Value("clientId").(string)
	clientId = strings.TrimSpace(clientId)
	if clientId == "" {
		return &types.Response{
			Code:    response.UnauthorizedCode,
			Message: "clientId不能为空，请重新登录",
		}, nil
	}

	// 检查 ID 是否有效
	if req.Id <= 0 {
		return &types.Response{
			Code:    response.ParameterErrorCode,
			Message: "ID 不能为空或无效",
		}, nil
	}

	// 查询文档是否存在
	doc, err := l.svcCtx.RawDocumentsModel.FindOneByClientId(l.ctx, clientId, req.Id)
	if err != nil {
		if err == sqlx.ErrNotFound || errors.Is(err, model.ErrNotFound) {
			return &types.Response{
				Code:    response.RecordNotExistCode,
				Message: "文档不存在",
			}, nil
		}
		l.Logger.Errorf("查询原始文档失败: %v, ID: %d", err, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "查询失败",
			Info:    err.Error(),
		}, nil
	}

	// 检查 IsAudit=1 的不能更改标签
	if doc.IsAudit == 1 {
		return &types.Response{
			Code:    response.ConflictCode,
			Message: "已审核的文档不能更改标签",
		}, nil
	}

	// 更新 tag 字段
	doc.Tag = req.Tag
	doc.UpdatedAt = time.Now()
	doc.IsToAi = 0

	err = l.svcCtx.RawDocumentsModel.Update(l.ctx, doc)
	if err != nil {
		l.Logger.Errorf("更新文档标签失败: %v, ID: %d", err, req.Id)
		return &types.Response{
			Code:    response.ServerErrorCode,
			Message: "更新标签失败",
			Info:    err.Error(),
		}, nil
	}

	// 更新 meta.yaml 文件内容
	// 确定正确的 meta.yaml 文件路径
	var tagFilePath string

	tagFilePath = doc.FilePath + ".meta.yaml"

	info := ""

	// 创建 YAML 结构
	metaData := map[string]interface{}{
		"tag": req.Tag,
	}

	// 将数据转换为 YAML 格式
	yamlData, err := yaml.Marshal(metaData)
	if err != nil {
		info = fmt.Sprintf("转换YAML格式失败: %v, 文件路径: %s", err, tagFilePath)
		l.Logger.Errorf(info)
		// 不中断流程，只记录错误
	} else {
		// 写入 YAML 文件
		err = os.WriteFile(tagFilePath, yamlData, 0644)
		if err != nil {
			info = fmt.Sprintf("更新meta.yaml文件失败: %v, 文件路径: %s", err, tagFilePath)
			l.Logger.Errorf(info)
			// 不中断流程，只记录错误
		} else {
			info = fmt.Sprintf("成功更新meta.yaml文件: %s, tag: %s", tagFilePath, req.Tag)
			l.Infof(info)
		}
	}

	// Touch 原始文档文件（更新修改时间）
	if err := os.Chtimes(doc.FilePath, time.Now(), time.Now()); err != nil {
		touchInfo := fmt.Sprintf("touch文档文件失败: %v, 文件路径: %s", err, doc.FilePath)
		l.Logger.Errorf(touchInfo)
		if info != "" {
			info += "; " + touchInfo
		} else {
			info = touchInfo
		}
	} else {
		l.Infof("成功touch文档文件: %s", doc.FilePath)
	}

	// 设置 Redis 标志，表示有新PDF
	if l.svcCtx.RedisClient != nil {
		if err := l.svcCtx.RedisClient.Set("have_new_pdf", "1"); err != nil {
			l.Logger.Errorf("设置Redis have_new_pdf标志失败: %v", err)
			// 不中断流程，只记录错误
		} else {
			l.Logger.Infof("成功设置Redis have_new_pdf=1")
		}
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "更新成功",
		Info:    info,
	}, nil
}
