import http.server
import socketserver
import os

PORT = int(os.getenv("PORT"))

class Handler(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        hostname = os.uname()[1]
        body = f"{hostname}\n".encode("utf-8")
        self.send_response(200)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


with socketserver.TCPServer(("", PORT), Handler) as httpd:
    print(f"Serving HTTP on port {PORT} ...")
    httpd.serve_forever()
