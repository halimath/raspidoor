
class Header:
    def __init__(self):
        self._values = {}

    def add_from_wire_line(self, line):
        (key, *value) = line.split(":")
        return self.set(key.lower(), ":".join(value).strip())

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
        return "\r\n".join(["{}: {}".format(k, v) for k, v in self._values.items()])


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
        r = "{} {} SIP/2.0\r\n".format(self._method, self._uri)

        r += self._header.to_wire_lines()

        r += "\r\n\r\n"

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

        s = str(data, "UTF-8")
        lines = s.splitlines()

        (protocol, status_code, *status_message) = lines[0].split(" ")

        header = Header()

        for line in lines[1:]:
            line = line.strip()
            if len(line) == 0:
                break
            header.add_from_wire_line(line)

        return Response(status_code=int(status_code), status_message=" ".join(status_message), protocol=protocol, header=header)

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
        r = "{} {} {}\r\n".format(
            self._protocol, self._status_code, self._status_message)
        r += self._header.to_wire_lines()
        r += "\r\n\r\n"
        if self._body is not None:
            r += self._body

        return r
