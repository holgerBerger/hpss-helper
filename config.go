package main

type configT struct {
	Workdir      string // working directory, like /tmp or .
	Cachedir     string // path to caching dir for archive lists
	Hpssserver   string // name of hpss server
	Hpssport     int    // TCP port of hpss server
	Hpssusername string // username
	Pftp_client  string // path to pftp_client
}
