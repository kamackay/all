package model

import (
	"os"
	"time"
)

type FileBean struct {
	info  os.FileInfo
	Count uint
	Name  string
	Size  uint64
}

func (bean FileBean) IsDir() bool {
	return bean.info.IsDir()
}

func (bean FileBean) LastModified() time.Time {
	return bean.info.ModTime()
}

func MakeFileBean(name string, info os.FileInfo, count uint, size uint64) *FileBean {
	return &FileBean{
		Name:  name,
		Count: count,
		info:  info,
		Size:  size,
	}
}
