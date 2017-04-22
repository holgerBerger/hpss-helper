package main

/*
	put files into archives
	and push the archives into hpss
*/

import (
	//	"bufio"
	"fmt"
	"github.com/holgerBerger/hpsshelper"
	//	"io"
	"log"
	"os"
	"os/exec"
	"path"
	//	"strconv"
	//	"strings"
	"sync"
	"time"
)

// dirEntry will get pushed through channels
type dirEntry struct {
	path string
	file os.FileInfo
}

// tarFile will get pushed torugh tarchannel
type tarFile struct {
	tarfilename  string
	filelistname string
	size         int64
}

var totalbytes int64
var firstHpsstransferset bool
var firstHpsstransfer time.Time
var tarWaiter sync.WaitGroup
var hpssWaiter sync.WaitGroup
var hpsschannel chan string

// walk directory tree recursive, push files into channel <process>
func walk(dir string, process chan dirEntry) {
	f, err := os.Open(dir)
	if err == nil {
		direntries, _ := f.Readdir(0)
		for _, entry := range direntries {
			process <- dirEntry{dir, entry}
			if entry.IsDir() {
				walk(dir+"/"+entry.Name(), process)
			}
		}
	}

}

// handler reads file from channel and puts them in file lists,
// and pushes the file lists further down the pipe into tar achive and hpss
// maxsize is in bytes here!
func fileHandler(archive string, maxsize int64, process chan dirEntry) {
	currentsize := int64(0)
	currentarchive := 0
	cachedir := os.ExpandEnv(config.General.Cachedir)
	workdir := os.ExpandEnv(config.General.Workdir)

	fullcatalogfile, err := os.Create(cachedir + "/" + archive + ".full")
	if err != nil {
		log.Println("could not open catalog file for writing!")
		os.Exit(1)
	}
	currentcatalogname := fmt.Sprintf(workdir+"/"+archive+"-%9.9d.cat", currentarchive)
	catalogfile, err := os.Create(currentcatalogname)
	if err != nil {
		log.Println("could not open catalog file for writing!")
		os.Exit(1)
	}

	// spawn hpps movers
	hpsschannel = make(chan string, 1024)
	for i := 1; i < config.General.Hpssmovers; i++ {
		go hpssHandler(hpsschannel)
	}

	// spawn tar processes
	tarchannel := make(chan tarFile, 1024)
	for i := 1; i < config.General.Tars; i++ {
		go tarHandler(tarchannel, hpsschannel)
	}

	for entry := range process {
		// finish current archive first?
		if currentsize != 0 && (currentsize+entry.file.Size() > maxsize) {
			catalogfile.Close()
			tarchannel <- tarFile{fmt.Sprintf(workdir+"/"+archive+"-%9.9d.tar",
				currentarchive), currentcatalogname, currentsize}
			// switch to next archive
			currentarchive += 1
			currentsize = 0
			currentcatalogname = fmt.Sprintf(workdir+"/"+archive+"-%9.9d.cat",
				currentarchive)
			catalogfile, err = os.Create(currentcatalogname)
			if err != nil {
				log.Println("could not open catalog file for writing!")
				os.Exit(1)
			}
		}
		fmt.Fprintf(fullcatalogfile, "%s|%d|%d\n", entry.path+"/"+entry.file.Name(),
			entry.file.Size(), currentarchive)
		fmt.Fprintln(catalogfile, entry.path+"/"+entry.file.Name())
		currentsize += entry.file.Size()
		totalbytes += entry.file.Size()
	}

	// write last files
	catalogfile.Close()
	tarchannel <- tarFile{fmt.Sprintf(workdir+"/"+archive+"-%9.9d.tar", currentarchive),
		currentcatalogname, currentsize}

	// close catalogue and send to hpss
	fullcatalogfile.Close()
	hpsschannel <- cachedir + "/" + archive + ".full"

	close(tarchannel)
}

// tarHandler receives file lists over channel and produces tar archives
func tarHandler(tarchannel chan tarFile, hpsschannel chan string) {
	tarWaiter.Add(1)
	for tar := range tarchannel {
		log.Print(" archiving into ", tar.tarfilename /*, " from ", tar.filelistname */)
		out, err := exec.Command("/bin/tar", "--no-recursion", "-T", tar.filelistname,
			"-cf", tar.tarfilename).Output()
		_ = out
		if err != nil {
			log.Fatal("error while writing "+tar.tarfilename, err)
		}
		log.Print(" finished archiving into ", tar.tarfilename, "  ",
			"(", tar.size/(1024*1024), " MB)")
		hpsschannel <- tar.tarfilename
	}
	tarWaiter.Done()
}

// hpssHandler uses pftp class for communication, latest version
// one session is used for all transfers
func hpssHandler(hpsschannel chan string) {
	hpssWaiter.Add(1)

	pftp := hpsshelper.NewPftp(config.Hpss)
	if pftp == nil {
		log.Fatal("error in init")
	}

	err := pftp.Cd(config.Hpss.Hpssbasedir)
	if err != nil {
		log.Fatal("error in cd")
	}

	for tarfile := range hpsschannel {
		if !firstHpsstransferset {
			firstHpsstransfer = time.Now()
			firstHpsstransferset = true
		}
		log.Print("  sending to hpss ", tarfile)

		err := pftp.Put(tarfile, path.Base(tarfile))
		if err != nil {
			log.Fatal("error in put for ", tarfile)
			log.Fatal("FTP protocoll -----------------------")
			log.Fatal(pftp.Protocoll.String())
			log.Fatal("FTP protocoll end -------------------")
		}

		log.Print("  finished sending to hpss ", tarfile)
	}

	pftp.Bye()

	hpssWaiter.Done()
}

