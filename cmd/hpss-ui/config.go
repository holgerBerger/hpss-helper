package main

// this is a copy from hpss-archive !!!

import (
	"github.com/holgerBerger/hpsshelper"
)

type configT struct {
	General GeneralConfigT
	Hpss    hpsshelper.HpssConfigT
}

type GeneralConfigT struct {
	Workdir    string // working directory, like /tmp or .
	Cachedir   string // path to caching dir for archive lists
	Hpssmovers int    // number of pftp_clients running in parallel
	Tars       int    // number of tar processes to run in parallel
}
