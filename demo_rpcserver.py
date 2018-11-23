#!/usr/bin/python3
# coding: utf-8
from __future__ import print_function
from flask import Flask, Response, request
import json
import sys
import random

app = Flask(__name__)

@app.route("/", methods=["POST","GET"])
def index():
    print("api call hit")
    if request.method == 'POST':
        req_json = request.get_json()
        if req_json and req_json.get('method', None) == 'random':
            return Response(json.dumps({
                           'jsonrpc': '2.0',
                           'id': 1,
                           'result': '%d' % random.randint(0, 10000),
                       }), mimetype='application/json')

    return Response(json.dumps({
        'jsonrpc': '2.0',
        'id': 1,
        'result': 'hello world ' + json.dumps(request.get_json()),
    }), mimetype='application/json')

if __name__ == '__main__':
    port = 5003
    if len(sys.argv) > 1:
        port = int(sys.argv[1])
    app.run(host = "127.0.0.1", port = port, debug = False)
