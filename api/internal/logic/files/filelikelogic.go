package files

import (
	"context"
	"database/sql"
	"encoding/json"
	"path"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type FileLikeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFileLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FileLikeLogic {
	return &FileLikeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FileLikeLogic) FileLike(req *types.FileLikeRequest) (resp *types.Response, err error) {

	// todo: add your logic here and delete this line

	uid, _ := l.ctx.Value("uid").(json.Number).Int64()

	user, err := l.svcCtx.UsersModel.FindOne(l.ctx, uint64(uid))
	if err != nil {
		return nil, err
	}

	md5, err := fastPartialMD5(l.svcCtx.Config.FilesRoot + req.File)
	if err != nil {
		return nil, err
	}

	//find if exists
	exists, err := l.svcCtx.FileLikeModel.FindOneByUserIdPmd5(l.ctx, uint64(uid), md5)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	filename := path.Base(req.File)
	if exists != nil {
		//update
		exists.Degree = req.Degree
		exists.Filename = filename
		exists.Pmd5 = md5
		exists.Username = user.Username
		exists.UserId = uint64(uid)
		err = l.svcCtx.FileLikeModel.Update(l.ctx, exists)
		if err != nil {
			return nil, err
		}
	} else {
		_, err = l.svcCtx.FileLikeModel.Insert(l.ctx, &model.FileLike{
			UserId:   uint64(uid),
			Username: user.Username,
			Pmd5:     md5,
			Filename: filename,
			Degree:   req.Degree,
		})
		if err != nil {
			return nil, err
		}
	}

	return &types.Response{
		Code:    response.SuccessCode,
		Message: "success",
	}, nil
}
