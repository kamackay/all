package unique

import "io/fs"

func Infos(input []fs.FileInfo) []fs.FileInfo {
	u := make([]fs.FileInfo, 0, len(input))
	m := make(map[fs.FileInfo]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}
