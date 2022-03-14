package sip

import (
	"strings"
	"testing"
)

type roundTripperMock struct {
	reqs  []*Request
	resps []*Response
}

var _ RoundTripper = &roundTripperMock{}

func (t *roundTripperMock) RoundTrip(r *Request) (*Response, error) {
	t.reqs = append(t.reqs, r)
	return t.resps[len(t.reqs)-1], nil
}
func (t *roundTripperMock) Close() error { return nil }

type authHandlerMock struct{}

var _ AuthenticationHandler = &authHandlerMock{}

func (*authHandlerMock) Solve(challenge AuthenticationChallenge, req *Request) error {
	req.Header.Add("Authorize", "Solved")
	return nil
}

func TestDialog_Ring(t *testing.T) {
	tp := &roundTripperMock{
		resps: []*Response{
			resp("SIP/2.0/TCP 401 Unauthorized\r\nContent-Length: 0\r\nWWW-Authenticate: Digest nonce=\"1234\", realm=\"test.example.com\"\r\n"),
			resp("SIP/2.0/TCP 603 Declined\r\nContent-Length: 0\r\n\r\n"),
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

	d := NewDialog(tp, caller, &authHandlerMock{})
	if err := d.Ring(callee); err != nil {
		t.Error(err)
	}
}

func resp(s string) *Response {
	r, err := ParseResponse(strings.NewReader(s))
	if err != nil {
		panic(err)
	}

	return r
}
