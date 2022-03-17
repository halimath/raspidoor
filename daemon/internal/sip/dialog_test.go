package sip

import (
	"strings"
	"testing"
	"time"
)

type transportMock struct {
	resps []*Response
}

var _ Transport = &transportMock{}

func (t *transportMock) Send(r *Request) (Connection, error) {
	return &connectionMock{
		reqs:  []*Request{r},
		resps: t.resps,
	}, nil
}

type connectionMock struct {
	reqs      []*Request
	respIndex int
	resps     []*Response
}

func (c *connectionMock) Send(r *Request) error {
	c.reqs = append(c.reqs, r)
	return nil
}

func (c *connectionMock) Recv() (*Response, error) {
	r := c.resps[c.respIndex]
	c.respIndex++
	return r, nil
}

func (*connectionMock) Close() error { return nil }

type authHandlerMock struct{}

var _ AuthenticationHandler = &authHandlerMock{}

func (*authHandlerMock) Solve(challenge AuthenticationChallenge, req *Request) error {
	req.Header.Add("Authorize", "Solved")
	return nil
}

func TestDialog_Ring_decline(t *testing.T) {
	tm := &transportMock{
		resps: []*Response{
			resp("SIP/2.0/TCP 401 Unauthorized\r\nContent-Length: 0\r\nWWW-Authenticate: Digest nonce=\"1234\", realm=\"test.example.com\"\r\n"),
			resp("SIP/2.0/TCP 603 Declined\r\nContent-Length: 0\r\n\r\n"),
			resp("SIP/2.0/TCP 200 OK\r\nContent-Length: 0\r\n\r\n"),
		},
	}

	caller, err := ParseURI("sip:caller@localhost")
	if err != nil {
		t.Fatal(err)
	}
	callee, err := ParseURI("sip:callee@localhost")
	if err != nil {
		t.Fatal(err)
	}

	d := NewDialog(tm, caller, &authHandlerMock{})
	accepted, err := d.Ring(callee, time.Second)

	if err != nil {
		t.Error(err)
	}

	if accepted {
		t.Errorf("expected declined but got accepted")
	}
}

func TestDialog_Ring_ringingThenOK(t *testing.T) {
	tm := &transportMock{
		resps: []*Response{
			resp("SIP/2.0/TCP 401 Unauthorized\r\nContent-Length: 0\r\nWWW-Authenticate: Digest nonce=\"1234\", realm=\"test.example.com\"\r\n"),
			resp("SIP/2.0/TCP 180 Ringing\r\nContent-Length: 0\r\n\r\n"),
			resp("SIP/2.0/TCP 200 OK\r\nContent-Length: 0\r\n\r\n"),
			resp("SIP/2.0/TCP 200 OK\r\nContent-Length: 0\r\n\r\n"),
		},
	}

	caller, err := ParseURI("sip:caller@localhost")
	if err != nil {
		t.Fatal(err)
	}
	callee, err := ParseURI("sip:callee@localhost")
	if err != nil {
		t.Fatal(err)
	}

	d := NewDialog(tm, caller, &authHandlerMock{})
	accepted, err := d.Ring(callee, time.Second)
	if err != nil {
		t.Error(err)
	}

	if !accepted {
		t.Errorf("expected accepted but got declined")
	}
}

func resp(s string) *Response {
	r, err := ParseResponse(strings.NewReader(s))
	if err != nil {
		panic(err)
	}

	return r
}
