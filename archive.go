package main

/*
	put files into archives
	and push the archives into hpss
*/

import (
	"fmt"
	"os"
)

// DirEntry will get pushed through channels
type DirEntry struct {
	path string
	file os.FileInfo
}

// walk directory tree recursive, push files into channel <process>
func walk(dir string, process chan DirEntry) {
	f, err := os.Open(dir)
	if err == nil {
		direntries, _ := f.Readdir(0)
		for _, entry := range direntries {
			process <- DirEntry{dir, entry}
			if entry.IsDir() {
				walk(dir+"/"+entry.Name(), process)
			}
		}
	}

}

// handler reads file from channel and puts them in fiel lists,
// and pushes the file lists further down the pipe intp tar achive and hpss
// maxsize is in bytes here!
func handler(archive string, maxsize int64, process chan DirEntry) {
	currentsize := int64(0)
	currentarchive := 0
	for entry := range process {
		// finish current archive first?
		if currentsize != 0 && (currentsize+entry.file.Size() > maxsize) {
			fmt.Println("---------- ARCHIVE END --------#", currentarchive)
			currentarchive += 1
			currentsize = 0
		} else {
			fmt.Println(entry.path, entry.file.Name(), entry.file.Size())
			currentsize += entry.file.Size()
		}
	}
}

// archive a directory tree into a number of archives
func archive(archivename string, directory string, maxsize int) {
	// traverse tree
	process := make(chan DirEntry, 10)
	go handler(archivename, int64(maxsize), process)
	walk(directory, process)
	close(process)
}
