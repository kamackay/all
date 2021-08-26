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
	"strings"
	"sync"
)

const (
	green = termbox.ColorGreen
	black = termbox.ColorBlack
)

type Browser struct {
	path          string
	Width, Height int
	SelectedLine  int
	Files         []File
	fileLoadMutex sync.Mutex
	loading       *model.LoadingInfo
	file          *model.FileMode
	pollChan      chan *termbox.Event
}

func (b *Browser) getFiles() {
	path := b.path
	go func() {
		b.fileLoadMutex.Lock()
		defer b.fileLoadMutex.Unlock()
		defer b.update()
		b.file = nil
		info, _ := os.Stat(path)
		if info != nil && !info.IsDir() {
			start, err := files.ReadStart(path, b.Height * b.Width)
			if err != nil {
				start = fmt.Sprintf("Error Reading file: %+v", err)
			}
			b.file = &model.FileMode{Contents: start}
			b.update()
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
			b.loading.Current = f.Name()
			b.update()
			filename := filepath.Join(path, f.Name())
			fileList[x] = File{
				Path: filename,
				Size: files.GetSize(path, f),
			}
			b.loading.Item = x
			b.update()
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
		pollChan:     make(chan *termbox.Event),
	}
	b.getFiles()
	b.setSize(h, w)
	return b, nil
}

func (b *Browser) Run() {
	//time.AfterFunc(time.Minute, b.kill)
	defer b.close()
	go b.poll()
	for {
		b.Render()
		e := <-b.pollChan
		if e == nil {
			continue
		} else if e.Ch == 'q' || e.Key == termbox.KeyCtrlC {
			return
		} else {
			b.keyPress(*e)
		}
	}
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
		text := fmt.Sprintf("Loading... %d of %d, currently: %s", b.loading.Item, b.loading.Total, b.loading.Current)
		b.drawString(text, 8, green, black)
		return
	}
	if b.file != nil {
		lines := strings.Split(b.file.Contents, "\n")
		for x, line := range lines {
			b.drawString(line, x+2, green, black)
		}
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
		fg := green
		bg := black
		if y == b.SelectedLine {
			fg = black
			bg = green
		}
		b.drawString(text, line, fg, bg)
		line++
	}
}

func (b *Browser) wipe() {
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorBlack)
		}
	}
}

func (b *Browser) poll() {
	for {
		event := termbox.PollEvent()
		switch event.Type {
		case termbox.EventKey:
			b.pollChan <- &event
		case termbox.EventResize:
			b.setSize(event.Height, event.Width)
			b.update()
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
		case '~':
			// Set to Home Path
			dirname, err := os.UserHomeDir()
			l.Error(err)
			b.setPath(dirname)
			break
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

func (b *Browser) update() {
	b.pollChan <- nil
}

func (b *Browser) kill() {
	b.close()
	os.Exit(0)
}

func (b *Browser) close() {
	l.Print("Closin'!")
	termbox.Close()
}

func (b *Browser) setPath(path string) {
	b.path = path
	b.setIndex(0)
	b.getFiles()
}

func (b *Browser) Select() {
	newPath := b.Files[b.SelectedLine].Path
	l.Print("Selecting " + newPath)
	b.setPath(newPath)
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
