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

func CountChildren(file string) uint {
	files, err := ioutil.ReadDir(file)
	if err != nil {
		return 0
	}
	return uint(len(files))
}

func CountChildrenRecursive(pathName string) uint {
	var count uint
	err := filepath.WalkDir(pathName, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		count++
		return nil
	})
	if err != nil {
		return 0
	}
	return count
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

func GetFiles(filename string) []fs.FileInfo {
	files, err := ioutil.ReadDir(filename)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return make([]fs.FileInfo, 0)
	}
	return files
}

func ScanFiles(dir string) <-chan File {
	files := make(chan File)
	go func() {
		ScanFilesWorker(dir, files)
		close(files)
	}()
	return files
}

func ScanFilesWorker(dir string, output chan<- File) {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}
	for _, info := range infos {
		file := NewFile(info, dir)
		if file.IsDir {
			ScanFilesWorker(file.Name, output)
		} else {
			output <- file
		}
	}
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

func ReadEntire(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
