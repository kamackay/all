package utils

import (
	"fmt"
	"math"
	"regexp"
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
		return HumanizeBytes(sizeBytes)
	} else {
		return fmt.Sprintf("%d", sizeBytes)
	}
}

func HumanizeBytes(s uint64) string {
	sizes := []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
	return humanizeBytes(s, 1000, sizes)
}

func humanizeBytes(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logN(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.1f %s"
	return fmt.Sprintf(f, val, suffix)
}

func logN(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
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

func ContainsIgnoreCase(a string, r *regexp.Regexp) bool {
	return r.MatchString(a)
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
