package main

import "github.com/halimath/raspidoor/daemon/internal/sip"

func main() {
	caller, err := sip.ParseURI("sip:klingel@fritz.box")
	if err != nil {
		panic(err)
	}
	callee, err := sip.ParseURI("sip:**611@fritz.box")
	if err != nil {
		panic(err)
	}

	t := sip.NewTCPTransport()
	t.DumpRoundTrips = true

	d := sip.NewDialog(t, caller, sip.NewDigestHandler("klingel1", "password001"))

	d.Ring(callee)
}
