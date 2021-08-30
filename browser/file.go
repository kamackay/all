package browser

import (
	"fmt"
	"gitlab.com/kamackay/all/files"
	"gitlab.com/kamackay/all/l"
	"os"
	"path/filepath"
)

type File struct {
	Path string
	Size uint64
}

func makeRelativeFile(path string, relative string) File {
	parentPath := filepath.Join(path, relative)
	parent, err := os.Stat(parentPath)
	var size uint64
	if err == nil {
		size = files.GetSize(parentPath, parent)
	} else {
		l.Print(fmt.Sprintf("Error Getting Size: %+v\n", err))
	}
	return File{
		Path: parentPath,
		Size: size,
	}
}
