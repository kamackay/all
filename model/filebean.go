package model

import (
	"os"
	"syscall"
	"time"
)

type FileBean struct {
	info  os.FileInfo
	Count uint
	Name  string
	Size  uint64
	stat  *syscall.Stat_t
}

func (bean FileBean) IsDir() bool {
	return bean.info.IsDir()
}

func (bean FileBean) LastModified() time.Time {
	return bean.info.ModTime()
}

func (bean FileBean) Created() time.Time {
	return time.Unix(bean.stat.Ctimespec.Sec, bean.stat.Ctimespec.Nsec)
}

func MakeFileBean(name string, info os.FileInfo, count uint, size uint64) *FileBean {
	stat := info.Sys().(*syscall.Stat_t)
	return &FileBean{
		Name:  name,
		Count: count,
		info:  info,
		Size:  size,
		stat:  stat,
	}
}
