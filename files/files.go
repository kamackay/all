package files

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetSize(path string, file fs.FileInfo) uint64 {
	filename := filepath.Join(path, file.Name())
	if file.IsDir() {
		return GetFolderSize(filename)
	} else {
		return GetFileSize(filename)
	}
}

func GetFileSize(file string) uint64 {
	fi, err := os.Stat(file)
	if err != nil {
		return 0
	}
	// get the size
	size := fi.Size()
	return uint64(size)
}

func GetFiles(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return make([]fs.FileInfo, 0)
	}
	return files
}

func GetFolderSize(path string) uint64 {
	var size uint64
	err := filepath.WalkDir(path, func(_ string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !entry.IsDir() {
			if info, err := entry.Info(); err != nil {
				size += uint64(info.Size())
			}
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return size
}

func ReadStart(path string, size int) (string, error) {
	r, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer r.Close()
	header := make([]byte, size)
	_, err = r.Read(header)
	if err != nil {
		return "", err
	}
	return string(header), nil
}
