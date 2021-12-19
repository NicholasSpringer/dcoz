import sys
import http.server
import requests
from threading import Thread
import time
import re

P = 2000
LADDR = '0.0.0.0'
PORT = int(sys.argv[1])

def power(p):
    x = 1
    for _ in range(p):
        x *= p
    return x


def power_times(t):
    print(f'Running power {t} times')
    start_time = time.time()
    for _ in range(t):
        power(P)
    end_time = time.time()
    print(f'Finished in {(end_time - start_time)*1000} ms')


def query(addr):
    requests.get(f'http://{addr}')


power_pattern = re.compile(r'^power:(\d+)$')
req_pattern = re.compile(r'^req:(.+)$')
wait_pattern = re.compile(r'^wait:(\d+)$')

# take input of the form: power:t_1 query:address1 t_2 wait:0, etc
def run(args):
    threads = []
    for arg in args:
        m = power_pattern.match(arg)
        if m != None:
            t = int(m.group(1))
            power_times(t)
            continue
        m = req_pattern.match(arg)
        if m != None:
            addr = m.group(1)
            print(f'Calling {addr}')
            thread = Thread(target=query, args=(addr,))
            thread.start()
            threads.append(thread)
            continue
        m = wait_pattern.match(arg)
        if m != None:
            idx = int(m.group(1))
            print(f'Waiting on {idx}')
            start_time = time.time()
            thread = threads[idx]
            thread.join()
            end_time = time.time()
            print(f'Finished waiting after {(end_time - start_time)*1000} ms')
            continue
        print(f'Error in input token {arg}')

class PowerHandler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        run(sys.argv[2:])
        self.send_response(200)
        self.end_headers()
    def log_message(self, format, *args):
        return

try:
    server = http.server.HTTPServer((LADDR, PORT), PowerHandler)
    print(f'Listening at {LADDR}:{PORT}')
    server.serve_forever()
except KeyboardInterrupt:
    print('^C received, shutting down server')
    server.socket.close()
