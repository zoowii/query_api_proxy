#!/usr/bin/python3
# coding: utf-8
from __future__ import print_function
from flask import Flask, Response, request
import json
import sys

app = Flask(__name__)

@app.route("/", methods=["POST","GET"])
def index():
    print("api call hit")
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
