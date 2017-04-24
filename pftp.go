package hpsshelper

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Pftp struct {
	hpssconfig HpssConfigT
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	bstdout    *bufio.Reader
	Protocoll  bytes.Buffer
}

func NewPftp(hpssconfig HpssConfigT) *Pftp {
	var err error
	pftp := new(Pftp)
	pftp.hpssconfig = hpssconfig

	if pftp.hpssconfig.Hpsswidth == 0 {
		pftp.hpssconfig.Hpsswidth = 1
	}
	pftp.cmd = exec.Command(pftp.hpssconfig.Pftp_client,
		"-w"+strconv.Itoa(pftp.hpssconfig.Hpsswidth), "-inv",
		pftp.hpssconfig.Hpssserver, strconv.Itoa(pftp.hpssconfig.Hpssport))
	pftp.stdin, err = pftp.cmd.StdinPipe()
	if err != nil {
		log.Fatal("error calling pftp_client", err)
	}
	pftp.stdout, err = pftp.cmd.StdoutPipe()
	pftp.bstdout = bufio.NewReader(pftp.stdout)

	pftp.cmd.Start()

	err = pftp.sendcmd("quote USER", pftp.hpssconfig.Hpssusername, 10*time.Second)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	err = pftp.sendcmd("quote pass", pftp.hpssconfig.Hpsspassword, 10*time.Second)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return pftp
}

func (p *Pftp) Bye() error {
	return p.sendcmd("bye", "", 1*time.Second)
}

func (p *Pftp) Cd(dir string) error {
	return p.sendcmd("cd", dir, 1*time.Second)
}

func (p *Pftp) Put(src string, tgt string) error {
	return p.sendcmd("put", src+" "+tgt, 1000*time.Second)
}

func (p *Pftp) Get(src string) error {
	return p.sendcmd("get", src, 1000*time.Second)
}

// sendcmd sends a command to pftp and waits for a return code
// as we are in verbose mode, we get <XYZ > codes back after commands,
// 1YZ, 2YZ and 3YZ is good, 4YZ and 5YZ is bad.
// 2YZ is completion
// list of important messages:
//   220 before login
//   215 information after login
//   230 logged in
//   150 dir output start
//   226 transfer complete
//   250 CWD success
//   550 no such file
func (p *Pftp) sendcmd(cmd string, args string, timeout time.Duration) error {

	io.WriteString(p.stdin, cmd+" "+args+"\n")

	phase2 := false

	ch := make(chan error)
	go func() {
		for {
			time.Sleep(10 * time.Millisecond)
			out, _ := p.bstdout.ReadBytes('\n')

			str := strings.Replace(string(out), p.hpssconfig.Hpsspassword, "***", -1)
			// log.Println(str) // DEBUG
			p.Protocoll.WriteString(str)

			if out == nil {
				log.Fatal("error in hpss communication!")
				ch <- errors.New("error in communication")
				break
			}
			if cmd == "cd" {
				if strings.Index(string(out), "250 CWD") != -1 {
					ch <- nil
					break
				}
				if strings.Index(string(out), "550 ") != -1 {
					ch <- errors.New("No such file or directory")
					break
				}
			}
			if cmd == "quote USER" {
				if strings.Index(string(out), "331 P") != -1 {
					ch <- nil
					break
				}
				if strings.Index(string(out), "Not connected") != -1 {
					ch <- errors.New("Not connected")
					break
				}
			}
			if cmd == "quote pass" {
				if strings.Index(string(out), "230 U") != -1 {
					ch <- nil
					break
				}
				if strings.Index(string(out), "Not connected") != -1 {
					ch <- errors.New("Not connected")
					break
				}
				if strings.Index(string(out), "503 ") != -1 {
					ch <- errors.New("Login with USER first")
					break
				}
				if strings.Index(string(out), "530 ") != -1 {
					ch <- errors.New("Login incorrect")
					break
				}
			}
			if cmd == "put" {
				if strings.Index(string(out), "bytes sent in ") != -1 {
					log.Print("  ", string(out))
					ch <- nil
					break
				}
				if strings.Index(string(out), "226 ") != -1 {
					//ch <- nil
					//break
				}
				if strings.Index(string(out), "559 ") != -1 {
					ch <- errors.New("permission problem")
					break
				}
			}
			if cmd == "get" {
				if strings.Index(string(out), "226 ") != -1 {
					// get first gives 226 and later 200
					phase2 = true
				}
				if phase2 && strings.Index(string(out), "200 ") != -1 {
					// log.Println("got 200")
					ch <- nil
					break
				}
				if strings.Index(string(out), " bytes received ") != -1 {
					// sometimes there is first 200, then 226, and afterwards this message
					ch <- nil
					break
				}
				if strings.Index(string(out), "559 ") != -1 {
					ch <- errors.New("permission problem")
					break
				}
			}
			if cmd == "bye" {
				if strings.Index(string(out), "221 ") != -1 {
					ch <- nil
					break
				}
			}
		}
		ch <- nil
	}()
	select {
	case <-ch:
		return <-ch
	case <-time.After(timeout):
		log.Print(p.Protocoll.String())
		log.Fatal("timeout in hpss communication")
		return errors.New("timeout in hpss")
		//os.Exit(1)
	}
	return nil
}
