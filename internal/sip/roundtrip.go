package sip

import (
	"errors"
	"fmt"
	"net"
)

var (
	ErrRoundTripFailed = errors.New("round trip failed")
)

type RoundTripper interface {
	RoundTrip(*Request) (*Response, error)
	Close() error
}

type TCPTransport struct {
	DumpRoundTrips bool
}

var _ RoundTripper = &TCPTransport{}

func NewTCPTransport() *TCPTransport {
	return &TCPTransport{}
}

func (c *TCPTransport) Close() error {
	return nil
}

func (c *TCPTransport) RoundTrip(req *Request) (*Response, error) {
	hostAndPort := fmt.Sprintf("%s:%d", req.URI.Host, req.URI.Port)

	con, err := net.Dial("tcp", hostAndPort)
	if err != nil {
		return nil, err
	}
	defer con.Close()

	req.Header.Set("Via", fmt.Sprintf("SIP/2.0/TCP %s;branch=1", hostAndPort))

	if c.DumpRoundTrips {
		fmt.Println(req.DebugString())
	}

	if err := req.Write(con); err != nil {
		return nil, fmt.Errorf("%w: failed to write request: %s", ErrRoundTripFailed, err)
	}

	var res *Response
	for {
		var err error
		res, err = ParseResponse(con)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to read response: %s", ErrRoundTripFailed, err)
		}

		if c.DumpRoundTrips {
			fmt.Println(res.DebugString())
		}

		if res.StatusCode > 199 {
			break
		}
	}

	return res, nil
}
