package files

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

func GetSize(path string, file fs.FileInfo) int64 {
	filename := filepath.Join(path, file.Name())
	if file.IsDir() {
		return GetFolderSize(filename)
	} else {
		return GetFileSize(filename)
	}
}

func GetFileSize(file string) int64 {
	fi, err := os.Stat(file)
	if err != nil {
		return 0
	}
	// get the size
	size := fi.Size()
	return size
}

func GetFiles(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return make([]fs.FileInfo, 0)
	}
	return files
}

func GetFolderSize(path string) int64 {
	var size int64
	err := filepath.WalkDir(path, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		size += info.Size()
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

func PrintTime(info os.FileInfo) string {
	return info.ModTime().Format("2006-01-02 15:04:05")
}
