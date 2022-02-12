
import time

from .request import *
from .auth import AuthenticationChallenge


class CallDeclinedError(Exception):
    pass


class Dialog:
    def __init__(self, transport, caller, username=None, password=None):
        self._transport = transport
        self._username = username
        self._password = password
        self._caller = caller
        self._cseq = 0
        self._call_id = "c{}".format(round(time.time() * 1000))
        self._transport.connect()

    def close(self):
        self._transport.close()

    def register(self):
        req = Request("REGISTER", "sip:{}".format(self._transport.host))
        req.header.\
            set("From", self._caller).\
            set("To", self._caller).\
            set("Contact", self._caller).\
            set("Max-Forwards", 70).\
            set("Expires", 7200)

        reponse = self._exchange(req)

    def invite(self, callee):
        req = Request("INVITE", callee)
        req.header.\
            set("From", self._caller).\
            set("To", callee).\
            set("Contact", self._caller).\
            set("Max-Forwards", 70)

        try:
            response = self._exchange(req)
            if response.status_code == RESPONSE_STATUS_OK:
                return self._bye(callee)
        except CallDeclinedError:
            return

    def _bye(self, callee):
        req = Request("BYE", callee)
        req.header.\
            set("From", self._caller).\
            set("To", callee).\
            set("Contact", self._caller).\
            set("Max-Forwards", 70)

        response = self._exchange(req)
        if response.status_code == RESPONSE_STATUS_OK:
            return
        else:
            raise ValueError("unexpected status code when sending BYE: {} {}".format(
                response.status_code, response.status_message))

    def _exchange(self, request):
        self._send(request)

        resp = self._transport.read_response()

        if resp.status_code == RESPONSE_STATUS_UNAUTHORIZED:
            wwwAuth = resp.header.get("www-authenticate")
            if wwwAuth == None:
                raise ValueError(
                    "Missing WWW-Authenticate header in response with status 401")

            authChallenge = AuthenticationChallenge.parse(wwwAuth)

            if not authChallenge.can_solve():
                raise ValueError(
                    "Unsolvable authentication challenge: {}", authChallenge.method)

            reqWithAuth = authChallenge.solve(
                request, self._username, self._password)
            self._send(request)

            resp = self._transport.read_response()

        while True:
            if resp.status_code > 399:
                raise self.handle_error_response(resp)
            elif resp.status_code < 200:
                resp = self._transport.read_response()
            else:
                break

        return resp

    def handle_error_response(self, response):
        if response.status_code == RESPONSE_STATUS_DECLINE:
            return CallDeclinedError()

        return ValueError("unexpected status code: {} {}".format(
            response.status_code, response.status_message))

    def _send(self, request):
        self._cseq += 1
        request.header.set_call_id(self._call_id).set_cseq(
            "{} {}".format(self._cseq, request.method))

        self._transport.send_request(request)
