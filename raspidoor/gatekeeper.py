
import functools
import logging
import threading
import time

from raspidoor.dialout import Dialout
from raspidoor.gpio import LED, Switch
from raspidoor import __version__


class Gatekeeper:

    @staticmethod
    def from_config(config):
        return Gatekeeper(
            dialout=Dialout.from_config(config),
            led=LED(config.gpio.led),
            door_bells=[Switch(gpio) for gpio in config.gpio.door_bells],
        )

    def __init__(self, dialout, led, door_bells):
        self._dialout = dialout
        self._door_bells = door_bells
        self._led = led
        self._event = threading.Event()

        for idx, b in enumerate(self._door_bells):
            b.add_listener(functools.partial(self._bell_activate, idx))

    def _bell_activate(self, idx):
        logging.info(f"Door bell {idx} pressed; initiating call")
        self._dialout.initiate_call()

    def run(self):
        logging.info("Starting gatekeeper")
        self._led.on()

        self._event.wait()

        self._led.off()

    def terminate(self):
        self._event.set()
