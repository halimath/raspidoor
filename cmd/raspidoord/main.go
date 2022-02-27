package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/halimath/raspidoor/internal/config"
	"github.com/halimath/raspidoor/internal/gatekeeper"
	"github.com/halimath/raspidoor/internal/systemd"
)

var (
	Version        = "0.1.0"
	Revision       = "local"
	BuildTimestamp = "0000-00-00T00:00:00"

	configFile = flag.String("config-file", "/etc/raspidoor/raspidoord.yaml", "The config file to read")
)

func main() {
	err := doMain()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func doMain() error {
	flag.Parse()

	c, err := config.ReadConfigFile(*configFile)
	if err != nil {
		return err
	}

	gc, err := c.GatekeeperOptions()
	if err != nil {
		return err
	}

	notifier := systemd.DetectNotifier(gc.Logger)

	gc.Logger.Info("github.com/halimath/raspidoor v%s (%s; %s)", Version, Revision, BuildTimestamp)

	g, err := gatekeeper.New(gc)
	if err != nil {
		return err
	}
	defer g.Close()

	g.Start()
	if err := notifier.Notify(systemd.NotificationReady); err != nil {
		gc.Logger.Err(err)
	}

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChan

	if err := notifier.Notify(systemd.NotificationStopping); err != nil {
		gc.Logger.Err(err)
	}

	return nil
}
