package hpsshelper

type HpssConfigT struct {
	Pftp_client  string // path to pftp_client
	Hpssbasedir  string // prefix to path in HPSS
	Hpssserver   string // name of hpss server
	Hpssport     int    // TCP port of hpss server
	Hpssusername string // username
	Hpsspassword string // password !! DAMN HPSS
	Hpsswidth    int    // numer of parallel conncections, -wX
}
