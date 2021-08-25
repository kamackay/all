package utils

import (
	"fmt"
	"github.com/dustin/go-humanize"
)

const (
	Tab = "  "
)

func Indentation(index int) string {
	str := ""
	for i := 1; i < index; i++ {
		str += Tab
	}
	return str + "| "
}

func FormatSize(sizeBytes uint64, human bool) string {
	if human {
		return humanize.Bytes(sizeBytes)
	} else {
		return fmt.Sprintf("%d", sizeBytes)
	}
}

func Max(x, y int) int {
	if x >= y {
		return x
	}
	return y
}

func Min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}