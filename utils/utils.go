package utils

import (
	"encoding/json"
	"fmt"
	"github.com/kamackay/all/model"
	"math"
	"os/exec"
	"regexp"
	"strconv"
)

const (
	Space = " "
)

func Spaces(index int) string {
	str := ""
	for i := 1; i < index; i++ {
		str += Space
	}
	return str
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

func Min(x, y int) int {
	if x <= y {
		return x
	}
	return y
}

func ContainsIgnoreCase(a string, r *regexp.Regexp) bool {
	return r.MatchString(a)
}

func NilCheckElse(obj interface{}, action func(), elseAction func()) {
	if obj != nil {
		action()
	} else {
		elseAction()
	}
}

type Json = map[string]interface{}

func GetVideoScore(bean *model.FileBean) (float64, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	out, err := exec.Command("ffprobe", bean.Name, "-show_streams", "-show_format", "-print_format", "json").Output()
	//fmt.Println(string(out))
	var output Json
	err = json.Unmarshal(out, &output)
	if err != nil {
		return 0, err
	}
	if val, ok := output["format"]; ok {
		if duration, ok := val.(Json)["duration"]; ok {
			durationString := duration.(string)
			duration, err := strconv.ParseFloat(durationString, 64)
			if err != nil {
				return 0, err
			}
			if istreamList, ok := output["streams"]; ok {
				streamList := istreamList.([]interface{})
				if len(streamList) == 0 {
					return 0, nil
				}
				firstStream := streamList[0].(Json)
				height := firstStream["height"].(float64)
				width := firstStream["width"].(float64)
				numPixels := duration * height * width
				return (float64(bean.Size) / numPixels) * 100, nil // Bytes per pixel
			} else {
				return 0, nil
			}
		} else {
			return 0, nil
		}
	} else {
		return 0, nil
	}
}
