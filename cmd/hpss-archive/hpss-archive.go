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
	Verbose bool `long:"verbose" short:"v" description:"show more output"`
}

var config configT

func main() {
	args, err := flags.Parse(&opts)
	_ = err

	// read config
	if _, err := toml.DecodeFile(os.ExpandEnv("${HOME}/.hpss-archive.conf"), &config); err != nil {
		log.Print("error in reading $HOME/.hpss-archive.conf")
		log.Fatal(err)
	}

	// check if cache dir exists
	if _, err := os.Stat(os.ExpandEnv(config.General.Cachedir)); os.IsNotExist(err) {
		log.Print("Cachdir ", os.ExpandEnv(config.General.Cachedir), " does not exist!")
		os.Exit(1)
	}

	if opts.Verbose {
		log.Println("Cachdir (metadata cache): ", os.ExpandEnv(config.General.Cachedir))
		log.Println("Workdir (place for temporary files): ", os.ExpandEnv(config.General.Workdir))
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
