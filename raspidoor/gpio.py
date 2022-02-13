
import logging
import time

from RPi import GPIO


class LED:
    def __init__(self, gpio_number):
        self._gpio_number = gpio_number
        GPIO.setup(self._gpio_number, GPIO.OUT)
        self.off()

    def on(self):
        self._set(True)

    def off(self):
        self._set(False)

    def blink(self):
        self.on()
        time.sleep(0.1)
        self.off()

    def _set(self, state):
        logging.debug(f"Setting LED on {self._gpio_number} to {state}")
        self._state = state
        GPIO.output(self._gpio_number, GPIO.HIGH if state else GPIO.LOW)


    @property
    def is_on(self):
        return self._state


class Switch:
    def __init__(self, gpio_number, listener=None):
        self._gpio_number = gpio_number
        GPIO.setup(gpio_number, GPIO.IN, pull_up_down=GPIO.PUD_DOWN)
        GPIO.add_event_detect(24, GPIO.RISING, callback=self._callback)

        self._listener = []
        if listener is not None:
            self._listener.append(listener)

    def add_listener(self, listener):
        self._listener.append(listener)

    def _callback(self, _):
        logging.debug(f"Received event for Switch on {self._gpio_number}")
        for l in self._listener:
            l()

def init():
    GPIO.setmode(GPIO.BCM)


def cleanup():
    GPIO.cleanup()
