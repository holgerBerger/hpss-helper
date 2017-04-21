package main

type configT struct {
	Workdir      string // working directory, like /tmp or .
	Cachedir     string // path to caching dir for archive lists
	Tars         int    // number of tar processes to run in parallel
	Hpssmovers   int    // number of pftp_clients running in parallel
	Hpssserver   string // name of hpss server
	Hpssport     int    // TCP port of hpss server
	Hpssusername string // username
	Hpsspassword string // password !! DAMN HPSS
	Hpssbasedir  string // prefix to path in HPSS
	Pftp_client  string // path to pftp_client
}
