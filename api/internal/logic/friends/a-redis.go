package friends

import (
	"context"

	"knowsource/api/internal/svc"
)

func isUserInBlacklist(ctx context.Context, svcCtx *svc.ServiceContext, uid uint64, targetUid uint64) (bool, error) {

	userBlackList, err := svcCtx.UserBlackListModel.FindByBlackIdWithCache(ctx, svcCtx.RedisClient, uid)
	if err != nil {
		return false, err
	}
	for _, userBlack := range userBlackList {
		if userBlack.BlackId == targetUid {
			return true, nil
		}
	}
	return false, nil
}
