package controller

import (
	"context"
	"net"

	"github.com/halimath/raspidoor/controller"
	"github.com/halimath/raspidoor/internal/gatekeeper"
	"github.com/halimath/raspidoor/logging"
	"google.golang.org/grpc"
)

type Controller struct {
	controller.UnimplementedControllerServer

	listener   net.Listener
	server     *grpc.Server
	gatekeeper *gatekeeper.Gatekeeper
	logger     logging.Logger
}

func (c *Controller) SetBellPushState(ctx context.Context, msg *controller.BellPushState) (*controller.Result, error) {
	c.logger.Info("Received SetBellPushState: %d %v", msg.Index, msg.State)

	if err := c.gatekeeper.SetBellPushState(int(msg.Index), msg.State); err != nil {
		return failed(err), nil
	}

	return ok(), nil
}

func ok() *controller.Result {
	return &controller.Result{
		Ok: true,
	}
}

func failed(err error) *controller.Result {
	return &controller.Result{
		Ok:    false,
		Error: err.Error(),
	}
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
