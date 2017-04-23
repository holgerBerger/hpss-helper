package main

import (
	"bufio"
	"fmt"
	"github.com/holgerBerger/hpsshelper"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

var fetchWaiter sync.WaitGroup

type fetchT struct {
	tarname  string
	filelist []string
}

func extract(archive string, patterns []string) {
	archiveset := make(map[string][]string)
	file, err := os.Open(config.General.Cachedir + "/" + archive + ".full")
	if err != nil {
		file, err := os.Open(config.General.Cachedir + "/" + archive + ".full-offline")
		if err != nil {
			log.Println("archive does not exist.")
		} else {
			log.Println("archive index is offline.")
		}
		file.Close()
		return
	}

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
		for _, pattern := range patterns {
			ok, err = path.Match(pattern, path.Base(fields[0]))
			if ok && err == nil {
				curbytes, _ := strconv.ParseInt(fields[1], 10, 64)
				filelist, ok := archiveset[strings.TrimSpace(fields[2])]
				if ok {
					filelist = append(filelist, string(fields[0]))
					archiveset[strings.TrimSpace(fields[2])] = filelist
				} else {
					filelist := make([]string, 128)
					filelist = append(filelist, fields[0])
					archiveset[strings.TrimSpace(fields[2])] = filelist
				}
				bytes += curbytes
				files++
			}
		}
	}
	log.Println("fetching", bytes/(1204*1024), "MB in", files,
		"files kept in total of", len(archiveset), "archive fragments")
	file.Close()

	fetchWaiter.Add(1)
	getch := make(chan fetchT, 128)

	for i := 0; i < config.General.Hpssmovers; i++ {
		go getworker(getch)
	}

	for a, fl := range archiveset {
		an, _ := strconv.ParseInt(a, 10, 32)
		tf := fmt.Sprintf("%s-%9.9d.tar", archive, an)
		getch <- fetchT{tf, fl}
	}
	close(getch)
	fetchWaiter.Done()
	log.Println("waiting for HPSS")
	fetchWaiter.Wait()
	log.Println("finished waiting for HPSS")
}

func getworker(getch chan fetchT) {
	fetchWaiter.Add(1)
	cwd, _ := os.Getwd()
	os.Chdir(config.General.Workdir)
	pftp := hpsshelper.NewPftp(config.Hpss)
	pftp.Cd(config.Hpss.Hpssbasedir)
	for a := range getch {
		log.Println("  fetching", a.tarname)
		pftp.Get(a.tarname)
		log.Println("  finished fetching", a.tarname)
		log.Println("  extracting", a.filelist)
	}
	pftp.Bye()
	os.Chdir(cwd)
	fetchWaiter.Done()
}
