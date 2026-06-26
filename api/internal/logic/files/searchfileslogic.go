package files

import (
	"context"
	"io/fs"
	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"
	"path/filepath"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchFilesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchFilesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchFilesLogic {
	return &SearchFilesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchFilesLogic) SearchFiles(req *types.SearchFilesRequest) (resp *types.SearchFilesResponse, err error) {

	files := make([]types.FileItem, 0)

	req.Path = l.svcCtx.Config.FilesRoot + req.Path
	req.Query = "*" + req.Query + "*"

	err = filepath.Walk(req.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if info.IsDir() {
			return nil
		}

		// 使用通配符模式匹配文件名
		match, err := filepath.Match(req.Query, info.Name())
		if err != nil {
			return err
		}

		if match {
			files = append(files, types.FileItem{
				Path:        strings.TrimPrefix(path, l.svcCtx.Config.FilesRoot),
				Name:        info.Name(),
				Size:        info.Size(),
				IsDirectory: info.IsDir(),
				CreatedAt:   uint64(info.ModTime().Unix()),
				UpdatedAt:   uint64(info.ModTime().Unix()),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	resp = &types.SearchFilesResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.SearchFilesResponseData{
			Files: files,
			Total: uint64(len(files)),
		},
	}

	return resp, nil
}
