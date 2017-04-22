package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func listfiles(archive, pattern string) {
	patternmatch := pattern != ""
	archiveset := make(map[string]bool)
	file, err := os.Open(config.General.Cachedir + "/" + archive + ".full")
	if err == nil {
		bfile := bufio.NewReader(file)
		files := 0
		bytes := int64(0)
		for {
			line, err := bfile.ReadBytes('\n')
			if err != nil {
				break
			}
			fields := strings.Split(string(line), "|")
			var ok bool
			if patternmatch {
				ok, err = path.Match(pattern, path.Base(fields[0]))
				// fmt.Println("match", path.Base(fields[0]), pattern, ok, err)
			} else {
				ok = true
				err = nil
			}
			if ok && err == nil {
				fmt.Printf("%s   (%s bytes)\n", fields[0], fields[1])
				curbytes, _ := strconv.ParseInt(fields[1], 10, 64)
				archiveset[strings.TrimSpace(fields[2])] = true
				bytes += curbytes
				files++
			}
		}
		fmt.Println("\n", bytes/(1204*1024), "MB in", files,
			"files kept in total of", len(archiveset), "archive fragments")
		file.Close()
	} else {
		file, err := os.Open(config.General.Cachedir + "/" + archive + ".full-offline")
		if err != nil {
			fmt.Println("archive does not exist.")
		} else {
			fmt.Println("archive index is offline.")
		}
		file.Close()
	}
}
