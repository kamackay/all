package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"io/fs"
	"io/ioutil"
	"os"
	"time"
)

const (
	Tab = "\t"
)

var Opts struct {
	Verbose bool `name:"v" help:"Verbose"`
	Directory string `name:"d" help:"Dir" default:""`
}

func getFiles(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Printf("%+v\n", err)
		return make([]fs.FileInfo, 0)
	}
	return files
}

func indentation(index int) string {
 	str := ""
	for i := 1; i < index; i++ {
		str += Tab
	}
	return str
}

func printFolder(path string, index int) {
	fmt.Printf("%s%s\n", indentation(index), path)
	files := getFiles(path)
	for _, file := range files {
		if file.IsDir() {
			fmt.Printf(file.Name())
		} else {
			fmt.Printf("%s%s\n", indentation(index), file.Name())
		}
	}
}

func main() {
	ctx := kong.Parse(&Opts)

	start := time.Now()

	dir := Opts.Directory

	if dir == "" {
		path, err := os.Getwd()
		if err != nil {
			fmt.Printf("%+v\n", path)
			return
		}
		dir = path
	}

	printFolder(dir, 0)

	fmt.Printf("Done in %s\n", humanize.RelTime(start, time.Now(), "", ""))
	ctx.Exit(0)
}
