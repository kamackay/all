package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/semaphore"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/kamackay/all/browser"
	"github.com/kamackay/all/files"
	"github.com/kamackay/all/l"
	"github.com/kamackay/all/model"
	"github.com/kamackay/all/utils"
	"github.com/kamackay/all/version"
	"github.com/kamackay/godash/parallel"
)

const (
	Gig = 1000000000
)

func printPath(file *model.FileBean, opts model.Opts) {
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
	if opts.Verbose {
		additional += fmt.Sprintf(" [%s]", file.LastModified().Format(time.RFC3339))
	}
	if opts.Large && size < Gig || opts.NoEmpty && size == 0 || size > opts.MaxSize || size < opts.MinSize {
		// File is less than a gig, quit
		return
	}
	sizeString := utils.FormatSize(size, opts.Humanize)
	if opts.NamesOnly {
		fmt.Println(file.Name)
	} else {
		fmt.Printf("%s%s- %s%s\n", sizeString, utils.Spaces(spacing-len(sizeString)), file.Name,
			additional)
	}
}

func main() {
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)
	green := color.New(color.FgGreen)
	var opts model.Opts
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
		var bytes uint64 = 0
		var filesRead uint = 0
		x := 0
		writer := uilive.New()
		writer.Start()
		for file := range items {
			content, err := files.ReadEntire(file.Name)
			utils.NilCheckElse(err, func() {
				fmt.Printf("Could not read file %s: %+v\n", content, err)
			}, func() {
				bytes += uint64(len(content))
				filesRead++
				if utils.ContainsIgnoreCase(content, r) {
					green.Printf("Found in %s\n", file.Name)
				} else if opts.Verbose {
					red.Printf("Not in %s\n", file.Name)
				}
				if x != 0 && opts.Verbose {
					fmt.Fprintf(writer, "Read %d files (%s)\n", filesRead, utils.HumanizeBytes(bytes))
				}
			})
			x++
		}
		writer.Stop()
		return
	}

	cache := make(files.FileCache)

	var fileList []*model.FileBean
	if opts.FirstOnly {
		fileList = files.GetFilesFirstLevel(base, cache)
	} else {
		fileList = files.GetFilesRecursive(base)
	}
	sorter := func() model.SortFunction {
		switch opts.Sort {
		default:
		case "none":
			break
		case "time":
		case "modified":
			return func(i, j int) bool {
				return fileList[i].LastModified().After(fileList[j].LastModified())
			}
		case "name":
			return func(i, j int) bool {
				return strings.Compare(strings.ToLower(fileList[i].Name), strings.ToLower(fileList[j].Name)) < 0
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

	defer func() {
		if !opts.Quiet && time.Now().Sub(start) > 100*time.Millisecond {
			fmt.Printf("Done in %s\n", humanize.RelTime(start, time.Now(), "", ""))
		}
		ctx.Exit(0)
	}()

	if opts.RmEmpty {
		for _, f := range utils.FlipSlice(utils.Unique(fileList, func(file *model.FileBean) string { return file.Name })) {
			if f.IsDir() {
				empty, err := utils.IsEmpty(f.Name)
				if err != nil {
					red.Printf("Could not read %s: %+v\n", f.Name, err)
					continue
				}
				if empty {
					if opts.Verbose {
						fmt.Printf("Deleting empty directory %s\n", f.Name)
					}
					if opts.Yes || utils.AskForConfirmation(fmt.Sprintf("Delete empty directory %s?", f.Name)) {
						err := os.Remove(f.Name)
						if err != nil {
							red.Printf("Could not delete %s: %+v\n", f.Name, err)
						} else {
							green.Printf("Deleted %s\n", f.Name)
						}
					}
				}
			}
		}
		return
	}

	if opts.VideoScore {
		scoreFunc := func(bean *model.FileBean) *model.VideoScore {
			couldRecover := utils.GetPotentialBytesToCompress(bean)
			score, err := utils.GetVideoScore(bean)
			if err != nil {
				return model.NewScore(1000, bean, couldRecover)
			}
			return model.NewScore(score, bean, couldRecover)
		}
		sem := semaphore.NewWeighted(1)
		scores := make([]*model.VideoScore, 0)
		_ = parallel.ForEach(fileList, runtime.NumCPU(), func(f *model.FileBean) error {
			score := scoreFunc(f)
			defer func() {
				sem.Release(1)
			}()
			sem.Acquire(context.Background(), 1)
			scores = append(scores, score)
			return nil
		})
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].CouldRecover < scores[j].CouldRecover
		})
		for _, s := range scores {
			var message string
			score := s.Score
			if opts.Verbose {
				message = fmt.Sprintf("%0.2f (%s)\t\t- %s (could recover: %s)\n", score, utils.HumanizeBytes(s.Size), s.Name, utils.HumanizeBytes(uint64(s.CouldRecover)))
			} else {
				message = fmt.Sprintf("%0.2f\t\t- %s\n", score, s.Name)
			}
			if score > 20 {
				red.Printf(message)
			} else if score > 15 {
				yellow.Printf(message)
			} else if score > 0 {
				fmt.Printf(message)
			}
		}
		return
	}

	names := make(map[string]bool)

	verifyFirstTime := func(name string) bool {
		defer func() {
			names[name] = true
		}()
		_, ok := names[name]
		return !ok
	}

	if opts.Reverse {
		for x := len(fileList) - 1; x >= 0; x-- {
			f := fileList[x]
			if verifyFirstTime(f.Name) {
				printPath(f, opts)
			}
		}
	} else {
		for _, f := range fileList {
			if verifyFirstTime(f.Name) {
				printPath(f, opts)
			}
		}
	}

}
