package browser

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"gitlab.com/kamackay/all/files"
	"gitlab.com/kamackay/all/l"
	"gitlab.com/kamackay/all/utils"
	"os"
	"path/filepath"
)

type Browser struct {
	path          string
	Width, Height int
	SelectedLine  int
	Files         []File
	StartIndex int
}

func getFiles(path string) []File {
	l.Print(fmt.Sprintf("Pulling files for %s", path))
	fs := files.GetFiles(path)
	fileList := make([]File, len(fs))
	l.Print(fmt.Sprintf("Pulled %d files for %s", len(fs), path))
	for x, f := range fs {
		filename := filepath.Join(path, f.Name())
		fileList[x] = File{
			Path: filename,
			Size: files.GetSize(path, f),
		}
	}
	parentPath := filepath.Join(path, "..")
	parent, err := os.Stat(parentPath)
	var parentSize uint64
	if err == nil {
		parentSize = files.GetSize(parentPath, parent)
	} else {
		l.Print(fmt.Sprintf("Error Getting Size: %+v\n", err))
	}
	fileList = append([]File{{
		Path: parentPath,
		Size: parentSize,
	}}, fileList...)
	return fileList
}

func New(root string) (*Browser, error) {
	_ = os.Remove(l.File)
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
		StartIndex: 0,
		Files:        getFiles(root),
	}
	b.setSize(h, w)
	return b, nil
}

func (b *Browser) Run() {
	defer b.Close()
	for {
		b.Render()
		e := b.Poll()
		if e.Ch == 'q' || e.Key == termbox.KeyCtrlC {
			return
		} else {
			b.keyPress(e)
		}
	}
}

func (b *Browser) Render() {
	l.Error(termbox.Clear(termbox.ColorWhite, termbox.ColorDefault))
	line := 0
	for y := b.StartIndex; y < len(b.Files); y++ {
		file := b.Files[y]
		text := fmt.Sprintf("%s -> %s", file.Path, utils.FormatSize(file.Size, true))
		fg := termbox.ColorGreen
		bg := termbox.ColorBlack
		if y == b.SelectedLine {
			fg = termbox.ColorBlack
			bg = termbox.ColorGreen
		}
		for x := 0; x < len(text); x++ {
			termbox.SetCell(x, line, rune(text[x]), fg, bg)
		}
		line++
	}
	err := termbox.Flush()
	if err != nil {
		l.Print(fmt.Sprintf("Error: %+v\n", err))
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
		b.SelectedLine = int(utils.Max(int64(b.SelectedLine-1), 0))
		if b.SelectedLine < b.StartIndex {
			b.StartIndex = b.SelectedLine
		}
		break
	case termbox.KeyArrowDown:
		b.SelectedLine = int(utils.Min(int64(b.SelectedLine+1), int64(len(b.Files)-1)))
		if b.SelectedLine >= b.Height {
			b.StartIndex++
		}
		break
	case termbox.KeyEnter:
		b.Select()
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
	b.Files = getFiles(newPath)
	b.path = newPath
	b.SelectedLine = 0
	b.StartIndex = 0
}

func (b *Browser) setSize(height int, width int) {
	l.Print(fmt.Sprintf("Setting Size to %dx%d", height, width))
	b.Height = height
	b.Width = width
}
