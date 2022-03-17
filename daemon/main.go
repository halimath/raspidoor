package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/halimath/raspidoor/daemon/internal/config"
	"github.com/halimath/raspidoor/daemon/internal/controller"
	"github.com/halimath/raspidoor/daemon/internal/gatekeeper"
	"github.com/halimath/raspidoor/systemd/notify"
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

	logger, err := c.NewLogger()
	if err != nil {
		return err
	}

	logger.Info("raspidoord/%s (+github.com/halimath/raspidoor, built %s, git %s)", Version, BuildTimestamp, Revision)

	gc, err := c.GatekeeperOptions()
	if err != nil {
		return err
	}

	notifier := notify.Detect(logger)

	g, err := gatekeeper.New(gc, logger)
	if err != nil {
		return err
	}
	defer g.Close()

	ctrl, err := controller.New(g, c.Controller.Socket, logger)
	if err != nil {
		return err
	}
	defer ctrl.Close()

	g.Start()
	if err := notifier.Notify(notify.NotificationReady); err != nil {
		logger.Err(err)
	}

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChan

	if err := notifier.Notify(notify.NotificationStopping); err != nil {
		logger.Err(err)
	}

	return nil
}
