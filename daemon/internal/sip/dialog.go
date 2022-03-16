package sip

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

const StatusDecline = 603

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

func (d *Dialog) Ring(callee URI) (bool, error) {
	inviteRequest := d.request("INVITE", callee)
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

	if inviteResponse.StatusCode == StatusDecline {
		return false, nil
	}

	if inviteResponse.StatusCode == http.StatusOK {
		to := inviteResponse.Header.Get("To")

		ack := d.request("ACK", callee)
		ack.SetBody("application/sdp", []byte(d.formatSDP()))
		ack.Header.Set("To", to)

		if authenticationChallenge.Method != "" {
			err = d.authenticate(ack, authenticationChallenge)
			if err != nil {
				return false, err
			}
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

		return true, nil
	}

	return false, fmt.Errorf("unexpected status code: %d", inviteResponse.StatusCode)
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
