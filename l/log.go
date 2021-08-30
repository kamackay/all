package l

import (
	"fmt"
	"os"
	"path/filepath"
)

func file() *string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting log file: %+v", err)
		return nil
	}
	path := filepath.Join(dirname, ".all.log")
	return &path
}

func Print(text string) {
	filename := file()
	if filename == nil {
		return
	}
	f, err := os.OpenFile(*filename,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text + "\n"); err != nil {
		fmt.Println(err)
	}
}

func Error(err error) {
	if err != nil {
		Print(fmt.Sprintf("Error! %+v", err))
	}
}

func Clear() {
	filename := file()
	if filename != nil {
		_ = os.Remove(*filename)
	}
}
