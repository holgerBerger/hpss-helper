package main

/* simple test programm for pftp interface,
   pushes files into hpssbasedir
*/

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/holgerBerger/hpsshelper"
	"log"
	"os"
	"path"
)

var config configT

func main() {
	if _, err := toml.DecodeFile(os.ExpandEnv("${HOME}/.hpsshelper.conf"), &config); err != nil {
		log.Print("error in reading $HOME/.hpsshelper.conf")
		log.Fatal(err)
	}

	pftp := hpsshelper.NewPftp(config.Hpss)
	if pftp == nil {
		log.Fatal("error in init")
	}

	err := pftp.Cd(config.Hpss.Hpssbasedir)
	if err != nil {
		log.Fatal("error in cd")
	}

	for _, file := range os.Args[1:] {
		err := pftp.Put(file, path.Base(file))
		if err != nil {
			log.Fatal("error in put for ", file)
		}
	}

	pftp.Bye()

	// print log of transfer
	fmt.Println(pftp.Protocoll.String())
}
