
import hashlib
import os


class AuthenticationChallenge:
    def parse(header):
        (method, *pairs) = header.strip().split(" ")
        realm = None
        nonce = None

        for p in " ".join(pairs).split(","):
            (key, val) = p.strip().split('=')
            val = val.strip('"')
            if key.lower() == "realm":
                realm = val
            elif key.lower() == "nonce":
                nonce = val
            else:
                raise ValueError(
                    "Unknown authentication challenge key: '{}'".format(key))

        return AuthenticationChallenge(method, realm=realm, nonce=nonce)

    def __init__(self, method, realm, nonce=None):
        self._method = method
        self._realm = realm
        self._nonce = nonce

    @ property
    def method(self):
        return self._method

    def can_solve(self):
        return self._method.lower() == "digest"

    def solve(self, request, username, password):
        cnonce = os.urandom(16).hex()
        nc = "00000001"

        h1 = hashlib.md5(
            ":".join([username, self._realm, password]).encode()).digest().hex()

        h2 = hashlib.md5(
            ":".join([request.method, request.uri]).encode()).digest().hex()

        response = hashlib.md5(
            ":".join([h1, self._nonce, h2]).encode()).digest().hex()

        return request.header.set("Authorization", 'Digest username="{}", realm="{}", nonce="{}", uri="{}", response="{}"'.format(username, self._realm, self._nonce, request.uri, response))
