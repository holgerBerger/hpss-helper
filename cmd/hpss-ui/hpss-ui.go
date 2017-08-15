package main

import (
	//"fmt"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jroimartin/gocui"
)

var config configT

var (
	startdir           string
	activeview         int
	currentdir         string
	currentfsselection int
	currentfsorigin    int
	fsselection        = make(map[int]bool)
	currentfiles       []os.FileInfo
	fsview             *gocui.View
	hpssview           *gocui.View
	logview            *gocui.View
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

// switch between views on <TAB>
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

// print some help to log view
func help(g *gocui.Gui, v *gocui.View) error {
	fmt.Fprintln(logview, "cursur up/down to move cursor, <Home> to go to start directory")
	fmt.Fprintln(logview, "<intert> or <space> to toggle a file/directory selection")
	fmt.Fprintln(logview, "<enter> to change directory")
	fmt.Fprint(logview, "<tab> to change between filesystem and hpss, ")
	fmt.Fprintln(logview, "<ctrl-c> to exit")
	fmt.Fprintln(logview, "F2 archive marked files/directories")
	fmt.Fprintln(logview, "F3 retrieve marked files/directories")
	return nil
}

// move files to archive
func archive(g *gocui.Gui, v *gocui.View) error {
	fmt.Fprint(logview, "archive files: ")
	for i, file := range currentfiles {
		if fsselection[i+1] {
			fmt.Fprint(logview, file.Name()+" ")
		}
	}
	fmt.Fprintln(logview, "")
	return nil
}

// create the views
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("fs", 0, 0, maxX/2-1, maxY-10); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fsview = v
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
		hpssview = v
		v.Title = "HPSS archive"
		v.Wrap = false
		v.Autoscroll = false

		fillHpss(v)
	}

	if v, err := g.SetView("log", 0, maxY-9, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		logview = v
		v.Title = "log"
		v.Wrap = true
		v.Autoscroll = true
		// v.Editable = true
		fmt.Fprintln(v, "F1 for help")
	}

	return nil
}

func main() {
	startdir, _ = os.Getwd()

	// read config
	if _, err := toml.DecodeFile(os.ExpandEnv("${HOME}/.hpsshelper.conf"), &config); err != nil {
		log.Print("error in reading $HOME/.hpsshelper.conf")
		log.Fatal(err)
	}

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

	// keyboard bindings for all views
	// quit
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	// switch view with TAB
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		log.Panicln(err)
	}
	// F1
	if err := g.SetKeybinding("", gocui.KeyF1, gocui.ModNone, help); err != nil {
		log.Panicln(err)
	}
	// F2
	if err := g.SetKeybinding("", gocui.KeyF2, gocui.ModNone, archive); err != nil {
		log.Panicln(err)
	}

	fsviewkeybindings(g)

	// mainloop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
