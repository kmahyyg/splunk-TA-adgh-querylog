#!/usr/bin/env python3

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "lib"))

import splunklib.client as splunkclient
import json
import base64
from dns.message import from_wire as from_dnsmsg
