import yaml


class SIPServer:
    """
    SIPServer defines the configuration for a SIP server (or proxy) to talk to.
    """

    def __init__(self, user, password, host, port, debug=False):
        self.host = host
        self.port = port
        self.user = user
        self.password = password
        self.debug = debug


class SIP:
    """
    SIP defines the overall SIP configuration.
    """

    def __init__(self, caller, callee, server):
        self.caller = caller
        self.callee = callee
        self.server = server


class GPIO:
    """
    GPIO defines the configuration for the GPIO ports to use.
    """

    def __init__(self, led, door_bells):
        self.led = led
        self.door_bells = door_bells


class Config:
    """
    Config is the root class for storing the configuration
    """

    def __init__(self, sip, gpio, debug):
        self.sip = sip
        self.gpio = gpio
        self.debug = debug


def load_yaml(filename):
    """
    load_yaml loads YAML configuration from the file named filename, parses it, and returns a Config instance
    holding the values.
    """
    with open(filename, 'r') as file:
        config = yaml.safe_load(file)

        return Config(
            sip=SIP(
                caller=config['sip']['caller'],
                callee=config['sip']['callee'],
                server=SIPServer(
                    host=config['sip']['server']['host'],
                    port=config['sip']['server']['port'],
                    user=config['sip']['server']['user'],
                    password=config['sip']['server']['password'],
                    debug=config['sip']['server']['debug'],
                )
            ),
            gpio=GPIO(
                led=config['gpio']['led'],
                door_bells=config['gpio']['door-bells'],
            ),
            debug=config['debug'],
        )