// some unused, old code

/*
// better hpss handler, logs in once, pushes several files within one session
func hpssHandler2(hpsschannel chan string) {
	// out := make([]byte, 8192)
	hpssWaiter.Add(1)
	cmd := exec.Command(config.Hpss.Pftp_client, "-w2", "-inv",
		config.Hpss.Hpssserver, strconv.Itoa(config.Hpss.Hpssport))
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal("error calling pftp_client", err)
	}
	stdout, err := cmd.StdoutPipe()
	bstdout := bufio.NewReader(stdout)

	cmd.Start()

	io.WriteString(stdin, "quote USER "+config.Hpss.Hpssusername+"\n")
	io.WriteString(stdin, "quote pass "+config.Hpss.Hpsspassword+"\n")
	io.WriteString(stdin, "cd "+config.Hpss.Hpssbasedir+"\n")

	// this is a timeout hack
	ch := make(chan bool)
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			out, _ := bstdout.ReadBytes('\n')
			if out == nil {
				log.Fatal("error in hpss login!")
				return
			}
			if strings.Index(string(out), "250 CWD") != -1 {
				break
			}
		}
		ch <- true
		// TODO collect output and print error message if timeout is triggered
	}()
	select {
	case <-ch:
		log.Println("  hpss login and cd succesfull")
	case <-time.After(10 * time.Second):
		log.Fatal("error in hpss login!")
		os.Exit(1)
	}

	for tarfile := range hpsschannel {
		if !firstHpsstransferset {
			firstHpsstransfer = time.Now()
			firstHpsstransferset = true
		}
		log.Print("  sending to hpss ", tarfile)

		stdin.Write([]byte("put " + tarfile + " " + path.Base(tarfile) + "\n"))
		stdin.Write([]byte("\n"))

		for {
			time.Sleep(100 * time.Millisecond)
			// fmt.Println("...waiting...")
			out, err := bstdout.ReadBytes('\n')
			if err != nil {
				fmt.Printf("%s\n", out)
				log.Fatal(err)
			}
			// fmt.Printf("[%s]", out)
			if strings.Index(string(out), "226 Transfer ") != -1 {
				break
			}
			if strings.Index(string(out), "HPSS Error:") != -1 {
				log.Fatal("error in transfer of ", tarfile)
				break
			}
		}
		log.Print("  finished sending to hpss ", tarfile)
	}

	io.WriteString(stdin, "bye\n")
	cmd.Wait()

	hpssWaiter.Done()
}

// simple hpss handler, does complete session for each transfer, robust
func hpssHandler1(hpsschannel chan string) {
	hpssWaiter.Add(1)
	for tarfile := range hpsschannel {
		if !firstHpsstransferset {
			firstHpsstransfer = time.Now()
			firstHpsstransferset = true
		}
		log.Print("  sending to hpss ", tarfile)
		cmd := exec.Command(config.Hpss.Pftp_client, "-w2", "-inv",
			config.Hpss.Hpssserver, strconv.Itoa(config.Hpss.Hpssport))
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Fatal("error calling pftp_client", err)
		}

		go func() {
			defer stdin.Close()
			io.WriteString(stdin, "quote USER "+config.Hpss.Hpssusername+"\n")
			io.WriteString(stdin, "quote pass "+config.Hpss.Hpsspassword+"\n")
			io.WriteString(stdin, "cd "+config.Hpss.Hpssbasedir+"\n")
			io.WriteString(stdin, "put "+tarfile+" "+path.Base(tarfile)+"\n")
			io.WriteString(stdin, "bye\n")
		}()

		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("%s\n", out)
			log.Fatal(err)
		}
		if strings.Index(string(out), "226 Transfer ") == -1 {
			fmt.Printf("%s\n", out)
			log.Fatal("error in transfer of file", tarfile)
		}
		log.Print("  finished sending to hpss ", tarfile)
	}
	hpssWaiter.Done()
}
*/

// archive a directory tree into a number of archives
func archive(archivename string, directory string, maxsize int) {
	start := time.Now()
	// traverse tree
	process := make(chan dirEntry, 10)
	go fileHandler(archivename, int64(maxsize)*1024*1024*1024, process)
	log.Print("scanning files")
	walk(directory, process)
	close(process)
	log.Print("finished scanning")

	log.Print("waiting for tar")
	tarWaiter.Wait()
	log.Print("finished waiting for tar")
	log.Print("time for tar: ", time.Since(start))
	log.Print("BW for tar: ", float64(totalbytes)/(1024.0*1024.0)/time.Since(start).Seconds(), " MB/s")

	close(hpsschannel)
	log.Print("waiting for hpss")
	hpssWaiter.Wait()
	log.Print("finished waiting for hpss")
	log.Print("time for hpss: ", time.Since(firstHpsstransfer))
	log.Print("BW for hpss: ", float64(totalbytes)/(1024.0*1024.0)/time.Since(firstHpsstransfer).Seconds(), " MB/s")
}
