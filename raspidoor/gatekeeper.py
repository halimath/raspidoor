
import time
import logging

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

    def __init__(self, dialout, led, door_bells, interval=0.2):
        self._dialout = dialout
        self._door_bells = door_bells
        self._led = led
        self._interval = interval
        self._keep_running = True

    def run(self):
        try:
            logging.info("Starting gatekeeper")

            last_states = [False for b in self._door_bells]

            self._led.on()

            while True:
                if not self._keep_running:
                    break

                time.sleep(self._interval)

                for idx, bell in enumerate(self._door_bells):
                    state = bell.is_pressed
                    last_state = last_states[idx]

                    if state and not last_state:
                        logging.info(
                            "Door bell {} pressed; initiating call".format(idx))
                        last_states[idx] = True
                        self._dialout.initiate_call()
                    else:
                        logging.debug("No state change for door bell {}")
                        last_states[idx] = state
        finally:
            self._led.off()

    def terminate(self):
        self._keep_running = False
