package main

import (
	"github.com/BurntSushi/toml"
	flags "github.com/jessevdk/go-flags"
	"log"
	"os"
)

var opts struct {
	Archive bool `long:"archive" short:"a" description:"archive directory"`
	Maxsize int  `long:"maxsize" short:"s" default:"1" description:"maximum size of fragment in HPSS in GB"`
}

var config configT

func main() {
	args, err := flags.Parse(&opts)
	_ = err

	if _, err := toml.DecodeFile(os.ExpandEnv("${HOME}/.hpps-archive.conf"), &config); err != nil {
		log.Print("error in reading ${HOME}/.hpps-archive.conf")
		log.Fatal(err)
	}

	if opts.Archive {
		if len(args) < 2 {
			log.Print("not enough arguments!")
			os.Exit(1)
		}
		archivename := args[0]
		directory := args[1]
		archive(archivename, directory, opts.Maxsize)
	}

}
