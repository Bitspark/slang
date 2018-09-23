import json, sys
from BaseHTTPServer import HTTPServer, BaseHTTPRequestHandler
from {{ .OpFile }} import {{ .Method }}

class OperatorServer(BaseHTTPRequestHandler):
    def _set_headers(self):
        self.send_response(200)
        self.send_header('Content-type', 'text/json')
        self.end_headers()

    def do_POST(self):
        request_body = self.rfile.read(int(self.headers.getheader('Content-Length')))
        response_body = json.dumps({{ .Method }}(json.loads(request_body)))

        self._set_headers()
        self.wfile.write(response_body)
        self.wfile.close()

if __name__ == '__main__':
    httpd = HTTPServer(("", 0), OperatorServer)
    port = httpd.socket.getsockname()[1]
    print "http://{}:{}".format("localhost", port)
    sys.stdout.flush()
    httpd.serve_forever()