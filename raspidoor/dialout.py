
from raspidoor.sip import TCPTransport, Client, DigestAuthenticationHandler
class Dialout:
    @staticmethod
    def from_config(config):
        sip_client = Client(
            transport=TCPTransport(
                host=config.sip.server.host, port=config.sip.server.port, debug=config.sip.server.debug),
            auth_handler=DigestAuthenticationHandler(
                username=config.sip.server.user, 
                password=config.sip.server.password,
            ),
        )
        return Dialout(
            client=sip_client,
            caller=config.sip.caller,
            callee=config.sip.callee,
        )

    def __init__(self, client, caller, callee):
        self._client = client
        self._caller = caller
        self._callee = callee

    def initiate_call(self):
        with self._client.start_dialog(self._caller) as dialog:
            dialog.invite(self._callee)
