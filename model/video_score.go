package model

type VideoScore struct {
	Name         string
	Score        float64
	Size         uint64
	CouldRecover int64
}

func NewScore(score float64, bean *FileBean, couldRecover int64) *VideoScore {
	return &VideoScore{
		Name:         bean.Name,
		Score:        score,
		Size:         bean.Size,
		CouldRecover: couldRecover,
	}
}
