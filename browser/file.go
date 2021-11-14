package browser

import (
	"fmt"
	"github.com/kamackay/all/files"
	"github.com/kamackay/all/l"
	"github.com/kamackay/all/utils"
	"os"
	"path/filepath"
)

type File struct {
	Path         string
	Size         int64
	Dir          bool
	LastModified string
	ToString     func() string
	Children     uint
}

func ToString(file File) string {
	if file.Dir {
		return fmt.Sprintf("%s    %s -> %s (#%d)",
			file.LastModified,
			file.Path,
			utils.FormatSize(uint64(file.Size), true),
			file.Children)
	}
	return fmt.Sprintf("%s    %s -> %s",
		file.LastModified,
		file.Path,
		utils.FormatSize(uint64(file.Size), true))
}

func makeRelativeFile(path string, relative string) File {
	relativePath := filepath.Join(path, relative)
	info, err := os.Stat(relativePath)
	if err != nil {
		l.Print(fmt.Sprintf("Error Getting Size: %+v\n", err))
	}
	return File{
		Path:         relativePath,
		Size:         0,
		LastModified: files.PrintTime(info),
		Dir:          info.IsDir(),
		Children:     files.CountChildren(relativePath),
	}
}
