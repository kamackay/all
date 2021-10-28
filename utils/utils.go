package utils

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"strings"
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

func ContainsIgnoreCase(a string, b string) bool {
	return strings.Contains(
		strings.ToLower(a),
		strings.ToLower(b),
	)
}

func NilCheck(obj interface{}, action func()) {
	if obj != nil {
		action()
	}
}

func NilCheckElse(obj interface{}, action func(), elseAction func()) {
	if obj != nil {
		action()
	} else {
		elseAction()
	}
}

