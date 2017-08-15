package main

import (
	"fmt"
  "path/filepath"
  "os"
	"github.com/jroimartin/gocui"
)

var (
	hpsstoplevel       bool = true
	hpsscurrentarchive string
)

func fillHpss(v *gocui.View) {
	if hpsstoplevel {
		v.Title = "HPSS archive - available archives"
		archives, err := filepath.Glob(os.ExpandEnv(config.General.Cachedir)+"/*.full")
		if err != nil {
      panic(err)
		} else {
			for _, file := range archives {
				fmt.Fprintln(v, filepath.Base(file)[:len(filepath.Base(file))-5])
			}
		}
	} else {
		v.Title = "HPSS archive -" + hpsscurrentarchive
	}
}
