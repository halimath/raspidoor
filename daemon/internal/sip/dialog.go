package sip

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	StatusOK               = 200
	StatusRequestCancelled = 487
	StatusDecline          = 603
)

type Dialog struct {
	transport              Transport
	caller                 URI
	authenticationHandlers []AuthenticationHandler
	cseq                   int
	callID                 string
}

func NewDialog(transport Transport, caller URI, authenticationHandlers ...AuthenticationHandler) *Dialog {
	return &Dialog{
		transport:              transport,
		caller:                 caller,
		authenticationHandlers: authenticationHandlers,
		cseq:                   1,
		callID:                 fmt.Sprintf("c%d", time.Now().Unix()),
	}
}

func (d *Dialog) Ring(callee URI, maxRingingTime time.Duration) (bool, error) {
	inviteRequest := d.invite(callee, maxRingingTime)
	con, err := d.transport.Send(inviteRequest)
	if err != nil {
		return false, err
	}
	defer con.Close()

	inviteResponse, err := RecvFinal(con)
	if err != nil {
		return false, err
	}

	var authenticationChallenge AuthenticationChallenge

	if inviteResponse.StatusCode == http.StatusUnauthorized {
		authenticationChallenge, err = parseWWWAuthenticateHeader(inviteResponse.Header.Get("WWW-Authenticate"))
		if err != nil {
			return false, err
		}

		if err := d.authenticate(inviteRequest, authenticationChallenge); err != nil {
			return false, err
		}

		if err := con.Send(inviteRequest); err != nil {
			return false, err
		}
		inviteResponse, err = RecvFinal(con)
	}

	accepted := inviteResponse.StatusCode == StatusOK

	to := inviteResponse.Header.Get("To")

	ack, err := d.ack(callee, to, authenticationChallenge)
	if err != nil {
		return false, err
	}

	if err := con.Send(ack); err != nil {
		return false, err
	}

	d.cseq++

	bye := d.request("BYE", callee)
	bye.Header.Set("To", to)
	if err := con.Send(bye); err != nil {
		return false, err
	}

	resp, err := RecvFinal(con)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != StatusOK {
		return false, fmt.Errorf("Got unexpected status from BYE: %d", resp.StatusCode)
	}

	return accepted, nil
}

func (d *Dialog) invite(callee URI, maxRingingTime time.Duration) *Request {
	r := d.request("INVITE", callee)
	r.Header.Add("Expires", strconv.Itoa(int(maxRingingTime.Seconds())))
	return r
}

func (d *Dialog) ack(callee URI, to string, authenticationChallenge AuthenticationChallenge) (*Request, error) {
	r := d.request("ACK", callee)
	r.Header.Set("To", to)
	r.SetBody("application/sdp", []byte(d.formatSDP()))

	if authenticationChallenge.Method != "" {
		err := d.authenticate(r, authenticationChallenge)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (d *Dialog) formatSDP() string {
	sessionId := time.Now().UnixMilli()
	version := sessionId

	addr := "fe80::b644:fe20:c499:dd4e"

	return fmt.Sprintf("v=0\r\no=klingel1 %d %d IN IP6 %s\r\ns=\r\nc=IN IP6 %s\r\nt=0 0\r\n", sessionId, version, addr, addr)
}

func (d *Dialog) request(method string, callee URI) *Request {
	req := NewRequest(method, callee)
	req.Header.Set("From", d.caller.String())
	req.Header.Set("To", callee.String())
	req.Header.Set("Contact", d.caller.String())
	req.Header.Set("Max-Forwards", "70")
	req.Header.Set("Cseq", fmt.Sprintf("%d %s", d.cseq, req.Method))
	req.Header.Set("Call-ID", d.callID)

	return req
}

func (d *Dialog) authenticate(req *Request, c AuthenticationChallenge) error {
	for _, h := range d.authenticationHandlers {
		err := h.Solve(c, req)
		if err == nil {
			d.cseq++
			req.Header.Set("Cseq", fmt.Sprintf("%d %s", d.cseq, req.Method))
			return nil
		}

		if errors.Is(err, ErrUnsolveableAuthenticationChallenge) {
			continue
		}

		return err
	}

	return ErrUnsolveableAuthenticationChallenge
}
