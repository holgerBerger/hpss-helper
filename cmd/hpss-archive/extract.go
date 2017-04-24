package main

import (
	"bufio"
	"fmt"
	"github.com/holgerBerger/hpsshelper"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var fetchWaiter sync.WaitGroup

type fetchT struct {
	tarname  string
	filelist []string
}

var CWD string

func extract(archive string, patterns []string) {
	CWD, _ = os.Getwd()
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
		var ok1, ok2 bool
		for _, pattern := range patterns {
			ok1, err = path.Match(pattern, path.Base(fields[0]))
			ok2, err = path.Match(pattern, fields[0])
			if (ok1 || ok2) && err == nil {
				curbytes, _ := strconv.ParseInt(fields[1], 10, 64)
				filelist, ok := archiveset[strings.TrimSpace(fields[2])]
				if ok {
					filelist = append(filelist, string(fields[0]))
					archiveset[strings.TrimSpace(fields[2])] = filelist
				} else {
					filelist := make([]string, 0, 128)
					filelist = append(filelist, fields[0])
					archiveset[strings.TrimSpace(fields[2])] = filelist
				}
				bytes += curbytes
				files++
			}
		}
	}
	log.Println("fetching", bytes/(1024.0*1024.0), "MB in", files,
		"files kept in total of", len(archiveset), "archive fragments")
	file.Close()

	fetchWaiter.Add(1)
	getch := make(chan fetchT, 128)
	tarWaiter.Add(1)
	tarch := make(chan fetchT, 128)

	// spawn tar workers
	for i := 0; i < config.General.Tars; i++ {
		go tarWorker(tarch)
	}

	// spawn movers
	for i := 0; i < config.General.Hpssmovers; i++ {
		go getWorker(getch, tarch)
	}

	for a, fl := range archiveset {
		an, _ := strconv.ParseInt(a, 10, 32)
		tf := fmt.Sprintf("%s-%9.9d.tar", archive, an)
		if len(patterns) == 1 && patterns[0] == "*" {
			getch <- fetchT{tf, []string{}}
		} else {
			getch <- fetchT{tf, fl}
		}
	}
	runtime.Gosched()
	close(getch)
	fetchWaiter.Done()
	log.Println("waiting for HPSS")
	fetchWaiter.Wait()
	log.Println("finished waiting for HPSS")
	tarWaiter.Done()
	close(tarch)
	log.Println("waiting for tar")
	tarWaiter.Wait()
	log.Println("finished waiting for tar")
}

func getWorker(getch chan fetchT, tarch chan fetchT) {
	fetchWaiter.Add(1)
	cwd, _ := os.Getwd()
	os.Chdir(config.General.Workdir)
	pftp := hpsshelper.NewPftp(config.Hpss)
	pftp.Cd(config.Hpss.Hpssbasedir)
	for a := range getch {
		log.Println(" fetching", a.tarname)
		pftp.Get(a.tarname)
		log.Println(" finished fetching", a.tarname)
		tarch <- a
	}
	pftp.Bye()
	os.Chdir(cwd)
	fetchWaiter.Done()
}

func tarWorker(tarch chan fetchT) {
	tarWaiter.Add(1)
	for f := range tarch {
		log.Println("  extracting from ", config.General.Workdir+"/"+f.tarname)
		arglist := []string{"-C", CWD, "-xvf", config.General.Workdir + "/" + f.tarname}
		arglist = append(arglist, f.filelist...)
		out, err := exec.Command("/bin/tar", arglist...).CombinedOutput()
		if err != nil {
			log.Println(string(out))
			log.Println("extracting: ", f.filelist)
			log.Fatal("error while extracting from "+f.tarname, " ", err)
		}
		log.Println("  finished extracting from ", config.General.Workdir+"/"+f.tarname)
	}
	tarWaiter.Done()
}
