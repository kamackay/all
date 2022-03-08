package model

type Opts struct {
	Version   bool   `help:"Print Version"`
	Browser   bool   `short:"b" help:"Run Browser"`
	Verbose   bool   `short:"v" help:"Verbose"`
	Quiet     bool   `short:"q" help:"Only Log file info, exclude logs like time to process"`
	Directory string `arg:"d" help:"Directory" default:"."`
	Sort      string `short:"S" enum:"size,time,name,none" help:"Sorting options" default:"name"`
	Reverse   bool   `short:"r" help:"Reverse order of the list"`
	Humanize  bool   `short:"z" help:"Humanize File Sizes"`
	NoEmpty   bool   `short:"e" help:"Don't show empty files and folders'"`
	Large     bool   `short:"G" help:"Only print files over 1 GB"`
	FirstOnly bool   `short:"f" help:"Only show the first level of the filetree"`
	FilesOnly bool   `short:"F" help:"Only Print Files, Exclude all directories"`
	Regex     string `short:"r" help:"Search for files that match this regex in it's entirety (Search does a substring search)"`
	Search    string `short:"s" help:"Search all files in this folder for this text" default:""`
	NoCase    bool   `short:"i" help:"Use Case Insensitivity for Search"`
}
