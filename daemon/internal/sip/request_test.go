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
			"Www-Authenticate": []string{`Digest nonce="1234", realm="test.example.com"`},
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

func TestHeader_Write(t *testing.T) {
	h := Header{}
	h.Add("Foo", "bar")
	h.Add("Max-Forwards", "70")

	var w strings.Builder

	if err := h.Write(&w); err != nil {
		t.Fatal(err)
	}

	got := w.String()

	wantLike := "Foo: bar\r\nMax-Forwards: 70\r\n"
	if len(got) != len(wantLike) {
		t.Errorf("expected '%s' to have length %d but got %d", got, len(wantLike), len(got))
	}

	if !strings.Contains(got, "Foo: bar\r\n") {
		t.Errorf("expected '%s' to contain 'Foo: bar'", got)
	}
	if !strings.Contains(got, "Max-Forwards: 70\r\n") {
		t.Errorf("expected '%s' to contain 'Max-Forwards: 70'", got)
	}
}

func TestRequest_Write(t *testing.T) {
	uri, err := ParseURI("sip:test@localhost")
	if err != nil {
		t.Fatal(err)
	}

	r := NewRequest("INVITE", uri)
	r.Header.Add("Max-Forwards", "70")

	var w strings.Builder

	if err := r.Write(&w); err != nil {
		t.Fatal(err)
	}

	got := w.String()
	wantLike := "INVITE sip:test@localhost:5060 SIP/2.0\r\nMax-Forwards: 70\r\nContent-Length: 0\r\n\r\n"

	if len(got) != len(wantLike) {
		t.Errorf("expected '%s' to have length %d but got %d", got, len(wantLike), len(got))
	}

	wantParts := strings.Split(wantLike, "\r\n")
	for i, p := range wantParts {
		var test func(string, string) bool

		if i == 0 {
			test = strings.HasPrefix
		} else {
			test = strings.Contains
		}

		if !test(got, p) {
			t.Errorf("expected '%s' to contain '%s", got, p)
		}
	}
}
