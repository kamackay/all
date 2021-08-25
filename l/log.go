package l

import (
	"fmt"
	"os"
)

const (
	File = "/Users/keithmackay/all.log"
)

func Print(text string) {
	f, err := os.OpenFile(File,
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