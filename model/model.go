package model

type SortType int

const (
	SortSize SortType = iota
	SortName
)

type LoadingInfo struct {
	Item    int
	Total   int
	Current string
	Render bool
}

type FileMode struct {
	Contents string
}

type Confirmation struct {
	Message string
	Action  func()
}

func SortTypeName(sortType SortType) string {
	switch sortType {
	case SortName:
		return "name"
	case SortSize:
		return "filesize"
	default:
		return "idk, randomly" // Again, shouldn't be possible
	}
}
