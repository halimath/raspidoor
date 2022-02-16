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

class LED:
    """
    Configuration for the status LED.
    """
    def __init__(self, gpio_number):
        self.gpio_number = gpio_number        

class BellPush:
    """
    A single bell push the system should react on.
    """
    def __init__(self, gpio_number, label=None):
        self.gpio_number = gpio_number
        self.label = label

class Config:
    """
    Config is the root class for storing the configuration
    """

    def __init__(self, sip, status_led, bell_pushes, debug):
        self.sip = sip
        self.status_led = status_led
        self.bell_pushes = bell_pushes
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
            status_led=LED(config['status_led']['gpio']),
            bell_pushes=[BellPush(gpio_number=p['gpio'], label=p['label']) for p in config['bell_pushes']],
            debug=config['debug'],
        )
