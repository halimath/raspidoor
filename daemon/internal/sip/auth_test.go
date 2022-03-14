package sip

import (
	"testing"

	"github.com/go-test/deep"
)

func TestParseWWWAuthenticateHeader(t *testing.T) {
	c, err := parseWWWAuthenticateHeader(` Digest realm="fritz.box", nonce="06FD8995AFA7E3EC"`)
	if err != nil {
		t.Fatal(err)
	}

	if diff := deep.Equal(c, AuthenticationChallenge{
		Method: "Digest",
		Properties: map[string]string{
			"realm": "fritz.box",
			"nonce": "06FD8995AFA7E3EC",
		},
	}); diff != nil {
		t.Error(diff)
	}
}

func TestDigestAuthenticationHandler(t *testing.T) {
	uri, err := ParseURI("sip:test@localhost")
	if err != nil {
		t.Fatal(err)
	}

	h := NewDigestHandler("user", "password")

	r := NewRequest("INVITE", uri)
	c := AuthenticationChallenge{
		Method: "Digest",
		Properties: map[string]string{
			"nonce": "123456789",
		},
	}

	if err := h.Solve(c, r); err != nil {
		t.Fatal(err)
	}

	exp := `Digest username="user", realm="", nonce="123456789", uri="sip:test@localhost:5060", response="6f2dfa09fb298150e9195a987182a7e0"`
	got := r.Header.Get("Authorization")

	if got != exp {
		t.Errorf("expected '%s' but got '%s'", exp, got)
	}
}
