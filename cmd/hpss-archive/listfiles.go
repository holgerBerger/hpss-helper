package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func listfiles(archive string) {
	file, err := os.Open(config.General.Cachedir + "/" + archive + ".full")
	if err == nil {
		bfile := bufio.NewReader(file)
		archives := int64(0)
		files := 0
		bytes := int64(0)
		for {
			line, err := bfile.ReadBytes('\n')
			if err != nil {
				break
			}
			fields := strings.Split(string(line), "|")
			fmt.Printf("%s   (%s bytes)\n", fields[0], fields[1])
			curbytes, _ := strconv.ParseInt(fields[1], 10, 64)
			curarchive, _ := strconv.ParseInt(strings.TrimSpace(fields[2]), 10, 32)
			if curarchive > archives {
				archives = curarchive
			}
			bytes += curbytes
			files++
		}
		fmt.Println("\n", bytes/(1204*1024), "MB in", files, "files kept in total of", archives, "archive fragments")
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
