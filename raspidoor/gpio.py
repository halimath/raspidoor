from RPi import GPIO
import time


class LED:
    def __init__(self, gpio_number):
        self._gpio_number = gpio_number
        GPIO.setup(self._gpio_number, GPIO.OUT)
        self.off()

    def on(self):
        self._state = True
        GPIO.output(self._gpio_number, GPIO.HIGH)

    def off(self):
        self._state = False
        GPIO.output(self._gpio_number, GPIO.LOW)

    def blink(self):
        self.on()
        time.sleep(0.1)
        self.off()

    @property
    def is_on(self):
        return self._state


class Switch:
    def __init__(self, gpio_number):
        self._gpio_number = gpio_number
        GPIO.setup(self._gpio_number, GPIO.IN)

    @property
    def is_pressed(self):
        return GPIO.input(self._gpio_number) == GPIO.HIGH


def init():
    GPIO.setmode(GPIO.BCM)


def cleanup():
    GPIO.cleanup()
