#!/usr/bin/python3
from __future__ import print_function
import os
import sys
import subprocess
import json
import time
import requests

def run_demo_rpc_servers(processes_count=3):
    project_dir = os.path.dirname(__file__)
    demo_program = os.path.join(project_dir, "demo_rpcserver.py")
    rpc_childs = []
    for i in range(processes_count):
        child = subprocess.Popen(["python", demo_program, '%d' % (5001+i)])
        rpc_childs.append(child)
        print("created demo rpcserver at port %d" % (5001 + i))
    return rpc_childs

def run_query_api_proxy():
    query_api_proxy = "./query_api_proxy"
    config_file = "./sample.yml"
    child = subprocess.Popen([query_api_proxy, config_file])
    return child

def send_rpc_request(proxy_url, method, *params):
    res = requests.post(proxy_url, json={
        'id': 1,
        'method': method,
        'params': params or [],
    })
    res_json = res.json()
    if res_json.get('error', None) is not None:
        raise Exception(json.dumps(res_json['error']))
    result = res_json.get('result', None)
    return result

def run_tests(proxy_url):
    r = send_rpc_request(proxy_url, "hello", "China")
    print(r)
    r = send_rpc_request(proxy_url, "random", "123")
    print(r)

def main():
    rpc_childs = run_demo_rpc_servers(3)
    try:
        proxy_proc = run_query_api_proxy()
        time.sleep(4)  # wait for childs started
        try:
            proxy_url = "http://127.0.0.1:5000"
            run_tests(proxy_url)
        except Exception as e:
            print(e)
        finally:
            proxy_proc.kill()
    finally:
        for child in rpc_childs:
            child.kill()

if __name__ == '__main__':
    main()
