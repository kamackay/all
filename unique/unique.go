package unique

import (
	"io/fs"

	"github.com/kamackay/all/model"
	"github.com/samber/lo"
)

func Infos(input []fs.FileInfo) []fs.FileInfo {
	return lo.UniqBy(input, func(item fs.FileInfo) string {
		return item.Name()
	})
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
