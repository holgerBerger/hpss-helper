package hpsshelper

type Pftp struct {
	hpssconfig HpssConfigT
}

func NewPftp(hpssconfig HpssConfigT) *Pftp {
	pftp := new(Pftp)
	pftp.hpssconfig = hpssconfig
	return pftp
}
