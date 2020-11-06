#!/usr/bin/env python
# -*- coding: utf-8 -*-

from time import sleep

from flask import Flask, make_response, Response

app = Flask(__name__)

counter = 1


@app.route("/")
def hello():
    global counter
    counter += 1
    resp = make_response(Response(str(counter)), 200)
    resp.headers['1'] = '1'
    resp.headers['2'] = '2'
    resp.headers['3'] = '3'
    resp.headers['4'] = '4'
    resp.headers['5'] = '5'
    return resp


timeout_s = 10


@app.route("/timeout")
def timeout():
    sleep(timeout_s)
    return str("Waited {}s".format(timeout_s))


PUB_KEY = "localhost.crt"
PRIV_KEY = "localhost.key"
app.run(threaded=True, ssl_context=(PUB_KEY, PRIV_KEY))
# app.run(threaded=True)# ,ssl_context=(PUB_KEY, PRIV_KEY))