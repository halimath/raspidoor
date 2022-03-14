package sip

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Dialog struct {
	roundTripper           RoundTripper
	caller                 URI
	authenticationHandlers []AuthenticationHandler
	cseq                   int
	callID                 string
}

func NewDialog(roundTripper RoundTripper, caller URI, authenticationHandlers ...AuthenticationHandler) *Dialog {
	return &Dialog{
		roundTripper:           roundTripper,
		caller:                 caller,
		authenticationHandlers: authenticationHandlers,
	}
}

func (d *Dialog) newCall() {
	d.cseq = 0
	d.callID = fmt.Sprintf("c%d", time.Now().Unix())
}

const StatusDecline = 603

func (d *Dialog) Ring(callee URI) error {
	d.newCall()

	res, err := d.sendRequest("INVITE", callee, true)

	if err != nil {
		return err
	}

	if res.StatusCode == StatusDecline {
		return nil
	}

	if res.StatusCode == http.StatusOK {
		res, err = d.sendRequest("ACK", callee, false)

		// TODO: Send ACK with same CSeq
		res, err = d.sendRequest("BYE", callee, true)
		if err != nil {
			return err
		}
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", res.StatusCode)
		}
	}

	return fmt.Errorf("unexpected status code: %d", res.StatusCode)
}

func (d *Dialog) sendRequest(method string, callee URI, incrementCSeq bool) (*Response, error) {
	req := NewRequest(method, callee)
	req.Header.Set("From", d.caller.String())
	req.Header.Set("To", callee.String())
	req.Header.Set("Contact", d.caller.String())
	req.Header.Set("Max-Forwards", "70")

	return d.exchange(req, incrementCSeq)
}

func (d *Dialog) exchange(req *Request, incrementCSeq bool) (*Response, error) {
	res, err := d.roundTrip(req, incrementCSeq)
	if err != nil {
		return nil, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		c, err := parseWWWAuthenticateHeader(res.Header.Get("WWW-Authenticate"))
		if err != nil {
			return nil, err
		}

		if err := d.authenticate(req, c); err != nil {
			return nil, err
		}

		res, err = d.roundTrip(req, true)
	}

	return res, nil
}

func (d *Dialog) authenticate(req *Request, c AuthenticationChallenge) error {
	for _, h := range d.authenticationHandlers {
		err := h.Solve(c, req)
		if err == nil {
			return nil
		}

		if errors.Is(err, ErrUnsolveableAuthenticationChallenge) {
			continue
		}

		return err
	}

	return ErrUnsolveableAuthenticationChallenge
}

func (d *Dialog) roundTrip(req *Request, incrementCSeq bool) (*Response, error) {
	if incrementCSeq {
		d.cseq++
	}
	req.Header.Set("Cseq", fmt.Sprintf("%d %s", d.cseq, req.Method))
	req.Header.Set("Call-ID", d.callID)

	return d.roundTripper.RoundTrip(req)
}
