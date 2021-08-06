package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	Tab = "  "
	Gig = 1000000000
)

type Opts struct {
	Verbose   bool   `short:"v" help:"Verbose"`
	Directory string `arg:"" optional:"" help:"Directory" default:"."`
	Humanize  bool   `short:"z" help:"Humanize File Sizes"`
	Large     bool   `short:"G" help:"Only print files over 1 GB"`
}

func getFiles(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return make([]fs.FileInfo, 0)
	}
	return files
}

func getFolderSize(path string) uint64 {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return err
	})
	if err != nil {
		return 0
	}
	return size
}

func getFileSize(file string) uint64 {
	fi, err := os.Stat(file)
	if err != nil {
		return 0
	}
	// get the size
	size := fi.Size()
	return uint64(size)
}

func indentation(index int) string {
	str := ""
	for i := 1; i < index; i++ {
		str += Tab
	}
	return str + "| "
}

func formatSize(file string, sizeBytes uint64, human bool) string {
	if human {
		return humanize.Bytes(sizeBytes)
	} else {
		return fmt.Sprintf("%d", sizeBytes)
	}
}

func printPath(file string, index int, isDir bool, opts Opts) {
	var size uint64
	if isDir {
		size = getFolderSize(file)
	} else {
		size = getFileSize(file)
	}
	if opts.Large && size < Gig {
		// File is less than a gig, quit
		return
	}
	fmt.Printf("%s%s - %s\n", indentation(index), formatSize(file, size, opts.Humanize), file)
}

func printFolder(dir string, index int, opts Opts) {
	files := getFiles(dir)
	for _, file := range files {
		if file.IsDir() {
			printPath(path.Join(dir, file.Name()), index, true, opts)
			printFolder(path.Join(dir, file.Name()), index+1, opts)
		} else {
			printPath(path.Join(dir, file.Name()), index, false, opts)
		}
	}
}

func main() {
	var opts Opts
	ctx := kong.Parse(&opts)

	start := time.Now()

	dir := opts.Directory

	if dir == "" {
		directory, err := os.Getwd()
		if err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
		dir = directory
	}

	base, err := filepath.Abs(dir)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return
	}

	printFolder(base, 0, opts)

	if time.Now().Sub(start) > 100*time.Millisecond {
		fmt.Printf("Done in %s\n", humanize.RelTime(start, time.Now(), "", ""))
	}
	ctx.Exit(0)
}
