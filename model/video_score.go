package model

type VideoScore struct {
	Name  string
	Score float64
	Size  uint64
}

func NewScore(score float64, bean *FileBean) *VideoScore {
	return &VideoScore{
		Name:  bean.Name,
		Score: score,
		Size:  bean.Size,
	}
}
