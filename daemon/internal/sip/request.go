package sip

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	ErrParsingError = errors.New("error parsing response")
	ErrInvaldURI    = errors.New("invalid uri")
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

type Header map[string][]string

func (h Header) Contains(key string) bool {
	_, ok := h[key]
	return ok
}

func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

func (h Header) ParseHeader(line string) error {
	parts := strings.Split(line, ":")
	if len(parts) == 1 {
		return fmt.Errorf("%w: invalid header line: %s", ErrParsingError, line)
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(strings.Join(parts[1:], ":"))

	h.Add(key, val)

	return nil
}

func (h Header) Write(w io.Writer) error {
	for k, vals := range h {
		for _, val := range vals {
			if _, err := io.WriteString(w, k); err != nil {
				return err
			}
			if _, err := io.WriteString(w, ": "); err != nil {
				return err
			}
			if _, err := io.WriteString(w, val); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\r\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h Header) DebugString() string {
	var b strings.Builder
	h.Write(&b)
	return b.String()
}

type Request struct {
	Method string
	URI    URI
	Header Header
	Body   []byte
}

func NewRequest(method string, uri URI) *Request {
	header := Header{}
	header.Set("content-length", "0")

	return &Request{
		Method: method,
		URI:    uri,
		Header: header,
	}
}

func (r *Request) Write(w io.Writer) error {
	if _, err := io.WriteString(w, fmt.Sprintf("%s %s SIP/2.0\r\n", r.Method, r.URI)); err != nil {
		return err
	}

	if err := r.Header.Write(w); err != nil {
		return err
	}

	if _, err := io.WriteString(w, "\r\n"); err != nil {
		return err
	}

	if len(r.Body) == 0 {
		return nil
	}

	_, err := w.Write(r.Body)
	return err
}

func (r *Request) DebugString() string {
	var b strings.Builder
	r.Write(&b)
	return b.String()
}

type Response struct {
	StatusCode    int
	StatusMessage string
	Protocol      string
	Header        Header
	Body          []byte
}

func (r *Response) DebugString() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d %s\n", r.Protocol, r.StatusCode, r.StatusMessage)
	b.WriteString(r.Header.DebugString())
	b.WriteRune('\n')
	return b.String()
}

func ParseResponse(r io.Reader) (*Response, error) {
	scanner := bufio.NewScanner(r)

	if !scanner.Scan() {
		return nil, fmt.Errorf("%w: empty reader", ErrParsingError)
	}

	protocol, statusCodeAsString, statusMessage, err := parseFirstResponseLine(scanner.Text())
	if err != nil {
		return nil, err
	}

	statusCode, err := strconv.ParseInt(statusCodeAsString, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing status code: %s", ErrParsingError, err)
	}

	res := &Response{
		Protocol:      protocol,
		StatusCode:    int(statusCode),
		StatusMessage: statusMessage,
		Header:        Header{},
		Body:          nil,
	}

	for scanner.Scan() {
		line := scanner.Text()
		if len(strings.TrimSpace(line)) == 0 {
			// End of body
			// TODO
			break
		}
		if err := res.Header.ParseHeader(line); err != nil {
			return nil, err
		}
	}

	return res, nil
}

// func ParseResponse(r io.Reader) (*Response, error) {
// 	br := bufio.NewReader(r)
// 	pr := textproto.NewReader(br)

// 	l, err := pr.ReadLine()
// 	if err != nil {
// 		return nil, err
// 	}

// 	protocol, statusCodeAsString, statusMessage, err := parseFirstResponseLine(l)
// 	if err != nil {
// 		return nil, err
// 	}

// 	statusCode, err := strconv.ParseInt(statusCodeAsString, 10, 32)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: error parsing status code: %s", ErrParsingError, err)
// 	}

// 	res := &Response{
// 		Protocol:      protocol,
// 		StatusCode:    int(statusCode),
// 		StatusMessage: statusMessage,
// 		Header:        Header{},
// 		Body:          nil,
// 	}

// 	for {
// 		line, err := pr.ReadContinuedLine()
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				break
// 			}
// 			return nil, err
// 		}

// 		if len(strings.TrimSpace(line)) == 0 {
// 			// End of body
// 			// TODO: Parse body
// 			break
// 		}

// 		if err := res.Header.ParseHeader(line); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return res, nil
// }

func parseFirstResponseLine(l string) (protocol, statusCode, statusMessage string, err error) {
	parts := strings.Split(l, " ")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("%w: invalid response line: %s", ErrParsingError, l)
	}

	return parts[0], parts[1], strings.TrimSpace(strings.Join(parts[2:], " ")), nil
}
