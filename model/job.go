package model

import "os"

type Job struct {
	Action func(path string, info os.FileInfo) interface{}
	Info   os.FileInfo
	Path   string
}

func (c *Job) Call() interface{} {
	//Do task
	return c.Action(c.Path, c.Info)
}
