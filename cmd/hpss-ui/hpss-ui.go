package main

import (
	//"fmt"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var (
	startdir         string
	activeview       int = 0
	currentdir       string
	currentselection int = 0
	currentfsorigin  int = 0
	fsselection          = make(map[int]bool)
	logview          *gocui.View
)

func setCurrentViewOnTop(g *gocui.Gui, name string) (*gocui.View, error) {
	if _, err := g.SetCurrentView(name); err != nil {
		return nil, err
	}
	return g.SetViewOnTop(name)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if activeview == 0 {
		activeview = 1
		if _, err := setCurrentViewOnTop(g, "hpss"); err != nil {
			return err
		}
	} else {
		activeview = 0
		if _, err := setCurrentViewOnTop(g, "fs"); err != nil {
			return err
		}
	}
	return nil
}

func CursorUp(g *gocui.Gui, v *gocui.View) error {
	if currentselection > 0 {
		currentselection--
	}
	//_, y := v.Size()
	if currentselection < currentfsorigin {
		currentfsorigin--
		v.SetOrigin(0, currentfsorigin)
	}
	v.Clear()
	v.Rewind()
	fillFs(v)
	return nil
}

func CursorDown(g *gocui.Gui, v *gocui.View) error {
	currentselection++
	_, y := v.Size()
	if currentselection >= currentfsorigin+y {
		currentfsorigin++
		v.SetOrigin(0, currentfsorigin)
	}
	v.Clear()
	v.Rewind()
	fillFs(v)
	return nil
}

func Enter(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == "fs" {
		line, _ := v.Line(currentselection)
		newdir := line[2:]
		err := os.Chdir(newdir)
		if err != nil {
			fmt.Fprintln(logview, "can not change directory, permission denied")
			return nil
		}
		currentdir, _ = os.Getwd()
		fsselection = make(map[int]bool)
		currentselection = 0
		v.Clear()
		v.Rewind()
		fillFs(v)
	}
	return nil
}

func Select(g *gocui.Gui, v *gocui.View) error {
	if v.Name() == "fs" {
		line, _ := v.Line(currentselection)
		if line[2:] == ".." {
			// ignore
		} else {
			if fsselection[currentselection] {
				fsselection[currentselection] = false
			} else {
				fsselection[currentselection] = true
			}
			v.Clear()
			v.Rewind()
			fillFs(v)
		}
	}
	return nil
}

func Home(g *gocui.Gui, v *gocui.View) error {
	os.Chdir(startdir)
	currentdir, _ = os.Getwd()
	fsselection = make(map[int]bool)
	currentselection = 0
	v.Clear()
	v.Rewind()
	fillFs(v)
	return nil
}

func Help(g *gocui.Gui, v *gocui.View) error {
	fmt.Fprintln(logview, "cursur up/down to move cursor, <Home> to go to start directory")
	fmt.Fprintln(logview, "<intert> or <space> to toggle a file/directory selection")
	fmt.Fprintln(logview, "<enter> to change directory")
	fmt.Fprintln(logview, "<tab> to change between filesystem and hpss")
	return nil
}

func fillFs(v *gocui.View) {
	// TODO chache ReadDir
	v.Title = "filesystem " + currentdir
	files, err := ioutil.ReadDir(currentdir)
	if err != nil {
		panic(err)
	}

	// cursor
	currentline := 0
	if currentline == currentselection {
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

	for _, f := range files {
		// cursor
		if currentline == currentselection {
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

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("fs", 0, 0, maxX/2-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "filesystem"
		v.Wrap = false
		v.Autoscroll = false

		fillFs(v)

		if _, err = setCurrentViewOnTop(g, "fs"); err != nil {
			return err
		}
	}

	if v, err := g.SetView("hpss", maxX/2, 0, maxX-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "HPSS archive"
		v.Wrap = false
		v.Autoscroll = false
	}

	if v, err := g.SetView("log", 0, maxY-9, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		logview = v
		v.Title = "log"
		v.Wrap = true
		v.Autoscroll = true
		fmt.Fprintln(v, "F1 for help")
	}

	return nil
}

func main() {
	startdir, _ = os.Getwd()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	currentdir, _ = os.Getwd()
	// currentdir += "/.."

	// properties of gui
	g.Highlight = true
	g.Cursor = false
	g.SelFgColor = gocui.ColorRed
	g.SelBgColor = gocui.ColorWhite
	g.BgColor = gocui.ColorWhite
	g.FgColor = gocui.ColorBlack

	// init windows
	g.SetManagerFunc(layout)

	// keyboard bindings
	// quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	// switch view with TAB
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	// cursor up
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	// cursor down
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	// Enter = chdir
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, Enter); err != nil {
		log.Panicln(err)
	}
	// CtrlEnter = select
	if err := g.SetKeybinding("", gocui.KeyInsert, gocui.ModNone, Select); err != nil {
		log.Panicln(err)
	}
	// Space = select
	if err := g.SetKeybinding("", gocui.KeySpace, gocui.ModNone, Select); err != nil {
		log.Panicln(err)
	}
	// Home
	if err := g.SetKeybinding("", gocui.KeyHome, gocui.ModNone, Home); err != nil {
		log.Panicln(err)
	}
	// F1
	if err := g.SetKeybinding("", gocui.KeyF1, gocui.ModNone, Help); err != nil {
		log.Panicln(err)
	}

	// mainloop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
