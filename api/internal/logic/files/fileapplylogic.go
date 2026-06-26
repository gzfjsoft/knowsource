package files

import (
	"context"
	"encoding/json"
	"path/filepath"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type FileApplyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileApplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileApplyLogic {
	return &FileApplyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileApplyLogic) FileApply(req *types.FileApplyRequest) (resp *types.ApplyResponse, err error) {

	uid, _ := l.ctx.Value("uid").(json.Number).Int64()

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(uid))
	if err != nil {
		return nil, err
	}

	md5, err := fastPartialMD5(l.svcCtx.Config.FilesRoot + req.File)
	if err != nil {
		return nil, err
	}

	//get file name from req.File
	fileName := filepath.Base(req.File)

	_, err = l.svcCtx.SongPm5Model.Insert(l.ctx, &model.SongPm5{
		Pmd5:     md5,
		Songname: fileName,
	})
	if err != nil {
		logx.Infof("insert songpm5 failed, err: %v", err)
	}

	l.svcCtx.PlayLogModel.Insert(l.ctx, &model.PlayLog{
		UserId:   uint64(uid),
		Username: user.Username,
		Songname: fileName,
		Pmd5:     md5,
	})

	//random code
	code := uuid.New().String()
	l.svcCtx.RedisClient.Setex("FA-"+code, req.File, 10)
	return &types.ApplyResponse{
		Code: code,
	}, nil

}
