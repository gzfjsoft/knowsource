package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"knowsource/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

type UploadIdcardBackLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	r      *http.Request
}

func NewUploadIdcardBackLogic(ctx context.Context, r *http.Request, svcCtx *svc.ServiceContext) *UploadIdcardBackLogic {
	return &UploadIdcardBackLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		r:      r,
		svcCtx: svcCtx,
	}
}

func (l *UploadIdcardBackLogic) UploadIdcardBack() (resp *types.UploadResponse, err error) {

	//上传文件到本地 uploadfile

	uid, _ := l.ctx.Value("uid").(json.Number).Int64()
	email, _ := l.ctx.Value("email").(string)

	logx.Info("JWT uid=", uid, " Name=", email)

	l.r.ParseMultipartForm(maxFileSize)
	file, handler, err := l.r.FormFile("file")
	if err != nil {
		// logx.Errorf("upload file: %+v, file size: %d, MIME header: %+v err: %+v",
		// 	handler.Filename, handler.Size, handler.Header, err)

		return &types.UploadResponse{
			Response: types.Response{
				Message: "上传失败",
				Code:    response.ServerErrorCode,
				Info:    err.Error(),
			},
		}, nil
	}
	defer file.Close()

	logx.Infof("upload file: %+v, file size: %d, MIME header: %+v",
		handler.Filename, handler.Size, handler.Header)
	// separator := string(filepath.Separator)

	path_str := path.Join(l.svcCtx.Config.UploadPath, strconv.Itoa(int(uid)), "idcard")

	//漏洞修复，文件类型限制
	fileExt := filepath.Ext(handler.Filename)

	fileExt = strings.ToLower(fileExt)

	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" && fileExt != ".webp" {
		return &types.UploadResponse{
			Response: types.Response{
				Message: "不支持的文件类型,只支持png,jpg,jpeg,webp",
				Code:    response.InvalidRequestParamCodeInHandler,
				Info:    fileExt,
			},
		}, nil
	}

	// Get file extension and create new filename with timestamp
	uuid := uuid.New().String()
	datestr := time.Now().Format("200601021504")
	newFilename := fmt.Sprintf("%d_%s_b_%s%s", uid, datestr, uuid[:8], fileExt)

	filename := path.Join(path_str, newFilename)
	// filename := path + string(filepath.Separator) + handler.Filename
	err = os.MkdirAll(path_str, 0755)
	if err != nil {
		return &types.UploadResponse{
			Response: types.Response{
				Message: "创建目录失败",
				Code:    response.ServerErrorCode,
				Info:    path_str + "," + err.Error(),
			},
		}, nil
	}
	tempFile, err := os.Create(filename)
	if err != nil {
		return &types.UploadResponse{
			Response: types.Response{
				Message: "create file failed",
				Code:    response.ServerErrorCode,
				Info:    filename + "," + err.Error(),
			},
		}, nil
	}
	defer tempFile.Close()
	io.Copy(tempFile, file)
	logx.Infof("upload file success %+v", file)

	// 插入数据库 useruploadfile
	_, err = l.svcCtx.UserUploadPhotoModel.Insert(l.ctx, &model.UserUploadPhoto{
		UserId:    uint64(uid),
		PhotoUrl:  filename,
		IsAudited: 0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		logx.Errorf("insert user upload photo failed %+v", err)

	}

	return &types.UploadResponse{
		Response: types.Response{
			Message: filename,
			Code:    response.SuccessCode,
		},
		Filename: filename,
	}, nil
}
