import os
from http.server import BaseHTTPRequestHandler
from socketserver import TCPServer

from urllib import parse

PORT = int(os.getenv("PORT"))


class Handler(BaseHTTPRequestHandler):
    queries = []

    def do_GET(self):
        try:
            path = parse.urlparse(self.path).path
            path = path.removesuffix("/")
            if path == "/":
                self.handle_root()
            elif path == "/queries":
                self.handle_queries()
        finally:
            self.queries.append(f"GET {self.path}")

    def handle_root(self):
        hostname = os.uname()[1]
        body = f"{hostname}\n".encode("utf-8")
        self.wfile.write(body)
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()

    def handle_queries(self):
        body = "\n".join(self.queries).encode("utf-8")
        self.wfile.write(body)
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()


with TCPServer(("", PORT), Handler) as httpd:
    print(f"Serving HTTP on port {PORT} ...")
    httpd.serve_forever()
