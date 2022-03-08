package model

import (
	"os"
	"time"
)

type SortType int

const (
	SortSize SortType = iota
	SortName
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

type LoadingInfo struct {
	Item    int
	Total   int
	Current string
	Render  bool
}

type FileMode struct {
	Contents string
}

type Confirmation struct {
	Message string
	Action  func()
}

type SortFunction = func(i, j int) bool

func SortTypeName(sortType SortType) string {
	switch sortType {
	case SortName:
		return "name"
	case SortSize:
		return "filesize"
	default:
		return "idk, randomly" // Again, shouldn't be possible
	}
}
