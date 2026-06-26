package mysql

import (
	"os"
	"path/filepath"
)

// ResolveBackupOutputDir 备份 zip 所在目录：与 api 目录同级。
// 当进程工作目录在 api 下时，使用其父目录；否则使用当前工作目录（便于在仓库根目录启动时仍把文件放在仓库根）。
func ResolveBackupOutputDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	wd = filepath.Clean(wd)
	base := wd
	if filepath.Base(base) == "api" {
		base = filepath.Dir(base)
	}
	return filepath.Abs(base)
}
