
from raspidoor.sip.transport import TCPTransport
from raspidoor.sip.dialog import Dialog


class Dialout:
    @staticmethod
    def from_config(config):
        return Dialout(
            dialog=Dialog(transport=TCPTransport(
                host=config.sip.server.host, port=config.sip.server.port, debug=config.debug),
                caller=config.sip.caller,
                username=config.sip.server.user,
                password=config.sip.server.password,
            ),
            callee=config.sip.callee
        )

    def __init__(self, callee, dialog):
        self.callee = callee
        self._dialog = dialog
        self._dialog.register()

    def close(self):
        self._dialog.close()

    def initiate_call(self):
        self._dialog.invite(self.callee)
