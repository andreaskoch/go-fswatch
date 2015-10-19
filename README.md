# fswatch

fswatch is a go library for recursively watching file system changes to **does not** depend on inotify and therefore is not limit by the ulimit of your operating system.

## Motivation

Why not use [inotify](http://en.wikipedia.org/wiki/Inotify)? Even though there are great libraries like [fsnotify](https://github.com/howeyc/fsnotify) that offer cross platform file system change notifications - the approach breaks when you want to watch a lot of files or folder.

For example the default ulimit for Mac OS is set to 512. If you need to watch more files you have to increase the ulimit for open files per process. And this sucks.

## Usage

### Watching a single file

If you want to watch a single file use the `NewFileWatcher` function to create a new file watcher:

```go
go func() {
	checkIntervalInSeconds := 2
	fileWatcher := fswatch.NewFileWatcher("Some-file", checkIntervalInSeconds).Start()

	for fileWatcher.IsRunning() {

		select {
		case <-fileWatcher.Modified:

			go func() {
				// file changed. do something.
			}()

		case <-fileWatcher.Moved:

			go func() {
				// file moved. do something.
			}()
		}

	}
}()
```

### Watching a folder

To watch a whole folder (and all of its child directories) for new, modified or deleted files you can use the `NewFolderWatcher` function.

Parameters:

1. The path to the directory you want to monitor
2. A flag indicating whether the folder shall be watched recursively
3. An expression which decides which files are skipped
4. The check interval in seconds (1 - n seconds)


```go
go func() {

	recurse := true // include all sub directories

	skipDotFilesAndFolders := func(path string) bool {
		return strings.HasPrefix(filepath.Base(path), ".")
	}

	checkIntervalInSeconds := 2

	folderWatcher := fswatch.NewFolderWatcher(
		"some-directory",
		recurse,
		skipDotFilesAndFolders,
		checkIntervalInSeconds
	)
	
	folderWatcher.Start()

	for folderWatcher.IsRunning() {

		select {

		case <-folderWatcher.Modified():
			fmt.Println("New or modified items detected")

		case <-folderWatcher.Moved():
			fmt.Println("Items have been moved")

		case changes := <-folderWatcher.ChangeDetails():

			fmt.Printf("%s\n", changes.String())
			fmt.Printf("New: %#v\n", changes.New())
			fmt.Printf("Modified: %#v\n", changes.Modified())
			fmt.Printf("Moved: %#v\n", changes.Moved())

		}
	}

}()
```
## go-fswatch in action

You can see go-fswatch in action in the **live-reload** feature of my [markdown webserver "allmark"](https://allmark.io/).

See:  [github.com/andreaskoch/allmark/blob/master/src/allmark.io/modules/dataaccess/filesystem/watcher.go](https://github.com/andreaskoch/allmark/blob/master/src/allmark.io/modules/dataaccess/filesystem/watcher.go)

I would still prefer using inotify, but go-fswatch has been doing it's job in allmark pretty well and works easily with relatively large folder structures.

## Build Status

[![Build Status](https://travis-ci.org/andreaskoch/go-fswatch.png?branch=master)](https://travis-ci.org/andreaskoch/go-fswatch)

## Contribute

All contributions are welcome. Especially if you have an idea

- how to reliably increase the limit for the maximum number of open files from within an application so we can use inotify for large folder structures.
- how to overcome the limitations of inotify without having to resort to checking the files for changes over and over again
- or how to make the existing code more efficient

please send me a message or a pull request.
