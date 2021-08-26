package browser

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"gitlab.com/kamackay/all/files"
	"gitlab.com/kamackay/all/l"
	"gitlab.com/kamackay/all/model"
	"gitlab.com/kamackay/all/utils"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

type Browser struct {
	path          string
	Width, Height int
	SelectedLine  int
	Files         []File
	fileLoadMutex sync.Mutex
	loading       *model.LoadingInfo
	file          bool
}

func (b *Browser) getFiles() {
	path := b.path
	go func() {
		b.fileLoadMutex.Lock()
		defer b.fileLoadMutex.Unlock()
		defer b.Render()
		b.file = false
		info, _ := os.Stat(path)
		if info != nil && !info.IsDir() {
			b.file = true
			return
		}
		l.Print(fmt.Sprintf("Pulling files for %s", path))
		fs := files.GetFiles(path)
		fileList := make([]File, len(fs))
		l.Print(fmt.Sprintf("Pulled %d files for %s", len(fs), path))
		b.loading = &model.LoadingInfo{
			Item:  0,
			Total: len(fs),
		}
		for x, f := range fs {
			filename := filepath.Join(path, f.Name())
			fileList[x] = File{
				Path: filename,
				Size: files.GetSize(path, f),
			}
			b.loading.Item = x
			b.Render()
		}

		sort.Slice(fileList, func(i, j int) bool {
			return fileList[i].Size > fileList[j].Size
		})
		fileList = append([]File{
			makeRelativeFile(path, ".."),
		}, fileList...)
		b.loading = nil
		b.Files = fileList
	}()
}

func New(root string) (*Browser, error) {
	err := termbox.Init()
	if err != nil {
		return nil, err
	}
	w, h := termbox.Size()
	b := &Browser{
		path:         root,
		Width:        w,
		Height:       h,
		SelectedLine: 0,
	}
	b.getFiles()
	b.setSize(h, w)
	return b, nil
}

func (b *Browser) Run() {
	var lock sync.Mutex
	go func() {
		lock.Lock()
		defer b.Close()
		defer lock.Unlock()
		for {
			b.Render()
			e := b.Poll()
			if e.Ch == 'q' || e.Key == termbox.KeyCtrlC {
				return
			} else {
				b.keyPress(e)
			}
		}
	}()
	// Await the goroutine above exiting
	lock.Lock()
	defer lock.Unlock()
}

func (b *Browser) drawString(line string, y int, fg, bg termbox.Attribute) {
	for x := 0; x < len(line); x++ {
		termbox.SetCell(x, y, rune(line[x]), fg, bg)
	}
}

func (b *Browser) Render() {
	l.Error(termbox.Clear(termbox.ColorWhite, termbox.ColorDefault))
	defer termbox.Flush()
	height := b.Height - 1
	if b.loading != nil {
		text := fmt.Sprintf("Loading... %d of %d", b.loading.Item, b.loading.Total)
		b.drawString(text, 8, termbox.ColorGreen, termbox.ColorBlack)
		return
	}
	if b.file {
		b.drawString("You selected a file. That behavior is in the works", 8, termbox.ColorLightYellow, termbox.ColorBlack)
		return
	}
	line := 1
	b.drawString(fmt.Sprintf("Current: %s", b.path), 0, termbox.ColorLightMagenta, termbox.ColorBlack)
	lastItem := utils.Min(len(b.Files), height+b.SelectedLine)
	start := b.SelectedLine
	l.Print(fmt.Sprintf("Printing from %d to %d", start, lastItem))
	for y := start; y < lastItem; y++ {
		file := b.Files[y]
		text := fmt.Sprintf("%s -> %s", file.Path, utils.FormatSize(file.Size, true))
		fg := termbox.ColorGreen
		bg := termbox.ColorBlack
		if y == b.SelectedLine {
			fg = termbox.ColorBlack
			bg = termbox.ColorGreen
		}
		b.drawString(text, line, fg, bg)
		line++
	}
}

func (b *Browser) Poll() termbox.Event {
	for {
		switch e := termbox.PollEvent(); e.Type {
		case termbox.EventKey:
			return e
		case termbox.EventResize:
			b.setSize(e.Height, e.Width)
		}
	}
}

func (b *Browser) keyPress(e termbox.Event) {
	switch e.Key {
	case termbox.KeyArrowUp:
		b.setIndex(b.SelectedLine - 1)
		break
	case termbox.KeyArrowDown:
		b.setIndex(b.SelectedLine + 1)
		break
	case termbox.KeyArrowLeft:
		b.setIndex(0)
		b.Select()
		break
	case termbox.KeyEnter, termbox.KeyArrowRight:
		b.Select()
		break
	default:
		switch e.Ch {
		case 'r':
			// Refresh
			b.getFiles()
			break
		case '[':
			b.setIndex(0)
			break
		case ']':
			b.setIndex(len(b.Files) - 1)
			break
		default:
			l.Print(fmt.Sprintf("Unhandled Press on %s", string(e.Ch)))
		}
		break
	}
}

func (b *Browser) Close() {
	l.Print("Closin'!")
	termbox.Close()
}

func (b *Browser) Select() {
	newPath := b.Files[b.SelectedLine].Path
	l.Print("Selecting " + newPath)
	b.path = newPath
	b.setIndex(0)
	b.getFiles()
}

func (b *Browser) setSize(height int, width int) {
	l.Print(fmt.Sprintf("Setting Size to %dx%d", height, width))
	b.Height = height
	b.Width = width
}

func (b *Browser) setIndex(i int) {
	if i < 0 {
		i = 0
	}
	if i >= len(b.Files) {
		i = len(b.Files) - 1
	}
	b.SelectedLine = i
}
