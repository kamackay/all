package files

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type FileCache = map[string]os.FileInfo

func GetSize(path string, file fs.FileInfo) int64 {
	filename := filepath.Join(path, file.Name())
	if file.IsDir() {
		return GetFolderSize(filename, make(FileCache))
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

func GetFolderSize(pathName string, cache FileCache) int64 {
	var size int64
	if val, ok := cache[pathName]; ok && val != nil {
		return val.Size()
	}
	err := filepath.WalkDir(pathName, func(_ string, d os.DirEntry, err error) error {
		cache[path.Join(pathName, d.Name())], _ = d.Info()
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		size += getSize(d)
		return nil
	})
	if err != nil {
		return 0
	}
	return size
}

func getSize(entry os.DirEntry) int64 {
	info, err := entry.Info()
	if err != nil {
		return 0
	}
	return info.Size()
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
