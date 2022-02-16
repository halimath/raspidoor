
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
            status_led=LED(config.status_led.gpio_number),
            bell_pushes=[Switch(p.gpio_number) for p in config.bell_pushes],
        )

    def __init__(self, dialout, status_led, bell_pushes):
        self._dialout = dialout
        self._bell_pushes = bell_pushes
        self._status_led = status_led
        self._event = threading.Event()

        for idx, p in enumerate(self._bell_pushes):
            p.add_listener(functools.partial(self._bell_activate, idx))

    def _bell_activate(self, idx):
        logging.info(f"Door bell {idx} pressed; initiating call")
        self._dialout.initiate_call()

    def run(self):
        logging.info("Starting gatekeeper")
        self._status_led.on()

        self._event.wait()

        self._status_led.off()

    def terminate(self):
        self._event.set()
