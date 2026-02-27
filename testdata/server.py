import os
from http.server import BaseHTTPRequestHandler
from socketserver import TCPServer

PORT = int(os.getenv("PORT"))


class Handler(BaseHTTPRequestHandler):
    queries = []

    def do_GET(self):
        hostname = os.uname()[1]
        body = f"{hostname}\n".encode("utf-8")
        self.wfile.write(body)
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()


with TCPServer(("", PORT), Handler) as httpd:
    print(f"Serving HTTP on port {PORT} ...")
    httpd.serve_forever()
