package main

import (
	"github.com/fatih/color"
	"os"
)

/*
   Creation Time: 2020 - Jan - 26
   Created by:  (ehsan)
   Maintainers:
      1.  Ehsan N. Moosa (E2)
   Auditor: Ehsan N. Moosa (E2)
   Copyright Ronak Software Group 2018
*/



func PrintOnErr(err error, extra ...interface{}) {
	if err != nil {
		color.Red("PrintOnErr:: %s (%v)", err.Error(), extra)
	}
}

func PanicOnErr(err error, extra ...interface{}) {
	if err != nil {
		color.Red("PrintOnErr:: %s (%v)", err.Error(), extra)
		os.Exit(1)
	}
}