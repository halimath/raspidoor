package sip

import (
	"strings"
	"testing"

	"github.com/go-test/deep"
)

func TestParseResponse(t *testing.T) {
	responseData := `SIP/2.0/TCP 401 Unauthorized
Content-Length: 0
WWW-Authenticate: Digest nonce="1234", realm="test.example.com"

	`

	res, err := ParseResponse(strings.NewReader(responseData))
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(*res, Response{
		Protocol:      "SIP/2.0/TCP",
		StatusCode:    401,
		StatusMessage: "Unauthorized",
		Header: Header{
			"Content-Length":   []string{"0"},
			"WWW-Authenticate": []string{`Digest nonce="1234", realm="test.example.com"`},
		},
	}); diff != nil {
		t.Error(diff)
	}
}

func TestParseURI(t *testing.T) {
	uri, err := ParseURI("sip:**612@192.168.1.1:5060")
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(uri, URI{
		Scheme:  "sip",
		Address: "**612",
		Host:    "192.168.1.1",
		Port:    5060,
	}); diff != nil {
		t.Error(diff)
	}
}
