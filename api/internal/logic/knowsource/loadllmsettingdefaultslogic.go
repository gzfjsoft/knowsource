package knowsource

import (
	"context"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoadLLMSettingDefaultsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoadLLMSettingDefaultsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadLLMSettingDefaultsLogic {
	return &LoadLLMSettingDefaultsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// LoadLLMSettingDefaults 返回主配置文件中的系统默认项，供前端重置表单（不写入 ai_{clientId}.yaml）
func (l *LoadLLMSettingDefaultsLogic) LoadLLMSettingDefaults() (resp *types.LLMSettingResponse, err error) {
	data := BuildSystemLLMSettingDefaults(&l.svcCtx.Config)
	return &types.LLMSettingResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: data,
	}, nil
}
