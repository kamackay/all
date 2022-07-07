package unique

import (
	"github.com/kamackay/all/model"
	"io/fs"
)

func Infos(input []fs.FileInfo) []fs.FileInfo {
	u := make([]fs.FileInfo, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val.Name()]; !ok {
			m[val.Name()] = true
			u = append(u, val)
		}
	}
	return u
}

func FileBeans(input []*model.FileBean) []*model.FileBean {
	u := make([]*model.FileBean, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val.Name]; !ok {
			m[val.Name] = true
			u = append(u, val)
		}
	}
	return u
}
