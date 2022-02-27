package sip

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidAuthenticationChallenge     = errors.New("invalid authentication challenge")
	ErrUnsolveableAuthenticationChallenge = errors.New("unsolveable authentication challenge")
)

type AuthenticationChallenge struct {
	Method     string
	Properties map[string]string
}

func parseWWWAuthenticateHeader(h string) (AuthenticationChallenge, error) {
	h = strings.TrimSpace(h)

	methodAndProps := strings.SplitN(strings.TrimSpace(h), " ", 2)
	if len(methodAndProps) != 2 {
		return AuthenticationChallenge{Method: h}, nil
	}

	c := AuthenticationChallenge{
		Method:     methodAndProps[0],
		Properties: make(map[string]string),
	}

	props := strings.Split(methodAndProps[1], ",")
	for _, prop := range props {
		keyVal := strings.SplitN(strings.TrimSpace(prop), "=", 2)
		if len(keyVal) != 2 {
			return AuthenticationChallenge{}, ErrInvalidAuthenticationChallenge
		}
		c.Properties[keyVal[0]] = strings.Trim(keyVal[1], `"`)
	}

	return c, nil
}

type AuthenticationHandler interface {
	Solve(challenge AuthenticationChallenge, req *Request) error
}

func NewDigestHandler(username, password string) AuthenticationHandler {
	return &digestAuthenticationHandler{
		username: username,
		password: password,
	}
}

type digestAuthenticationHandler struct {
	username, password string
}

func (h *digestAuthenticationHandler) Solve(challenge AuthenticationChallenge, req *Request) error {
	if strings.ToLower(challenge.Method) != "digest" {
		return ErrInvalidAuthenticationChallenge
	}

	h1Sum := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", h.username, challenge.Properties["realm"], h.password)))
	h1 := hex.EncodeToString(h1Sum[:])

	h2Sum := md5.Sum([]byte(fmt.Sprintf("%s:%s", req.Method, req.URI)))
	h2 := hex.EncodeToString(h2Sum[:])

	responseSum := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", h1, challenge.Properties["nonce"], h2)))
	response := hex.EncodeToString(responseSum[:])

	req.Header.Set("Authorization", fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`, h.username, challenge.Properties["realm"], challenge.Properties["nonce"], req.URI, response))

	return nil
}
