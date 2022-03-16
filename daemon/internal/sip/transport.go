package sip

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrRoundTripFailed = errors.New("round trip failed")
)

type (
	Connection interface {
		Send(*Request) error
		Recv() (*Response, error)
		Close() error
	}

	Transport interface {
		Send(*Request) (Connection, error)
	}
)

type TCPTransport struct {
	DumpRoundTrips bool
}

var _ Transport = &TCPTransport{}

func (t *TCPTransport) Send(req *Request) (Connection, error) {
	addr := fmt.Sprintf("%s:%d", req.URI.Host, req.URI.Port)
	con, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &tcpConnection{
		addr: addr,
		con:  con,
		dump: t.DumpRoundTrips,
	}

	if err := c.Send(req); err != nil {
		return nil, err
	}

	return c, nil
}

type tcpConnection struct {
	addr string
	con  net.Conn
	dump bool
}

var _ Connection = &tcpConnection{}

func (c *tcpConnection) Close() error {
	return c.con.Close()
}

func (c *tcpConnection) Send(req *Request) error {
	req.Header.Set("Via", fmt.Sprintf("SIP/2.0/TCP %s;branch=1", c.addr))
	if c.dump {
		fmt.Println(req.DebugString())
	}

	if err := req.Write(c.con); err != nil {
		return fmt.Errorf("%w: failed to write request: %s", ErrRoundTripFailed, err)
	}

	return nil
}

func (c *tcpConnection) Recv() (*Response, error) {
	res, err := ParseResponse(c.con)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %s", ErrRoundTripFailed, err)
	}

	if c.dump {
		fmt.Println(res.DebugString())
	}

	res.LocalAddr = c.con.LocalAddr()
	res.RemoteAddr = c.con.RemoteAddr()

	return res, nil
}

func RecvFinal(c Connection) (*Response, error) {
	var res *Response
	var err error

	for {
		res, err = c.Recv()
		if err != nil {
			return nil, err
		}

		if res.StatusCode > 199 {
			break
		}
	}

	return res, nil
}
