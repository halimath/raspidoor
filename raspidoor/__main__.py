#!/usr/bin/bash

import signal
import time
import os
import sys
import logging

from raspidoor.config import load_yaml
from raspidoor.gatekeeper import Gatekeeper
from raspidoor.gpio import init, cleanup
from raspidoor import __version__

print(f"raspi-door v {__version__}")

config = load_yaml('./raspi-door.yaml')

logging.basicConfig(format='%(asctime)s %(levelname)s - %(message)s', datefmt='%Y-%m-%dT%H:%M:%S%z',
                    stream=sys.stdout, encoding='utf-8', level=logging.DEBUG if config.debug else logging.INFO)

logging.debug("Initializing GPIO")
init()

logging.debug("Configuring Gatekeeper")
gatekeeper = Gatekeeper.from_config(config)

logging.debug("Installing signal handler")


def signal_handler(signum, frame):
    logging.info("Received SIGTERM; shutting down")
    gatekeeper.terminate()


signal.signal(signal.SIGTERM, signal_handler)

try:
    logging.debug("Waiting for events")
    gatekeeper.run()
except KeyboardInterrupt:
    print()
    gatekeeper.terminate()


logging.debug("Performing cleanup on GPIO")
cleanup()

logging.info("All done; raspi-door about to exit")
