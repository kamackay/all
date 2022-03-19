package files

import (
	"fmt"
	"github.com/kamackay/all/model"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type FileCache = map[string]*model.FileBean

func GetSize(path string, file fs.FileInfo) int64 {
	filename := filepath.Join(path, file.Name())
	if file.IsDir() {
		size, _ := GetFolderInfo(filename, make(FileCache))
		return int64(size)
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

func GetFilesFirstLevel(dir string, cache FileCache) []*model.FileBean {
	return WalkFiles(dir, cache, true)
}

func WalkFiles(dir string, cache FileCache, topOnly bool) []*model.FileBean {
	list := make([]*model.FileBean, 0)
	_ = filepath.WalkDir(dir, func(subPath string, d os.DirEntry, err error) error {
		if topOnly && path.Dir(subPath) != dir {
			return nil
		}
		if val, ok := cache[subPath]; ok && val != nil {
			list = append(list, val)
			return nil
		}
		if err != nil {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		list = append(list, convertInfoToBean(subPath, info, cache))
		return nil
	})
	return list
}

func GetFilesRecursive(dir string) []*model.FileBean {
	fileInfos, err := ioutil.ReadDir(dir)
	beans := make([]*model.FileBean, 0)
	if err != nil {
		return beans
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		return strings.Compare(fileInfos[i].Name(), fileInfos[j].Name()) > 0
	})
	for _, f := range fileInfos {
		filePath := path.Join(dir, f.Name())
		if f.IsDir() {
			subFiles := GetFilesRecursive(filePath)
			var size uint64
			for _, subFile := range subFiles {
				size += subFile.Size
			}
			beans = append(beans, model.MakeFileBean(path.Join(dir, f.Name()), f, uint(len(subFiles)), size))
			for _, subFile := range subFiles {
				beans = append(beans, subFile)
			}
		} else {
			beans = append(beans, model.MakeFileBean(path.Join(dir, f.Name()), f, 1, uint64(f.Size())))
		}
	}
	return beans
}

func convertInfoToBean(filePath string, f fs.FileInfo, cache FileCache) *model.FileBean {
	if val, ok := cache[filePath]; ok && val != nil {
		return val
	}
	if f.IsDir() {
		size, count := GetFolderInfo(filePath, cache)
		return model.MakeFileBean(filePath, f, count, size)
	} else {
		bean := model.MakeFileBean(filePath, f, 0, uint64(f.Size()))
		cache[filePath] = bean
		return bean
	}
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

func GetFolderInfo(pathName string, cache FileCache) (uint64, uint) {
	var size uint64
	var count uint
	if val, ok := cache[pathName]; ok && val != nil {
		return val.Size, val.Count
	}
	err := filepath.WalkDir(pathName, func(fullPath string, d os.DirEntry, err error) error {
		if val, ok := cache[pathName]; ok && val != nil {
			count++
			size += val.Size
			return nil
		}
		info, _ := d.Info()
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		} else {
			count++
		}
		currentSize := uint64(getSize(d))
		size += currentSize
		cache[fullPath] = model.MakeFileBean(fullPath, info, 0, currentSize)
		return nil
	})
	if err != nil {
		return 0, 0
	}
	return size, count
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
