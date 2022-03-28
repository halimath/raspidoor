package sip

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

var (
	ErrParsingError = errors.New("error parsing response")
)

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

func (r *Request) SetBody(contentType string, body []byte) {
	r.Header.Set("Content-Type", contentType)
	r.Header.Set("Content-Length", strconv.Itoa(len(body)))
	r.Body = body
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
	StatusCode            int
	StatusMessage         string
	Protocol              string
	Header                Header
	Body                  []byte
	LocalAddr, RemoteAddr net.Addr
}

func (r *Response) DebugString() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %d %s\n", r.Protocol, r.StatusCode, r.StatusMessage)
	b.WriteString(r.Header.DebugString())
	b.WriteRune('\n')

	if len(r.Body) > 0 {
		b.Write(r.Body)
	}

	return b.String()
}

func ParseResponse(r io.Reader) (*Response, error) {
	br := bufio.NewReader(r)
	pr := textproto.NewReader(br)

	l, err := pr.ReadLine()
	if err != nil {
		return nil, err
	}

	protocol, statusCodeAsString, statusMessage, err := parseFirstResponseLine(l)
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

	for {
		line, err := pr.ReadContinuedLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if len(strings.TrimSpace(line)) == 0 {
			// End of body
			// TODO: Parse body
			break
		}

		if err := res.Header.ParseHeader(line); err != nil {
			return nil, err
		}
	}

	var contentLength int

	contentLengthHeader := res.Header.Get("Content-Length")
	if contentLengthHeader != "" {
		contentLength, err = strconv.Atoi(contentLengthHeader)
		if err != nil {
			return nil, err
		}
	}

	if contentLength > 0 {
		body := make([]byte, contentLength)
		_, err = io.ReadFull(br, body)
		if err != nil {
			return nil, err
		}
		res.Body = body
	}

	return res, nil
}

func parseFirstResponseLine(l string) (protocol, statusCode, statusMessage string, err error) {
	parts := strings.Split(l, " ")
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("%w: invalid response line: %s", ErrParsingError, l)
	}

	return parts[0], parts[1], strings.TrimSpace(strings.Join(parts[2:], " ")), nil
}
