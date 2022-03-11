package controller

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/halimath/raspidoor/controller"
	"github.com/halimath/raspidoor/daemon/internal/gatekeeper"
	"github.com/halimath/raspidoor/daemon/internal/logging"
	"google.golang.org/grpc"
)

type Controller struct {
	controller.UnimplementedControllerServer

	listener   net.Listener
	server     *grpc.Server
	gatekeeper *gatekeeper.Gatekeeper
	logger     logging.Logger
}

func (c *Controller) SetState(ctx context.Context, msg *controller.EnabledState) (*controller.Result, error) {
	c.logger.Info("Received SetState: %s %d %v", msg.Target.String(), msg.Index, msg.State)

	if msg.Target == controller.Target_BELL_PUSH {
		if err := c.gatekeeper.SetBellPushState(int(msg.Index), msg.State); err != nil {
			return failed(err.Error())
		}
		return ok()
	}

	if msg.Target == controller.Target_EXTERNAL_BELL {
		c.gatekeeper.SetExternalBellState(msg.State)
		return ok()
	}

	if msg.Target == controller.Target_PHONE {
		c.gatekeeper.SetPhoneBellState(msg.State)
		return ok()
	}

	return failed(fmt.Sprintf("unknown target: %d", msg.Target))
}

func (c *Controller) Ring(context.Context, *controller.Empty) (*controller.Empty, error) {
	c.logger.Info("Received Ring")
	c.gatekeeper.Ring()
	return &controller.Empty{}, nil
}

func (c *Controller) Info(ctx context.Context, _ *controller.Empty) (*controller.StateInfo, error) {
	i := c.gatekeeper.Info()

	r := controller.StateInfo{
		BellPushes: make([]*controller.ItemState, len(i.BellPushes)),
		Bells:      make([]*controller.ItemState, len(i.Bells)),
	}

	for idx, p := range i.BellPushes {
		r.BellPushes[idx] = &controller.ItemState{
			Label:   p.Label,
			Enabled: p.Enabled,
		}
	}

	for idx, b := range i.Bells {
		r.Bells[idx] = &controller.ItemState{
			Label:   b.Label,
			Enabled: b.Enabled,
		}
	}

	return &r, nil
}

func ok() (*controller.Result, error) {
	return &controller.Result{
		Ok: true,
	}, nil
}

func failed(e string) (*controller.Result, error) {
	return &controller.Result{
		Ok:    false,
		Error: e,
	}, nil
}

func (c *Controller) Close() {
	c.logger.Info("Closing socket")
	c.server.GracefulStop()
}

func New(g *gatekeeper.Gatekeeper, socket string, logger logging.Logger) (*Controller, error) {
	l, err := net.Listen("unix", socket)
	if err != nil {
		return nil, err
	}

	err = os.Chmod(socket, 0666)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		listener:   l,
		gatekeeper: g,
		logger:     logger,
	}

	s := grpc.NewServer()
	controller.RegisterControllerServer(s, c)

	c.server = s

	go func() {
		if err := s.Serve(l); err != nil {
			logger.Err(err)
		}
	}()

	return c, nil
}
