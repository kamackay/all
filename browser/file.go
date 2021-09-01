package browser

import (
	"fmt"
	"gitlab.com/kamackay/all/files"
	"gitlab.com/kamackay/all/l"
	"os"
	"path/filepath"
)

type File struct {
	Path         string
	Size         int64
	LastModified string
}

func makeRelativeFile(path string, relative string) File {
	relativePath := filepath.Join(path, relative)
	info, err := os.Stat(relativePath)
	var size int64
	if err == nil {
		size = files.GetSize(relativePath, info)
	} else {
		l.Print(fmt.Sprintf("Error Getting Size: %+v\n", err))
	}
	return File{
		Path: relativePath,
		Size: size,
		LastModified: files.PrintTime(info),
	}
}