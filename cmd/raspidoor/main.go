package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/halimath/raspidoor/controller"
	"google.golang.org/grpc"
)

var (
	socket = flag.String("socket", "/var/run/raspidoor.socket", "Path of the unix socket to connect to the daemon")
)

type withController func(context.Context, controller.ControllerClient) (*controller.Result, error)

func doWithController(cb withController) {
	dialer := func(addr string, t time.Duration) (net.Conn, error) {
		return net.Dial("unix", addr)
	}

	con, err := grpc.Dial(*socket, grpc.WithInsecure(), grpc.WithDialer(dialer))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Failed to open socket %s: %s\n", os.Args[0], *socket, err)
		os.Exit(2)
	}
	defer con.Close()

	ctrl := controller.NewControllerClient(con)

	res, err := cb(context.Background(), ctrl)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Failed to send command: %s\n", os.Args[0], err)
		os.Exit(3)
	}

	if !res.Ok {
		fmt.Fprintf(os.Stderr, "%s: Got error from raspidoord: %s\n", os.Args[0], res.Error)
		os.Exit(4)
	}
}

func main() {
	flag.Parse()

	switch flag.Arg(0) {
	case "":
		fmt.Fprintf(os.Stderr, "%s: Missing command\n", os.Args[0])
		os.Exit(1)

	case "bellpush":
		if flag.NArg() != 3 {
			fmt.Fprintf(os.Stderr, "%s: Invalid arguments for %s; must be exactly 2", os.Args[0], flag.Arg(0))
			os.Exit(1)
		}

		idx, err := strconv.Atoi(flag.Arg(1))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: Invalid index arguments: %s", os.Args[0], err)
			os.Exit(1)
		}

		doWithController(func(ctx context.Context, ctrl controller.ControllerClient) (*controller.Result, error) {
			return ctrl.SetBellPushState(ctx, &controller.BellPushState{
				Index: int32(idx),
				State: flag.Arg(2) == "true",
			})
		})
	default:
		fmt.Fprintf(os.Stderr, "%s: Unknown command: %s\n", os.Args[0], flag.Arg(0))
		os.Exit(1)
	}
}
