#!/usr/bin/env python
# -*- coding: utf-8 -*-

from flask import Flask
from time import sleep
app = Flask(__name__)

counter = 1

@app.route("/")
def hello():
    global counter
    counter += 1
    return str(counter)

timeout_s = 10
@app.route("/timeout")
def timeout():
    sleep(timeout_s)
    return str("Waited {}s".format(timeout_s))

PUB_KEY = "localhost.crt"
PRIV_KEY = "localhost.key"
app.run(threaded=True ,ssl_context=(PUB_KEY, PRIV_KEY))
#app.run(threaded=True)# ,ssl_context=(PUB_KEY, PRIV_KEY))
