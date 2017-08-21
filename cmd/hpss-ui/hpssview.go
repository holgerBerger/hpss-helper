package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	hpsstoplevel         bool = true
	hpsscurrentarchive   string
	currenthpssselection int
	currenthpssorigin    int
	hpssviewprefix       string
	hpsslines            []string // file as lines
	hpssfolderlines      []string // contents of current folder
)

func fillHpss(v *gocui.View) {
	if hpsstoplevel {
		v.Title = "HPSS archive - available archives"
		archives, err := filepath.Glob(os.ExpandEnv(config.General.Cachedir) + "/*.full")
		if err != nil {
			panic(err)
		} else {
			currentline := 0

			for _, file := range archives {
				if currenthpssselection == currentline {
					fmt.Fprint(v, "\x1b[0;30;47m>")
				} else {
					fmt.Fprint(v, " ")
				}
				fmt.Fprintln(v, filepath.Base(file)[:len(filepath.Base(file))-5])
				currentline++
			}
		}
	} else {
		v.Title = "HPSS archive -" + hpsscurrentarchive

		currentline := 0
		for _, i := range hpssfolderlines {
			if currenthpssselection == currentline {
				fmt.Fprint(v, "\x1b[0;30;47m>")
			} else {
				fmt.Fprint(v, " ")
			}
			fmt.Fprintln(hpssview, i)
			currentline++
		}

	}
}

// move cursor up and scroll if necessary
func hpssCursorUp(g *gocui.Gui, v *gocui.View) error {
	if currenthpssselection > 0 {
		currenthpssselection--
	}
	//_, y := v.Size()
	if currenthpssselection < currenthpssorigin {
		currenthpssorigin--
		v.SetOrigin(0, currenthpssorigin)
	}
	v.Clear()
	v.Rewind()
	fillHpss(v)
	return nil
}

// move cursur down and scroll if necessary
func hpssCursorDown(g *gocui.Gui, v *gocui.View) error {
	currenthpssselection++
	newline, err := v.Line(currenthpssselection)
	if err != nil || len(newline) == 0 {
		currenthpssselection--
	}

	_, y := v.Size()
	if currenthpssselection >= currenthpssorigin+y {
		currenthpssorigin++
		v.SetOrigin(0, currenthpssorigin)
	}
	v.Clear()
	v.Rewind()
	fillHpss(v)
	return nil
}

func readArchiveCache() {
	file, err := ioutil.ReadFile(config.General.Cachedir + "/" + hpsscurrentarchive + ".full")
	if err != nil {
		fmt.Fprintln(logview, "could not read archive cache!")
		return
	}
	hpsslines = strings.Split(string(file), "\n")
	hpssviewprefix = ""
	getFolderContens()
}

func getFolderContens() {
	shown := make(map[string]bool)
  hpssfolderlines=make([]string, 0,0 )

	for _, line := range hpsslines {
		if (len(hpssviewprefix)<=len(line)) && (line[:len(hpssviewprefix)] == hpssviewprefix) {
			dir := line[len(hpssviewprefix):] // prefix removed
			entry := strings.Split(dir, "/")[0]
			if _, ok := shown[entry]; ok {
				// already displayed, skip
			} else {
				shown[entry] = true
				hpssfolderlines=append(hpssfolderlines, entry)
			}
		}
	}
}

// chamge mode if on top level or change directory
func hpssEnter(g *gocui.Gui, v *gocui.View) error {
	if hpsstoplevel {
		hpsstoplevel = false
		line, _ := v.Line(currenthpssselection)
		hpsscurrentarchive = line[1:]
		currenthpssselection = 0
		readArchiveCache()
		v.Clear()
		v.Rewind()
		fillHpss(v)
	} else {
		line, _ := v.Line(currenthpssselection)
		fmt.Fprintln(logview, "selected "+line )
		hpssviewprefix = hpssviewprefix + line[1:]+"/"
		fmt.Fprintln(logview, hpssviewprefix)
		currenthpssselection = 0
		currenthpssorigin = 0
		getFolderContens()
		v.Clear()
		v.Rewind()
		fillHpss(v)

		/*
			line, _ := v.Line(currentfsselection)
			newdir := line[2:]
			err := os.Chdir(newdir)
			if err != nil {
				fmt.Fprintln(logview, err)
				return nil
			}
			currentdir, _ = os.Getwd()
			fsselection = make(map[int]bool)
			currentfsselection = 0
			currentfiles = make([]os.FileInfo, 0, 0)
			v.Clear()
			v.Rewind()
			fillFs(v)
		*/
	}
	return nil
}

// create key bindings for fs viewer
func hpssviewkeybindings(g *gocui.Gui) {
	// cursor up
	if err := g.SetKeybinding("hpss", gocui.KeyArrowUp, gocui.ModNone, hpssCursorUp); err != nil {
		log.Panicln(err)
	}
	// cursor down
	if err := g.SetKeybinding("hpss", gocui.KeyArrowDown, gocui.ModNone, hpssCursorDown); err != nil {
		log.Panicln(err)
	}

	// Enter = chdir
	if err := g.SetKeybinding("hpss", gocui.KeyEnter, gocui.ModNone, hpssEnter); err != nil {
		log.Panicln(err)
	}
	/*
		// CtrlEnter = select
		if err := g.SetKeybinding("fs", gocui.KeyInsert, gocui.ModNone, hpssSelect); err != nil {
			log.Panicln(err)
		}
		// Space = select
		if err := g.SetKeybinding("fs", gocui.KeySpace, gocui.ModNone, hpssSelect); err != nil {
			log.Panicln(err)
		}
		// Home
		if err := g.SetKeybinding("fs", gocui.KeyHome, gocui.ModNone, hpssHome); err != nil {
			log.Panicln(err)
		}
	*/
}
