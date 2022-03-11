package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/halimath/raspidoor/controller"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var (
	Version        = "0.1.0"
	Revision       = "local"
	BuildTimestamp = "0000-00-00T00:00:00"

	socket  *string
	rootCmd *cobra.Command
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "raspidoor",
		Short: "CLI controller for raspidoord",
		Long: `A command line interface for controlling raspidoord.
	See https://github.com/halimath/raspidoor for details.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(os.Stderr, "Missing command\n")
			cmd.Usage()
			os.Exit(1)
		},
	}

	socket = rootCmd.PersistentFlags().StringP("socket", "s", "/var/run/raspidoor.socket", "Path of the unix socket to connect to the daemon")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Show version of raspidoor",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("raspidoor cli v%s built %s (%s)\n", Version, BuildTimestamp, Revision)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "bellpush",
		Short: "Enable/Disable a bell push",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			idx, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Invalid bell push index: %s: %s\n", os.Args[0], args[0], err)
				os.Exit(3)
			}

			state, err := parseEnabled(args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Invalid state: %s: %s\n", os.Args[0], args[1], err)
				os.Exit(3)
			}

			doWithController(func(ctx context.Context, ctrl controller.ControllerClient) error {
				r, err := ctrl.SetState(ctx, &controller.EnabledState{
					Target: controller.Target_BELL_PUSH,
					State:  state,
					Index:  int32(idx),
				})
				if err != nil {
					return err
				}

				if !r.Ok {
					fmt.Fprintf(os.Stderr, "%s: Failed to set bell push state: %s\n", os.Args[0], r.Error)
				}

				return nil
			})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "externalbell",
		Short: "Enable/Disable the external bell",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			state, err := parseEnabled(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Invalid state: %s: %s\n", os.Args[0], args[0], err)
				os.Exit(3)
			}

			doWithController(func(ctx context.Context, ctrl controller.ControllerClient) error {
				r, err := ctrl.SetState(ctx, &controller.EnabledState{
					Target: controller.Target_EXTERNAL_BELL,
					State:  state,
				})

				if err != nil {
					return err
				}

				if !r.Ok {
					fmt.Fprintf(os.Stderr, "%s: Failed to set external bell state: %s\n", os.Args[0], r.Error)
				}

				return nil
			})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "phone",
		Short: "Enable/Disable the phone bell",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			state, err := parseEnabled(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Invalid state: %s: %s\n", os.Args[0], args[0], err)
				os.Exit(3)
			}

			doWithController(func(ctx context.Context, ctrl controller.ControllerClient) error {
				r, err := ctrl.SetState(ctx, &controller.EnabledState{
					Target: controller.Target_PHONE,
					State:  state,
				})

				if err != nil {
					return err
				}

				if !r.Ok {
					fmt.Fprintf(os.Stderr, "%s: Failed to set phone state: %s\n", os.Args[0], r.Error)
				}

				return nil
			})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "ring",
		Short: "Ring to test output",
		Run: func(cmd *cobra.Command, args []string) {
			doWithController(func(ctx context.Context, ctrl controller.ControllerClient) error {
				_, err := ctrl.Ring(ctx, &controller.Empty{})
				return err
			})
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "info",
		Short: "Display info on the current sate",
		Run: func(cmd *cobra.Command, args []string) {
			doWithController(func(ctx context.Context, ctrl controller.ControllerClient) error {
				i, err := ctrl.Info(ctx, &controller.Empty{})

				if err != nil {
					return err
				}

				fmt.Printf("Bell Pushes\n")
				for _, p := range i.BellPushes {
					fmt.Printf("\t%20s: %s\n", p.Label, formatEnabled(p.Enabled))
				}

				fmt.Printf("\nBells\n")
				for _, b := range i.Bells {
					fmt.Printf("\t%20s: %s\n", b.Label, formatEnabled(b.Enabled))
				}

				return nil
			})
		},
	})
}

const (
	enabledValues  = "on true active enabled 1"
	enabledValue   = "on"
	disabledValues = "off false inactive disabled 0"
	disabledValue  = "off"
)

func parseEnabled(s string) (bool, error) {
	if strings.Contains(enabledValues, strings.ToLower(strings.TrimSpace(s))) {
		return true, nil
	}

	if strings.Contains(disabledValues, strings.ToLower(strings.TrimSpace(s))) {
		return false, nil
	}

	return false, fmt.Errorf("invalid state string: %s", s)
}

func formatEnabled(b bool) string {
	if b {
		return enabledValue
	}

	return disabledValue
}

type withController func(context.Context, controller.ControllerClient) error

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

	err = cb(context.Background(), ctrl)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: Failed to send command: %s\n", os.Args[0], err)
		os.Exit(3)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(2)
	}
}
