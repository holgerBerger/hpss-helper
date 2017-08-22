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

type hpssT struct {
	toplevel       bool     // true if on top level = view of archive names
	currentarchive string   // name of current archive
	cursorline     int      // line with cursorline
	origin         int      // top line of displayed
	prefix         string   // commom prefix part of path which is not displayed
	lines          []string // lines of file
	folderlines    []string // content of current folder to display
}

var hpss hpssT

// dar the view frame
func fillHpss(v *gocui.View) {
	if hpss.toplevel { // list of archives
		v.Title = "HPSS archive - available archives"

		archives, err := filepath.Glob(os.ExpandEnv(config.General.Cachedir) + "/*.full")
		if err != nil {
			panic(err)
		} else {
			currentline := 0

			for _, file := range archives {
				if hpss.cursorline == currentline {
					fmt.Fprint(v, "\x1b[0;30;47m>")
				} else {
					fmt.Fprint(v, " ")
				}
				fmt.Fprintln(v, filepath.Base(file)[:len(filepath.Base(file))-5])
				currentline++
			}
		}
	} else { // not toplevel, but in archive
		v.Title = "HPSS archive " + hpss.currentarchive + " " + hpss.prefix

		// ..
		currentline := 0
		if hpss.cursorline == currentline {
			fmt.Fprint(v, "\x1b[0;30;47m>")
		} else {
			fmt.Fprint(v, " ")
		}
		fmt.Fprintln(v, "\x1b[0;32m..")
		currentline++
		// entries
		for _, i := range hpss.folderlines {
			if hpss.cursorline == currentline {
				fmt.Fprint(v, "\x1b[0;30;47m>")
			} else {
				fmt.Fprint(v, " ")
			}
			if strings.Contains(i, "|") {
				fmt.Fprintln(hpssview, "\x1b[0;30;47m"+i)
			} else {
				fmt.Fprintln(hpssview, "\x1b[0;32m"+i)
			}
			currentline++
		}

	}
}

// move cursor up and scroll if necessary
func hpssCursorUp(g *gocui.Gui, v *gocui.View) error {
	if hpss.cursorline > 0 {
		hpss.cursorline--
	}
	//_, y := v.Size()
	if hpss.cursorline < hpss.origin {
		hpss.origin--
		v.SetOrigin(0, hpss.origin)
	}
	v.Clear()
	v.Rewind()
	fillHpss(v)
	return nil
}

// move cursur down and scroll if necessary
func hpssCursorDown(g *gocui.Gui, v *gocui.View) error {
	hpss.cursorline++
	newline, err := v.Line(hpss.cursorline)
	if err != nil || len(newline) == 0 {
		hpss.cursorline--
	}

	_, y := v.Size()
	if hpss.cursorline >= hpss.origin+y {
		hpss.cursorline++
		v.SetOrigin(0, hpss.origin)
	}
	v.Clear()
	v.Rewind()
	fillHpss(v)
	return nil
}

// read archive cache file into hpss.lines and fill current folder data
func readArchiveCache() {
	file, err := ioutil.ReadFile(config.General.Cachedir + "/" + hpss.currentarchive + ".full")
	if err != nil {
		fmt.Fprintln(logview, "could not read archive cache!")
		return
	}
	hpss.lines = strings.Split(string(file), "\n")
	hpss.prefix = ""
	getFolderContens()
}

// filter current folder from hpss.lines by showing lines with common prefix
func getFolderContens() {
	shown := make(map[string]bool)
	hpss.folderlines = make([]string, 0, 0)

	for _, line := range hpss.lines {
		if (len(hpss.prefix) <= len(line)) && (line[:len(hpss.prefix)] == hpss.prefix) {
			dir := line[len(hpss.prefix):] // prefix removed
			entry := strings.Split(dir, "/")[0]
			if _, ok := shown[entry]; ok {
				// already displayed, skip
			} else {
				shown[entry] = true
				hpss.folderlines = append(hpss.folderlines, entry)
			}
		}
	}
}

// chamge mode if on top level or change directory
func hpssEnter(g *gocui.Gui, v *gocui.View) error {
	if hpss.toplevel {
		hpss.toplevel = false
		line, _ := v.Line(hpss.cursorline)
		hpss.currentarchive = line[1:]
		hpss.cursorline = 0
		readArchiveCache()
		v.Clear()
		v.Rewind()
		fillHpss(v)
	} else {
		line, _ := v.Line(hpss.cursorline)
		fmt.Fprintln(logview, "selected "+line)
		if line[1:] == ".." { // going up
			fmt.Fprintln(logview, "to split: "+hpss.prefix)
			splitted := strings.Split(hpss.prefix, "/")
			fmt.Fprintln(logview, splitted, len(splitted))
			if len(splitted) <= 1 { // top level
				hpss.toplevel = true
				hpss.prefix = ""
			} else if len(splitted) <= 2 { // one below toplevel
				hpss.prefix = ""
			} else { // deeper levels
				hpss.prefix = strings.Join(splitted[:len(splitted)-2], "/") + "/"
			}
		} else { // going down
			hpss.prefix = hpss.prefix + line[1:] + "/"
		}
		fmt.Fprintln(logview, "prefix: "+hpss.prefix)
		hpss.cursorline = 0
		hpss.origin = 0
		getFolderContens()
		v.Clear()
		v.Rewind()
		fillHpss(v)
	}
	return nil
}

// create key bindings for fs viewer
func hpssviewkeybindings(g *gocui.Gui) {
	hpss.toplevel = true
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
