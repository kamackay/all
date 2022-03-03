package main

import (
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/kamackay/all/browser"
	"github.com/kamackay/all/files"
	"github.com/kamackay/all/l"
	"github.com/kamackay/all/model"
	"github.com/kamackay/all/utils"
	"github.com/kamackay/all/version"
	"os"
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
	Quiet     bool   `short:"q" help:"Only Log file info, exclude logs like time to process"`
	Directory string `arg:"d" help:"Directory" default:"."`
	Sort      string `short:"S" enum:"size,size-invert,name,none" help:"Sorting options" default:"name"`
	Humanize  bool   `short:"z" help:"Humanize File Sizes"`
	NoEmpty   bool   `short:"e" help:"Don't show empty files and folders'"`
	Large     bool   `short:"G" help:"Only print files over 1 GB"`
	FirstOnly bool   `short:"f" help:"Only show the first level of the filetree"`
	FilesOnly bool   `short:"F" help:"Only Print Files, Exclude all directories"`
	Regex     string `short:"r" help:"Search for files that match this regex in it's entirety (Search does a substring search)"`
	Search    string `short:"s" help:"Search all files in this folder for this text" default:""`
	NoCase    bool   `short:"i" help:"Use Case Insensitivity for Search"`
}

func printPath(file *model.FileBean, opts Opts) {
	if file.IsDir() && opts.FilesOnly {
		return
	}
	var spacing int
	if opts.Humanize {
		spacing = 11
	} else {
		spacing = 16
	}
	size := file.Size
	var additional = ""
	if file.IsDir() && opts.Verbose {
		// Add info on file count
		additional = fmt.Sprintf(" (#%d)", file.Count)
	}
	if opts.Large && size < Gig || opts.NoEmpty && size == 0 {
		// File is less than a gig, quit
		return
	}
	sizeString := utils.FormatSize(size, opts.Humanize)
	fmt.Printf("%s%s- %s%s\n", sizeString, utils.Spaces(spacing-len(sizeString)), file.Name,
		additional)
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

	cache := make(files.FileCache)

	var fileList []*model.FileBean
	if opts.FirstOnly {
		fileList = files.GetFilesFirstLevel(base, cache)
	} else {
		fileList = files.GetFilesRecursive(base, cache)
	}
	sorter := func() model.SortFunction {
		switch opts.Sort {
		default:
		case "none":
			break
		case "name":
			return func(i, j int) bool {
				return strings.Compare(strings.ToLower(fileList[i].Name), strings.ToLower(fileList[j].Name)) < 0
			}
		case "size-invert":
			return func(i, j int) bool {
				return fileList[i].Size > fileList[j].Size
			}
		case "size":
			return func(i, j int) bool {
				return fileList[i].Size < fileList[j].Size
			}
		}
		return func(i, j int) bool {
			return i < j
		}
	}()
	sort.Slice(fileList, sorter)

	for _, f := range fileList {
		printPath(f, opts)
	}

	if !opts.Quiet && time.Now().Sub(start) > 100*time.Millisecond {
		fmt.Printf("Done in %s\n", humanize.RelTime(start, time.Now(), "", ""))
	}
	ctx.Exit(0)
}
