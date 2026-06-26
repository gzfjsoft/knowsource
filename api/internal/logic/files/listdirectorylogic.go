package files

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"

	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListDirectoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListDirectoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListDirectoryLogic {
	return &ListDirectoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListDirectoryLogic) ListDirectory(req *types.ListDirectoryRequest) (resp *types.ListDirectoryResponse, err error) {
	// Initialize response

	//base on /mnt/music
	root := l.svcCtx.Config.FilesRoot + req.Path + "/*"
	fileItems := make([]types.FileItem, 0)
	req.IsAllFile = l.svcCtx.Config.IsAllFile
	if req.IsAllFile != 0 {
		//list all files
		root := l.svcCtx.Config.FilesRoot + req.Path
		err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip the root directory itself
			if path == root {
				return nil
			}

			if !info.IsDir() {
				//get the file ext
				ext := strings.ToLower(filepath.Ext(info.Name()))

				if (ext == ".mp3") || (ext == ".mp4") || (ext == ".m4a") || (ext == ".flac") || (ext == ".wav") || (ext == ".aac") || (ext == ".wma") {
					fileItems = append(fileItems, types.FileItem{
						Name:        info.Name(),
						Path:        strings.TrimPrefix(path, l.svcCtx.Config.FilesRoot),
						IsDirectory: info.IsDir(),
						Size:        info.Size(),
						CreatedAt:   uint64(info.ModTime().Unix()),
					})
				}
			}
			return nil
		})

		if err != nil {
			return &types.ListDirectoryResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "Failed to list directory recursively",
					Info:    err.Error(),
				}}, err
		}
	} else {
		// get all files in /mnt/music
		files, err := filepath.Glob(root)
		if err != nil {
			return &types.ListDirectoryResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "Failed to list directory",
					Info:    err.Error(),
				}}, err

		}

		// Extract just the base names from the full paths
		for _, file := range files {

			fileInfo, err := os.Stat(file)
			if err != nil {
				continue
			}

			baseName := filepath.Base(file)
			fileItems = append(fileItems, types.FileItem{
				Name:        baseName,
				Path:        strings.TrimPrefix(file, l.svcCtx.Config.FilesRoot),
				IsDirectory: fileInfo.IsDir(),
				Size:        fileInfo.Size(),
				CreatedAt:   uint64(fileInfo.ModTime().Unix()),
			})
		}
	}

	resp = &types.ListDirectoryResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.ListDirectoryResponseData{
			Files:       fileItems,
			Total:       uint64(len(fileItems)),
			CurrentPath: req.Path,
		},
	}

	return resp, nil
}
