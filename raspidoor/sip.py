import hashlib
import os
import socket
import time

class Header:
    def __init__(self):
        self._values = {}

    def add_from_wire_line(self, line):
        (key, *value) = line.split(':')
        return self.set(key.lower(), ':'.join(value).strip())

    def set(self, key, value):
        self._values[key] = value
        return self

    def get(self, key):
        return self._values[key]

    def contains(self, key):
        return key in self._values

    def set_via(self, via):
        return self.set("Via", via)

    def set_call_id(self, call_id):
        return self.set("Call-ID", call_id)

    def set_cseq(self, cseq):
        return self.set("Cseq", cseq)

    def to_wire_lines(self):
        return '\r\n'.join([f"{k}: {v}" for k, v in self._values.items()])


class Request:
    def __init__(self, method, uri, header=None, body=None):
        self._method = method
        self._uri = uri
        self._header = Header() if header == None else header
        self._body = body

        self._header.set("Content-Length",
                         0 if self._body is None else len(self._body))

    @ property
    def method(self):
        return self._method

    @ property
    def uri(self):
        return self._uri

    @ property
    def header(self):
        return self._header

    @ property
    def body(self):
        return self._body

    def __str__(self):
        return self.to_wire()

    def to_wire(self):
        r = f"{self._method} {self._uri} SIP/2.0\r\n"

        r += self._header.to_wire_lines()

        r += '\r\n\r\n'

        if self._body != None:
            r += self._body

        return r


RESPONSE_STATUS_OK = 200
RESPONSE_STATUS_UNAUTHORIZED = 401
RESPONSE_STATUS_DECLINE = 603


class Response:
    @ staticmethod
    def parse(data):
        if len(data) == 0:
            raise ValueError("Invalid response data length")

        s = str(data, ' utf-8')
        lines = s.splitlines()

        (protocol, status_code, *status_message) = lines[0].split(' ')

        header = Header()

        for line in lines[1:]:
            line = line.strip()
            if len(line) == 0:
                break
            header.add_from_wire_line(line)

        return Response(status_code=int(status_code), status_message=' '.join(status_message), protocol=protocol, header=header)

    def __init__(self, status_code, status_message, protocol, header, body=None):
        self._status_code = status_code
        self._status_message = status_message
        self._protocol = protocol
        self._header = header
        self._body = body

    @ property
    def status_code(self):
        return self._status_code

    @ property
    def status_message(self):
        return self._status_message

    @ property
    def protocol(self):
        return self._protocol

    @ property
    def header(self):
        return self._header

    @ property
    def body(self):
        return self._body

    def __str__(self):
        r = f"{self._protocol} {self._status_code} {self._status_message}\r\n"
        r += self._header.to_wire_lines()
        r += '\r\n\r\n'
        if self._body is not None:
            r += self._body

        return r

class DigestAuthenticationChallenge:
    @staticmethod
    def from_request_response(request, response):
        www_authenticate_header = response.header.get('www-authenticate')
        if www_authenticate_header == None:
            raise ValueError(
                "Missing WWW-Authenticate header in response with status 401")

        (scheme, *pairs) = www_authenticate_header.strip().split(' ')

        if scheme.lower() != "digest":
            raise ValueError(f"Unsupported authentication challenge scheme: '{scheme}'")

        realm = None
        nonce = None

        for p in ' '.join(pairs).split(','):
            (key, val) = p.strip().split('=')
            val = val.strip('"')
            if key.lower() == 'realm':
                realm = val
            elif key.lower() == 'nonce':
                nonce = val
            else:
                raise ValueError(f"Unsupported authentication challenge key: '{key}'")

        return DigestAuthenticationChallenge(
            method=request.method, 
            uri=request.uri,
            realm=realm, 
            nonce=nonce,
        )

    def __init__(self, method, uri, realm, nonce=None):
        self.method = method
        self.uri = uri
        self.realm = realm
        self.nonce = nonce

class DigestAuthenticationHandler:
    def __init__(self, username, password):
        self._username = username
        self._password = password

    def solve(self, request, challenge):
        h1 = hashlib.md5(f"{self._username}:{challenge.realm}:{self._password}".encode()).digest().hex()
        h2 = hashlib.md5(f"{challenge.method}:{challenge.uri}".encode()).digest().hex()
        response = hashlib.md5(f"{h1}:{challenge.nonce}:{h2}".encode()).digest().hex()

        request.header.set('Authorization', f'Digest username="{self._username}", realm="{challenge.realm}", nonce="{challenge.nonce}", uri="{challenge.uri}", response="{response}"')
        return request

class TCPTransport:
    def __init__(self, host, port=5060, debug=False):
        self._host = host
        self._port = port
        self._debug = debug

    def roundtripper(self):
        return TCPRoundtripper(self._host, self._port, self._debug)

