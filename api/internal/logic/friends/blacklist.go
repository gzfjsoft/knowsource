package friends

import (
	"context"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetBlacklistMap 获取用户的黑名单映射
func GetBlacklistMap(ctx context.Context, svcCtx *svc.ServiceContext, uid uint64) (map[uint64]bool, error) {
	// 获取黑名单列表
	blacklist, err := svcCtx.UserBlackListModel.FindByUserIdWithCache(ctx, svcCtx.RedisClient, uid)
	if err != nil {
		logx.Error("GetBlacklistMap error", logx.Field("error", err))
		return make(map[uint64]bool), err
	}

	blacklist2, err := svcCtx.UserBlackListModel.FindByBlackIdWithCache(ctx, svcCtx.RedisClient, uid)
	if err != nil {
		logx.Error("GetBlacklistMap error", logx.Field("error", err))
		return make(map[uint64]bool), err
	}

	// 构建黑名单用户ID集合
	blacklistMap := make(map[uint64]bool)
	for _, black := range blacklist {
		blacklistMap[black.BlackId] = true
	}
	for _, black := range blacklist2 {
		blacklistMap[black.UserId] = true
	}

	return blacklistMap, nil
}

// IsInBlacklist 检查用户是否在黑名单中
func IsInBlacklist(blacklistMap map[uint64]bool, userId uint64) bool {
	return blacklistMap[userId]
}

// GetBlacklistErrorResponse 获取黑名单错误响应
func GetBlacklistErrorResponse() types.Response {
	return types.Response{
		Code:    response.ServerErrorCode,
		Message: "获取黑名单列表失败",
	}
}

// GetBlacklistMap 获取用户的黑名单映射
func GetUserAcquireUserMap(ctx context.Context, svcCtx *svc.ServiceContext, uid uint64) (map[uint64]uint64, error) {
	userAcquireUser, err := svcCtx.UserAcquireUserModel.FindAllByUserId(ctx, uint64(uid))
	if err != nil {
		return nil, err
	}

	var payInfoMap map[uint64]uint64 = make(map[uint64]uint64)
	for _, userAcquireUser := range userAcquireUser {
		payInfoMap[userAcquireUser.AcqUserId] = userAcquireUser.UserId
	}
	return payInfoMap, nil
}
