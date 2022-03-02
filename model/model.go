package model

import "os"

type SortType int

const (
	SortSize SortType = iota
	SortName
)

type FileBean struct {
	info  os.FileInfo
	Count uint
}

func (bean FileBean) Size() int64 {
	return bean.info.Size()
}

func MakeFileBean(info os.FileInfo, count uint) *FileBean {
	return &FileBean{Count: count, info: info}
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
