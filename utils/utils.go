package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/kamackay/all/model"
)

const (
	Space                = " "
	TheoreticalBestScore = 10
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
				if !MapContains(firstStream, "height") || !MapContains(firstStream, "width") {
					// Wasn't a height and width, so skip
					return 0, nil
				}
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

func GetPotentialBytesToCompress(bean *model.FileBean) int64 {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in GetPotentialBytesToCompress", r)
		}
	}()
	out, err := exec.Command("ffprobe", bean.Name, "-show_streams", "-show_format", "-print_format", "json").Output()
	//fmt.Println(string(out))
	var output Json
	err = json.Unmarshal(out, &output)
	if err != nil {
		return 0
	}
	if val, ok := output["format"]; ok {
		if duration, ok := val.(Json)["duration"]; ok {
			durationString := duration.(string)
			duration, err := strconv.ParseFloat(durationString, 64)
			if err != nil {
				return 0
			}
			if istreamList, ok := output["streams"]; ok {
				streamList := istreamList.([]interface{})
				if len(streamList) == 0 {
					return 0
				}
				firstStream := streamList[0].(Json)
				if !MapContains(firstStream, "height") || !MapContains(firstStream, "width") {
					// Wasn't a height and width, so skip
					return 0
				}
				height := firstStream["height"].(float64)
				width := firstStream["width"].(float64)
				numPixels := duration * height * width
				idealSize := (TheoreticalBestScore * numPixels) / 100
				if float64(bean.Size) <= idealSize {
					return 0
				}
				return int64(float64(bean.Size) - idealSize)
			} else {
				return 0
			}
		} else {
			return 0
		}
	} else {
		return 0
	}
}

func AskForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func FlipSlice[T any](s []T) []T {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func Unique[T any, K comparable](slice []T, accessor func(T) K) []T {
	keys := make(map[K]bool)
	list := make([]T, 0)
	for _, entry := range slice {
		key := accessor(entry)
		if _, value := keys[key]; !value {
			keys[key] = true
			list = append(list, entry)
		}
	}
	return list
}

func MapContains[K comparable, T any](m map[K]T, key K) bool {
	_, ok := m[key]
	return ok
}
