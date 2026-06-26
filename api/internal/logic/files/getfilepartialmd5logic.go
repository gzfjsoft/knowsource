package files

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"knowsource/api/internal/svc"
	"knowsource/api/internal/types"
	"knowsource/common/response"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFilePartialMD5Logic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFilePartialMD5Logic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFilePartialMD5Logic {
	return &GetFilePartialMD5Logic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFilePartialMD5Logic) GetFilePartialMD5(req *types.GetFilePartialMD5Request) (resp *types.GetFilePartialMD5Response, err error) {
	// todo: add your logic here and delete this line

	md5Str, err := fastPartialMD5(l.svcCtx.Config.FilesRoot + req.Path)
	if err != nil {
		return nil, err
	}

	return &types.GetFilePartialMD5Response{
		Response: types.Response{
			Code:    response.SuccessCode,
			Message: "success",
		},
		MD5: md5Str,
	}, nil
}

func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func fastPartialMD5(filePath string) (string, error) {

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("打开文件出错: %v\n", err)
		return "", err
	}
	defer file.Close()

	// 获取文件大小
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("获取文件信息出错: %v\n", err)
		return "", err
	}
	fileSize := fileInfo.Size()

	if fileSize <= 3*1024 {
		// 文件小于 3KB，读取全量数据并计算 MD5
		buffer := make([]byte, fileSize)
		_, err := file.Read(buffer)
		if err != nil {
			fmt.Printf("读取文件出错: %v\n", err)
			return "", err
		}
		md5Str := calculateMD5(buffer)
		fmt.Printf("文件小于 3KB，全量数据的 MD5 值: %s\n", md5Str)
		return md5Str, nil
	} else {
		// 文件大于等于 3KB，读取开头、中间、结尾各 1KB 并计算 MD5
		// 读取文件开头 1KB
		startBuffer := make([]byte, 1024)
		n, err := file.Read(startBuffer)
		if err != nil {
			fmt.Printf("读取文件开头出错: %v\n", err)
			return "", err
		}
		startBuffer = startBuffer[:n]

		// 计算中间位置
		middleOffset := fileSize/2 - 512
		if middleOffset < 0 {
			middleOffset = 0
		}

		// 移动文件指针到中间位置
		_, err = file.Seek(middleOffset, os.SEEK_SET)
		if err != nil {
			fmt.Printf("移动文件指针到中间位置出错: %v\n", err)
			return "", err
		}

		// 读取文件中间 1KB
		middleBuffer := make([]byte, 1024)
		n, err = file.Read(middleBuffer)
		if err != nil {
			fmt.Printf("读取文件中间出错: %v\n", err)
			return "", err
		}
		middleBuffer = middleBuffer[:n]

		// 移动文件指针到文件末尾前 1KB 位置
		endOffset := fileSize - 1024
		if endOffset < 0 {
			endOffset = 0
		}
		_, err = file.Seek(endOffset, os.SEEK_SET)
		if err != nil {
			fmt.Printf("移动文件指针到末尾前 1KB 位置出错: %v\n", err)
			return "", err
		}

		// 读取文件末尾 1KB
		endBuffer := make([]byte, 1024)
		n, err = file.Read(endBuffer)
		if err != nil {
			fmt.Printf("读取文件末尾出错: %v\n", err)
			return "", err
		}
		endBuffer = endBuffer[:n]

		// 拼接三段数据
		combinedData := append(startBuffer, middleBuffer...)
		combinedData = append(combinedData, endBuffer...)

		strSize := strconv.Itoa(int(fileSize))
		combinedData = append(combinedData, []byte(strSize)...)

		// 计算 MD5 值
		md5Str := calculateMD5(combinedData)
		fmt.Printf("文件大于等于 3KB，拼接后 3KB 数据的 MD5 值: %s\n", md5Str)
		return md5Str, nil
	}

}
