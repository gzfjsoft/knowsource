package logic

import (
	"context"
	"fmt"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/api/internal/utils"
	"knowsource/common/response"
	"knowsource/consts"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminServerListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminServerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminServerListLogic {
	return &AdminServerListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminServerListLogic) AdminServerList(req *types.ServerListRequest) (resp *types.ServerListResponse, err error) {

	newresp := func(code int64, message, info string) *types.ServerListResponse {
		return &types.ServerListResponse{
			Response: types.Response{
				Code:    code,
				Message: message,
				Info:    info,
			},
		}
	}

	sysrole, _ := l.ctx.Value("role").(string)
	if sysrole != consts.SUPER_ADMIN && sysrole != consts.ONLY_ADMIN {
		return newresp(response.UnauthorizedCode, "没有权限", sysrole), nil
	}

	condition := fmt.Sprint("where 1=1")

	if req.RegionId > 0 {
		condition += fmt.Sprintf(" and region_id = %d", req.RegionId)
	}
	if req.TagId > 0 {
		condition += fmt.Sprintf(" and server_id in (select server_id from server_tags where tag_id= %d )", req.TagId)
	}
	if req.IsPayMin > 0 {
		condition += fmt.Sprintf(" and is_pay_min = %d", req.IsPayMin)
	}
	if req.IsPayDaily > 0 {
		condition += fmt.Sprintf(" and is_pay_daily = %d", req.IsPayDaily)
	}
	if req.IsPayMonthly > 0 {
		condition += fmt.Sprintf(" and is_pay_monthly = %d", req.IsPayMonthly)
	}
	if req.IsPayYearly > 0 {
		condition += fmt.Sprintf(" and is_pay_yearly = %d", req.IsPayYearly)
	}

	if req.GpuCount > 0 {
		condition += fmt.Sprintf(" and gpu_count = %d", req.GpuCount)
	}
	if len(req.GpuModel) > 0 {
		condition += " and gpu_model in ("
		for i, v := range req.GpuModel {
			condition += fmt.Sprintf("'%s'", v)
			if i < len(req.GpuModel)-1 {
				condition += ","
			}
		}
		condition += ")"
	}

	if req.GpuMem > 0 {
		condition += fmt.Sprintf(" and gpu_mem = %d", req.GpuMem)
	}
	if len(req.CpuModel) > 0 {
		condition += " and cpu_model in ("
		for i, v := range req.CpuModel {
			condition += fmt.Sprintf("'%s'", v)
			if i < len(req.CpuModel)-1 {
				condition += ","
			}
		}
		condition += ")"
	}
	if req.CpuCores > 0 {
		condition += fmt.Sprintf(" and cpu_cores = %d", req.CpuCores)
	}
	if req.Supplier > 0 {
		condition += fmt.Sprintf(" and supplier = %d", req.Supplier)
	}

	total, err := l.svcCtx.ServerModel.Count(l.ctx, condition)
	if err != nil {
		return newresp(response.ServerErrorCode, "获取服务器列表计数失败", err.Error()), nil
	}

	condition += " order by gpu_count-gpu_used desc "

	if req.Page > 0 && req.PageSize > 0 {
		condition += fmt.Sprintf(" limit %d,%d", (req.Page-1)*req.PageSize, req.PageSize)
	}

	data, err := l.svcCtx.ServerModel.FindList(l.ctx, condition)
	if err != nil {
		return newresp(response.ServerErrorCode, "获取服务器列表失败", err.Error()), nil
	}

	var serviceList []types.Server

	server := types.Server{}
	for _, v := range *data {
		utils.CopyStruct(&v, &server, true)
		utils.CopyStruct(&v, &server.ServerBaseField, true)
		server.ServerBaseField.ExpireDate = uint64(v.ExpireDate.Unix())
		serviceList = append(serviceList, server)

	}

	return &types.ServerListResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
			Info:    "",
		},
		Data: &types.ServerListDataResp{Servers: serviceList, Total: total},
	}, nil

}