class TCPRoundtripper:
    def __init__(self, host, port=5060, debug=False):
        self._host = host
        self._port = port
        self._debug = debug

    @property
    def host(self):
        return self._host

    @ property
    def target(self):
        return '{}:{}'.format(self._host, self._port)

    @ property
    def protocol(self):
        return 'TCP'

    def connect(self):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect((self._host, self._port))

    def close(self):
        self._socket.close()

    def _send_str(self, s):
        self._socket.send(bytes(s, 'utf-8'))

    def send_request(self, req):
        req.header.set_via(f"SIP/2.0/TCP {self._host}:{self._port};branch=1")

        if self._debug:
            print(str(req))

        self._send_str(req.to_wire())

    def read_response(self):
        buf = self._socket.recv(4096)
        resp = Response.parse(buf)
        if self._debug:
            print(str(resp))
        return resp


class CallDeclinedError(Exception):
    pass


class Client:
    def __init__(self, transport, auth_handler=None):
        self._transport = transport
        self._auth_handler = auth_handler

    def start_dialog(self, caller):
        return Dialog(
            roundtripper=self._transport.roundtripper(),
            caller=caller,
            auth_handler=self._auth_handler,
        )


class Dialog:
    def __init__(self, roundtripper, caller, auth_handler=None):
        self._roundtripper = roundtripper
        self._caller = caller        
        self._auth_handler=auth_handler

    def __enter__(self):
        self.connect()
        return self

    def __exit__(self, type, value, traceback):
        self.close()

    def connect(self):
        self._roundtripper.connect()

    def close(self):
        self._roundtripper.close()

    def register(self):
        self._new_call()

        req = Request('REGISTER', f"sip:{self._roundtripper.host}")
        req.header.\
            set('From', self._caller).\
            set('To', self._caller).\
            set('Contact', self._caller).\
            set('Max-Forwards', 70).\
            set('Expires', 7200)

        reponse = self._exchange(req)

    def invite(self, callee):
        self._new_call()

        req = Request('INVITE', callee)
        req.header.\
            set('From', self._caller).\
            set('To', callee).\
            set('Contact', self._caller).\
            set('Max-Forwards', 70)

        try:
            response = self._exchange(req)              
            self._ack(callee)
            return self._bye(callee)
        except CallDeclinedError:
            return self._bye(callee)

    def _bye(self, callee):
        req = Request('BYE', callee)
        req.header.\
            set('From', self._caller).\
            set('To', callee).\
            set('Contact', self._caller).\
            set('Max-Forwards', 70)

        response = self._exchange(req)
        if response.status_code == RESPONSE_STATUS_OK:
            return
        else:
            raise ValueError(f"Unexpected status code when sending BYE: {response.status_code} {response.status_message}")

    def _ack(self, callee):
        req = Request('ACK', callee)
        req.header.\
            set('From', self._caller).\
            set('To', callee).\
            set('Contact', self._caller).\
            set('Max-Forwards', 70)

        response = self._exchange(req)
        if response.status_code == RESPONSE_STATUS_OK:
            return
        else:
            raise ValueError(f"Unexpected status code when sending ACK: {response.status_code} {response.status_message}")

    def _new_call(self):
        self._cseq = 0
        self._call_id = f"c{round(time.time() * 1000)}"

    def _exchange(self, request):
        self._send(request)

        response = self._roundtripper.read_response()

        if response.status_code == RESPONSE_STATUS_UNAUTHORIZED:
            auth_challenge = DigestAuthenticationChallenge.from_request_response(request, response)

            reqWithAuth = self._auth_handler.solve(request, auth_challenge)
            
            self._send(reqWithAuth)
            response = self._roundtripper.read_response()

        while True:
            if response.status_code > 399:
                raise self.handle_error_response(response)
            elif response.status_code < 200:
                response = self._roundtripper.read_response()
            else:
                break

        return response

    def handle_error_response(self, response):
        if response.status_code == RESPONSE_STATUS_DECLINE:
            return CallDeclinedError()

        return ValueError(f"Unexpected status code: {response.status_code} {response.status_message}")

    def _send(self, request):
        self._cseq += 1
        request.header.set_call_id(self._call_id).set_cseq(
            "{} {}".format(self._cseq, request.method))

        self._roundtripper.send_request(request)


if __name__ == '__main__':
    client = Client(
        transport=TCPTransport(host='fritz.box', debug=True),
        auth_handler=DigestAuthenticationHandler('klingel1', 'password001'),
    )

    with client.start_dialog('sip:klingel1@fritz.box') as dialog:
        dialog.invite('sip:**611@fritz.box')
