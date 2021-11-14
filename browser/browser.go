package browser

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/nsf/termbox-go"
	"github.com/skratchdot/open-golang/open"
	"github.com/kamackay/all/files"
	"github.com/kamackay/all/l"
	"github.com/kamackay/all/model"
	"github.com/kamackay/all/utils"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	green = termbox.ColorGreen
	black = termbox.ColorBlack
)

type Browser struct {
	path              string
	Width, Height     int
	SelectedLine      int
	Files             []File
	fileLoadMutex     sync.Mutex
	loading           *model.LoadingInfo
	file              *model.FileMode
	pollChan          chan *termbox.Event
	sort              model.SortType
	confirmations     []model.Confirmation
	timeReport        string
	autoUpdateEnabled bool
	updatedString     string
}

func (b *Browser) getFiles() {
	path := b.path
	go func() {
		start := time.Now()
		defer func() {
			defer b.update()
			now := time.Now()
			diff := now.Sub(start)
			if diff > time.Second {
				b.timeReport = fmt.Sprintf("Done in %s", humanize.RelTime(start, now, "", ""))
			} else {
				b.timeReport = fmt.Sprintf("Done in %dms", diff.Milliseconds())
			}
			b.updatedString = time.Now().Format("2006-01-02 15:04:05")
		}()
		b.fileLoadMutex.Lock()
		defer b.fileLoadMutex.Unlock()
		defer b.update()
		b.file = nil
		info, _ := os.Stat(path)
		if info != nil && !info.IsDir() {
			start, err := files.ReadStart(path, b.Height*b.Width)
			if err != nil {
				start = fmt.Sprintf("Error Reading file: %+v", err)
			}
			b.file = &model.FileMode{Contents: start}
			b.Files = []File{makeRelativeFile(path, "..")}
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
				Path:         filename,
				Size:         files.GetSize(path, f),
				LastModified: files.PrintTime(f),
				Dir:          f.IsDir(),
				Children:     files.CountChildren(filename),
			}
			b.loading.Item = x
			b.update()
		}
		sort.Slice(fileList, func(i, j int) bool {
			switch b.sort {
			case model.SortName:
				return strings.Compare(fileList[i].Path, fileList[j].Path) < 0
			case model.SortSize:
				return fileList[i].Size > fileList[j].Size
			default:
				return true // Shouldn't be possible, so whatever
			}
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
		path:              root,
		Width:             w,
		Height:            h,
		SelectedLine:      0,
		pollChan:          make(chan *termbox.Event),
		sort:              model.SortSize,
		confirmations:     make([]model.Confirmation, 0),
		autoUpdateEnabled: false,
	}
	b.getFiles()
	b.setSize(h, w)
	return b, nil
}

func (b *Browser) Run() {
	defer b.close()
	go b.poll()
	for {
		b.Render()
		select {
		case <-time.After(time.Second * 5):
			if b.loading != nil {
				break
			}
			if !b.autoUpdateEnabled {
				break
			}
			b.getFiles()
			break
		case e := <-b.pollChan:
			if e == nil {
				continue
			} else if e.Ch == 'q' || e.Key == termbox.KeyCtrlC {
				return
			} else {
				b.keyPress(*e)
			}
			break
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
	if loading := b.loading; loading != nil {
		text := fmt.Sprintf("Loading... %d of %d, currently: %s", loading.Item, loading.Total, loading.Current)
		b.drawString(text, 8, green, black)
		return
	} else if len(b.confirmations) > 0 {
		confirmation := b.confirmations[0]
		b.drawString(confirmation.Message, 8, green, black)
		b.drawString("Press y to confirm, n to dismiss", 9, green, black)
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
	b.drawString(fmt.Sprintf("Current: %s (Sorting by %s) [Auto Update: %s] (%s) {Updated: %s}", b.path,
		model.SortTypeName(b.sort),
		b.getAutoUpdateString(),
		strings.TrimSpace(b.timeReport),
		b.updatedString),
		0, termbox.ColorLightMagenta, termbox.ColorBlack)
	lastItem := utils.Min(len(b.Files), height+b.SelectedLine)
	start := b.SelectedLine
	//l.Print(fmt.Sprintf("Printing from %d to %d", start, lastItem))
	for y := start; y < lastItem; y++ {
		if y > len(b.Files)-1 {
			break
		}
		file := b.Files[y]
		text := ToString(file)
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

func (b *Browser) getAutoUpdateString() string {
	if b.autoUpdateEnabled {
		return "on"
	}
	return "off"
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
	case termbox.KeyDelete, termbox.KeyCtrlD:
		b.deleteCurrent()
		break
	default:
		switch e.Ch {
		case 'a':
			b.autoUpdateEnabled = true
			break
		case '~':
			// Set to Home Path
			dirname, err := os.UserHomeDir()
			l.Error(err)
			b.setPath(dirname)
			break
		case 's':
			switch b.sort {
			case model.SortSize:
				b.sort = model.SortName
				break
			case model.SortName:
				b.sort = model.SortSize
				break
			}
			b.getFiles()
			break
		case 'n':
			if len(b.confirmations) > 0 {
				// There is a pending confirmation, remove it
				b.confirmations = b.confirmations[1:]
			}
			break
		case 'y':
			if len(b.confirmations) > 0 {
				// There is a pending confirmation, confirm it
				confirmation := b.confirmations[0]
				b.confirmations = b.confirmations[1:]
				confirmation.Action()
			}
			break
		case 'o':
			_ = open.Run(b.Files[b.SelectedLine].Path)
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
			l.Print(fmt.Sprintf("Unhandled Press %+v", e))
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

func (b *Browser) getCurrentFile() File {
	return b.Files[b.SelectedLine]
}

func (b *Browser) Select() {
	newPath := b.getCurrentFile().Path
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

func (b *Browser) deleteCurrent() {
	path := b.getCurrentFile().Path
	b.confirmations = append(b.confirmations, model.Confirmation{
		Message: fmt.Sprintf("Are you sure you want to delete %s?", path),
		Action: func() {
			l.Print(fmt.Sprintf("Deleting %s", path))
			l.Error(os.RemoveAll(path))
			b.getFiles()
		},
	})
}
