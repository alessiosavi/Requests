#!/usr/bin/env python
# -*- coding: utf-8 -*-

from flask import Flask
app = Flask(__name__)

counter = 1

@app.route("/")
def hello():
    global counter
    counter += 1
    return str(counter)


PUB_KEY = "localhost.crt"
PRIV_KEY = "localhost.key"
app.run(threaded=True ,ssl_context=(PUB_KEY, PRIV_KEY))
