package main

import (
  "os"
  "log"
  "strings"
  "github.com/holgerBerger/hpsshelper"
)

func refresh() {
  cwd, _ := os.Getwd()
  os.Chdir(config.General.Cachedir)
  pftp := hpsshelper.NewPftp(config.Hpss)
  pftp.Cd(config.Hpss.Hpssbasedir)

  log.Println(" refreshing cache")
  _, out := pftp.Ls("*.full")
  for _,file := range strings.Fields(out) {
    log.Println("  fetching",file)
    pftp.Get(file)
  }
  log.Println(" finished refreshing cache",)

  pftp.Bye()
  os.Chdir(cwd)
}
