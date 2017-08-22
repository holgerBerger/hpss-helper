package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/nsf/termbox-go"
)

// move cursor up and scroll if necessary
func fsCursorUp(g *gocui.Gui, v *gocui.View) error {
	if currentfsselection > 0 {
		currentfsselection--
	}
	//_, y := v.Size()
	if currentfsselection < currentfsorigin {
		currentfsorigin--
		v.SetOrigin(0, currentfsorigin)
	}
	v.Clear()
	v.Rewind()
	fillFs(v)
	return nil
}

// move cursur down and scroll if necessary
func fsCursorDown(g *gocui.Gui, v *gocui.View) error {
	currentfsselection++
	newline, err := v.Line(currentfsselection)
	if err != nil || len(newline) == 0 {
		currentfsselection--
	}

	_, y := v.Size()
	if currentfsselection >= currentfsorigin+y {
		currentfsorigin++
		v.SetOrigin(0, currentfsorigin)
	}
	v.Clear()
	v.Rewind()
	fillFs(v)
	return nil
}

// change directory
func fsEnter(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == "fs" {
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
	}
	return nil
}

// select or deselect entry
func fsSelect(g *gocui.Gui, v *gocui.View) error {
	// FIXME if looks unnecessary
	if v.Name() == "fs" {
		line, _ := v.Line(currentfsselection)
		if line[2:] == ".." {
			// ignore
		} else {
			if fsselection[currentfsselection] {
				fsselection[currentfsselection] = false
			} else {
				fsselection[currentfsselection] = true
			}
			v.Clear()
			v.Rewind()
			fillFs(v)
		}
	}
	return nil
}

// change to directory in which the hpss-ui was started
func fsHome(g *gocui.Gui, v *gocui.View) error {
	os.Chdir(startdir)
	currentdir, _ = os.Getwd()
	fsselection = make(map[int]bool)
	currentfsselection = 0
	currentfiles = make([]os.FileInfo, 0, 0)
	v.Clear()
	v.Rewind()
	fillFs(v)
	return nil
}

// create key bindings for fs viewer
func fsviewkeybindings(g *gocui.Gui) {
	// cursor up
	if err := g.SetKeybinding("fs", gocui.KeyArrowUp, gocui.ModNone, fsCursorUp); err != nil {
		log.Panicln(err)
	}
	// cursor down
	if err := g.SetKeybinding("fs", gocui.KeyArrowDown, gocui.ModNone, fsCursorDown); err != nil {
		log.Panicln(err)
	}
	// Enter = chdir
	if err := g.SetKeybinding("fs", gocui.KeyEnter, gocui.ModNone, fsEnter); err != nil {
		log.Panicln(err)
	}
	// CtrlEnter = select
	if err := g.SetKeybinding("fs", gocui.KeyInsert, gocui.ModNone, fsSelect); err != nil {
		log.Panicln(err)
	}
	// Space = select
	if err := g.SetKeybinding("fs", gocui.KeySpace, gocui.ModNone, fsSelect); err != nil {
		log.Panicln(err)
	}
	// Home
	if err := g.SetKeybinding("fs", gocui.KeyHome, gocui.ModNone, fsHome); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlL, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return termbox.Sync()
		}); err != nil {
		// TODO: handle the error
	}

}

// populate fs view
func fillFs(v *gocui.View) {

	v.Title = "filesystem " + currentdir

	if len(currentfiles) == 0 {
		var err error
		currentfiles, err = ioutil.ReadDir(currentdir)
		if err != nil {
			panic(err)
		}
	}

	// cursor
	currentline := 0
	if currentline == currentfsselection {
		fmt.Fprint(v, "\x1b[0;30;47m>")
	} else {
		fmt.Fprint(v, " ")
	}
	// selection marker
	if fsselection[currentline] {
		fmt.Fprint(v, "\x1b[0;33;47m+")
	} else {
		fmt.Fprint(v, " ")
	}
	// entry .. as dir
	fmt.Fprintln(v, "\x1b[0;32m..")
	currentline++

	for _, f := range currentfiles {
		// cursor
		if currentline == currentfsselection {
			fmt.Fprint(v, "\x1b[0;30;47m>")
		} else {
			fmt.Fprint(v, " ")
		}
		// selection marker
		if fsselection[currentline] {
			fmt.Fprint(v, "\x1b[0;34;47m+")
		} else {
			fmt.Fprint(v, " ")
		}
		// entry
		if f.IsDir() {
			fmt.Fprintln(v, "\x1b[0;32m"+f.Name())
		} else {
			fmt.Fprintln(v, "\x1b[0;30;47m"+f.Name())
		}
		currentline++
	}
}
