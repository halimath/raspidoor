package controller

import (
	"context"
	"log"
	"net"

	"github.com/halimath/raspidoor/controller"
	"github.com/halimath/raspidoor/internal/gatekeeper"
	"google.golang.org/grpc"
)

type Controller struct {
	controller.UnimplementedControllerServer

	listener   net.Listener
	server     *grpc.Server
	gatekeeper *gatekeeper.Gatekeeper
}

func (c *Controller) SetBellSwitch(context.Context, *controller.BellSwitchMessage) (*controller.Result, error) {
	return &controller.Result{Ok: false, Error: "not implemented"}, nil
}

func (c *Controller) Close() {
	c.server.GracefulStop()
}

func New(g *gatekeeper.Gatekeeper, socket string) (*Controller, error) {
	l, err := net.Listen("unixpacket", socket)
	if err != nil {
		return nil, err
	}

	c := &Controller{
		listener:   l,
		gatekeeper: g,
	}

	s := grpc.NewServer()
	controller.RegisterControllerServer(s, c)

	if err := s.Serve(l); err != nil {
		log.Fatal(err)
	}

	return c, nil
}
