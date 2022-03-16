package sip

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvaldURI = errors.New("invalid uri")
)

type URI struct {
	Scheme  string
	Address string
	Host    string
	Port    int
}

func NewURI(scheme, address, host string, port int) URI {
	return URI{
		Scheme:  scheme,
		Address: address,
		Host:    host,
		Port:    port,
	}
}

func ParseURI(uri string) (URI, error) {
	uri = strings.TrimSpace(uri)

	p := strings.SplitN(uri, ":", 2)
	if len(p) != 2 {
		return URI{}, fmt.Errorf("%w: missing scheme", ErrInvaldURI)
	}
	if p[0] != "sip" {
		return URI{}, fmt.Errorf("%w: missing scheme: %s", ErrInvaldURI, p[0])
	}

	scheme := p[0]

	p = strings.Split(p[1], "@")
	if len(p) != 2 {
		return URI{}, fmt.Errorf("%w: missing host", ErrInvaldURI)
	}

	address := p[0]

	p = strings.Split(p[1], ":")
	var host string
	var port int64
	var err error
	if len(p) == 1 {
		host = p[0]
		port = 5060
	} else {
		host = p[0]
		port, err = strconv.ParseInt(p[1], 10, 32)
		if err != nil {
			return URI{}, fmt.Errorf("%w: invalid host: %s", ErrInvaldURI, err)
		}
	}

	return URI{
		Scheme:  scheme,
		Address: address,
		Host:    host,
		Port:    int(port),
	}, nil
}

func (u URI) String() string {
	return fmt.Sprintf("%s:%s@%s:%d", u.Scheme, u.Address, u.Host, u.Port)
}
