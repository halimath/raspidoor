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
