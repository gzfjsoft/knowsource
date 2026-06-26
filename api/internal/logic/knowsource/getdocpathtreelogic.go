package knowsource

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

type GetDocPathTreeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取文档路径文件树
func NewGetDocPathTreeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDocPathTreeLogic {
	return &GetDocPathTreeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetDocPathTreeLogic) GetDocPathTree(req *types.KnowsourceDocPathTreeRequest) (resp *types.KnowsourceDocPathTreeResponse, err error) {
	// 获取文档路径根目录
	docRoot := l.svcCtx.Config.Knowdata.DocumentPath
	if docRoot == "" {
		return &types.KnowsourceDocPathTreeResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "文档路径未配置",
			},
		}, nil
	}

	// 如果请求路径为空，使用根路径
	targetPath := docRoot
	if req.Path != "" {
		// 确保路径安全，防止路径遍历攻击
		reqPath := strings.TrimPrefix(req.Path, "/")
		reqPath = strings.TrimPrefix(reqPath, "..")
		targetPath = filepath.Join(docRoot, reqPath)
		// 确保目标路径在根目录下
		if !strings.HasPrefix(targetPath, docRoot) {
			return &types.KnowsourceDocPathTreeResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "无效的路径",
				},
			}, nil
		}
	}

	// 检查路径是否存在
	info, err := os.Stat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.KnowsourceDocPathTreeResponse{
				Response: types.Response{
					Code:    response.ServerErrorCode,
					Message: "路径不存在",
					Info:    err.Error(),
				},
			}, nil
		}
		l.Logger.Errorf("检查路径失败: %v", err)
		return &types.KnowsourceDocPathTreeResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "检查路径失败",
				Info:    err.Error(),
			},
		}, nil
	}

	// 如果不是目录，返回错误
	if !info.IsDir() {
		return &types.KnowsourceDocPathTreeResponse{
			Response: types.Response{
				Code:    response.ServerErrorCode,
				Message: "路径不是目录",
			},
		}, nil
	}

	// 递归构建文件树
	tree := l.buildTree(targetPath, docRoot)

	return &types.KnowsourceDocPathTreeResponse{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		Data: &types.KnowsourceDocPathTreeData{
			Tree: tree,
		},
	}, nil
}

// buildTree 递归构建文件树
func (l *GetDocPathTreeLogic) buildTree(rootPath string, basePath string) []types.KnowsourceDocPathTreeNode {
	var nodes []types.KnowsourceDocPathTreeNode

	// 读取目录内容
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		l.Logger.Errorf("读取目录失败: %v, 路径: %s", err, rootPath)
		return nodes
	}

	// 先处理目录，再处理文件
	for _, entry := range entries {
		// 跳过隐藏文件（以 . 开头）
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(rootPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			l.Logger.Errorf("获取文件信息失败: %v, 路径: %s", err, fullPath)
			continue
		}

		// 计算相对路径
		relPath, err := filepath.Rel(basePath, fullPath)
		if err != nil {
			relPath = strings.TrimPrefix(fullPath, basePath)
			relPath = strings.TrimPrefix(relPath, "/")
		}

		node := types.KnowsourceDocPathTreeNode{
			Name:  entry.Name(),
			Path:  relPath,
			IsDir: entry.IsDir(),
			Size:  0,
		}

		if entry.IsDir() {
			// 递归获取子节点
			node.Children = l.buildTree(fullPath, basePath)
		} else {
			// 文件大小
			node.Size = info.Size()
			node.Children = []types.KnowsourceDocPathTreeNode{}
		}

		nodes = append(nodes, node)
	}

	return nodes
}
