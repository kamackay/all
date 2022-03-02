package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/kamackay/all/browser"
	"github.com/kamackay/all/files"
	"github.com/kamackay/all/l"
	"github.com/kamackay/all/utils"
	"github.com/kamackay/all/version"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	Gig = 1000000000
)

type Opts struct {
	Version   bool   `help:"Print Version"`
	Browser   bool   `short:"b" help:"Run Browser"`
	Verbose   bool   `short:"v" help:"Verbose"`
	Directory string `arg:"d" help:"Directory" default:"."`
	Sort      string `short:"S" enum:"size,name,none" help:"Sorting options" default:"none"`
	Humanize  bool   `short:"z" help:"Humanize File Sizes"`
	NoEmpty   bool   `short:"e" help:"Don't show empty files and folders'"`
	Large     bool   `short:"G" help:"Only print files over 1 GB"`
	FirstOnly bool   `short:"f" help:"Only show the first level of the filetree"`
	FilesOnly bool   `short:"F" help:"Only Print Files, Exclude all directories"`
	Regex     string `short:"r" help:"Search for files that match this regex in it's entirety (Search does a substring search)"`
	Search    string `short:"s" help:"Search all files in this folder for this text" default:""`
	NoCase    bool   `short:"i" help:"Use Case Insensitivity for Search"`
}

func printPath(file string, isDir bool, opts Opts, cache *files.FileCache) {
	var size int64
	var count uint
	if isDir {
		size, count = files.GetFolderInfo(file, *cache)
	} else {
		size = files.GetFileSize(file)
	}
	var additional = ""
	if isDir && opts.Verbose {
		// Add info on file count
		additional = fmt.Sprintf(" (#%d)", count)
	}
	if opts.Large && size < Gig || opts.NoEmpty && size == 0 {
		// File is less than a gig, quit
		return
	}
	fmt.Printf("%s\t- %s%s\n", utils.FormatSize(uint64(size), opts.Humanize), file,
		additional)
}

func printFolder(dir string, opts Opts, cache files.FileCache) {
	fileList := files.GetFiles(dir)
	sort.Slice(fileList, func(i, j int) bool {
		switch opts.Sort {
		default:
		case "none":
			break
		case "name":
			return strings.Compare(strings.ToLower(fileList[i].Name()), strings.ToLower(fileList[j].Name())) < 0
		case "size":
			return fileList[i].Size() > fileList[j].Size()
		}
		return true
	})
	for _, f := range fileList {
		handleFile(f, dir, opts, cache)
	}
}

func handleFile(f fs.FileInfo, dir string, opts Opts, cache files.FileCache) {
	if f.IsDir() {
		if !opts.FilesOnly {
			printPath(path.Join(dir, f.Name()), true, opts, &cache)
		}
		if !opts.FirstOnly {
			printFolder(path.Join(dir, f.Name()), opts, cache)
		}
	} else {
		printPath(path.Join(dir, f.Name()), false, opts, &cache)
	}
}

func main() {
	red := color.New(color.FgRed)
	green := color.New(color.FgGreen)
	var opts Opts
	ctx := kong.Parse(&opts)

	start := time.Now()

	if opts.Version {
		fmt.Printf("%s\n", version.VERSION)
		return
	}

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

	if len(opts.Search) > 0 || len(opts.Regex) > 0 {
		var r *regexp.Regexp
		if len(opts.Regex) > 0 {
			r, err = regexp.Compile(fmt.Sprintf(".*%s.*", opts.Regex))
		} else if opts.NoCase {
			r, err = regexp.Compile(fmt.Sprintf("(?i).*%s.*", opts.Search))
		} else {
			r, err = regexp.Compile(fmt.Sprintf(".*%s.*", opts.Search))
		}
		if err != nil {
			red.Printf("Couldn't parse %s into Golang Regex", opts.Search)
			return
		}
		items := files.ScanFiles(base)
		for file := range items {
			content, err := files.ReadEntire(file.Name)
			utils.NilCheckElse(err, func() {
				fmt.Printf("Could not read file %s: %+v\n", content, err)
			}, func() {
				if utils.ContainsIgnoreCase(content, r) {
					green.Printf("Found in %s\n", file.Name)
				} else if opts.Verbose {
					red.Printf("Not in %s\n", file.Name)
				}
			})
		}
		return
	}

	printFolder(base, opts, make(files.FileCache))

	if time.Now().Sub(start) > 100*time.Millisecond {
		fmt.Printf("Done in %s\n", humanize.RelTime(start, time.Now(), "", ""))
	}
	ctx.Exit(0)
}
