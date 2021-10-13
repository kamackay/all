package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"gitlab.com/kamackay/all/browser"
	"gitlab.com/kamackay/all/files"
	"gitlab.com/kamackay/all/l"
	"gitlab.com/kamackay/all/utils"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	Gig = 1000000000
)

type Opts struct {
	Browser   bool   `short:"b" help:"Run Browser"`
	Verbose   bool   `short:"v" help:"Verbose"`
	Directory string `arg:"d" help:"Directory" default:"."`
	Humanize  bool   `short:"z" help:"Humanize File Sizes"`
	Large     bool   `short:"G" help:"Only print files over 1 GB"`
	FirstOnly bool   `short:"f" help:"Only show the first level of the filetree"`
}

func printPath(file string, index int, isDir bool, opts Opts) {
	var size int64
	if isDir {
		size = files.GetFolderSize(file)
	} else {
		size = files.GetFileSize(file)
	}
	if opts.Large && size < Gig {
		// File is less than a gig, quit
		return
	}
	fmt.Printf("%s%s - %s\n", utils.Indentation(index), utils.FormatSize(uint64(size), opts.Humanize), file)
}

func printFolder(dir string, index int, opts Opts) {
	fs := files.GetFiles(dir)
	for _, file := range fs {
		if file.IsDir() {
			printPath(path.Join(dir, file.Name()), index, true, opts)
			if !opts.FirstOnly {
				printFolder(path.Join(dir, file.Name()), index+1, opts)
			}
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

	if opts.Browser {
		// Run Browser
		l.Print("Running Browser!")
		b, err := browser.New(base)
		if err != nil {
			fmt.Printf("%+v\n", err)
			return
		}
		b.Run()
		return
	}

	printFolder(base, 0, opts)

	if time.Now().Sub(start) > 100*time.Millisecond {
		fmt.Printf("Done in %s\n", humanize.RelTime(start, time.Now(), "", ""))
	}
	ctx.Exit(0)
}
