from .transport import TCPTransport
from .dialog import Dialog

try:
    dialog = Dialog(transport=TCPTransport(host="fritz.box", debug=True),
                    caller="sip:klingel1@fritz.box",
                    username="klingel1",
                    password="password001",
                    )

    dialog.register()
    dialog.invite("sip:**611@fritz.box")

finally:
    dialog.close()
