# all
CLI tool to list out files in a folder recursively, track file system size with a built in interactive browser, and search for file contents recursively

## Installation
```
> brew install kamackay/homebrew-tap/all
```

### List all files recursively in current folder
```
> all
```

### List all files in a given folder
```
> all ~/files
```

### List only first level of files
```
> all -f
```

### Show file size in human readable format
```
> all -h
```

### Search for a string inside all files in a directory recursively
```
> all -s "hello world" ~/files
```

### Launch interactive filesystem browser
```
> all -b
```

#### Browser Commands

- Arrow Up/Down: Navigate in current folder
- Left Arrow: Go one folder higher in directory
- Right Arrow/Enter: Drill into currently selected folder
- Delete/Ctrl+d: Delete current highlighted file (will show a prompt first)
- 'a': Turn on auto update, will refresh current folder every 5 seconds
- 'q'/ctrl+c: Exit
- '~': Go to home directory
- 's': Toggle Sort Mode, currently supports by name or by file size
- 'o': Open current file (calls Golang's `open-golang Run function`)
- 'r': Refresh current folder
- '\[': Go to top of list
- '\]': Go to bottom of list
