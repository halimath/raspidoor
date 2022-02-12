import socket

from .request import Response


class TCPTransport:
    def __init__(self, host, port=5060, debug=False):
        self._host = host
        self._port = port
        self._debug = debug

    @property
    def host(self):
        return self._host

    @ property
    def target(self):
        return "{}:{}".format(self._host, self._port)

    @ property
    def protocol(self):
        return "TCP"

    def connect(self):
        self._socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._socket.connect((self._host, self._port))

    def close(self):
        self._socket.close()

    def _send_str(self, s):
        self._socket.send(bytes(s, "UTF-8"))

    def send_request(self, req):
        req.header.set_via(
            "SIP/2.0/TCP {}:{};branch=1".format(self._host, self._port))

        if self._debug:
            print(str(req))

        self._send_str(req.to_wire())

    def read_response(self):
        buf = self._socket.recv(4096)
        resp = Response.parse(buf)
        if self._debug:
            print(str(resp))
        return resp
