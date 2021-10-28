package files

import (
	"io/fs"
	"path"
	"strings"
)

type File struct {
	Name  string
	IsDir bool
}

func NewFile(info fs.FileInfo, parentFolder string) File {
	return File{
		Name:  path.Join(parentFolder, info.Name()),
		IsDir: info.IsDir(),
	}
}

type ByName []File

func (a ByName) Len() int           { return len(a) }
func (a ByName) Less(i, j int) bool { return strings.Compare(a[i].Name, a[j].Name) > 0 }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
