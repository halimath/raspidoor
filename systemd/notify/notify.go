package notify

import (
	"fmt"
	"net"
	"os"

	"github.com/halimath/raspidoor/systemd/logging"
)

type Notification int

const (
	NotificationReady Notification = iota
	NotificationStopping
	NotificationReloading
	NotificationWatchdog
)

func (n Notification) msg() string {
	switch n {
	case NotificationReady:
		return "READY=1"
	case NotificationStopping:
		return "STOPPING=1"
	case NotificationReloading:
		return "RELOADING=1"
	case NotificationWatchdog:
		return "WATCHDOG=1"
	default:
		panic(fmt.Sprintf("unknown notification: %v", n))
	}
}

type Notifier interface {
	Notify(Notification) error
}

func Detect(l logging.Logger) Notifier {
	s, ok := os.LookupEnv("NOTIFY_SOCKET")
	if !ok {
		l.Info("EnvVar NOTIFY_SOCKET not found")
		return &nopNotifier{}
	}

	l.Info("EnvVar NOTIFY_SOCKET found; sending systemd notifications to %s", s)

	return &socketNotifier{
		addr: &net.UnixAddr{
			Name: s,
			Net:  "unixgram",
		},
	}

}

type nopNotifier struct{}

var _ Notifier = &nopNotifier{}

func (*nopNotifier) Notify(Notification) error {
	return nil
}

type socketNotifier struct {
	addr *net.UnixAddr
}

var _ Notifier = &socketNotifier{}

func (s *socketNotifier) Notify(n Notification) error {
	conn, err := net.DialUnix(s.addr.Net, nil, s.addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(n.msg())); err != nil {
		return err
	}

	return nil
}
