#!/usr/bin/env python3
import hmac
import json
import os
import subprocess
import time
from base64 import b64encode
from http import HTTPStatus
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from urllib.parse import parse_qs, quote_plus, urlsplit


HOST = os.environ.get("IMAIPLAY_WEBHOOK_HOST", "127.0.0.1")
PORT = int(os.environ.get("IMAIPLAY_WEBHOOK_PORT", "19090"))
PATH = os.environ.get("IMAIPLAY_WEBHOOK_PATH", "/deploy-webhook")
SECRET = os.environ.get("IMAIPLAY_WEBHOOK_SECRET", "")
DEPLOY_SCRIPT = os.environ.get(
    "IMAIPLAY_DEPLOY_SCRIPT", "/usr/local/bin/imaiplay-deploy.sh"
)
MAX_BODY_SIZE = 1024 * 1024


class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):  # noqa: N802 - required by BaseHTTPRequestHandler
        request_url = urlsplit(self.path)
        if request_url.path != PATH:
            self.send_error(HTTPStatus.NOT_FOUND)
            return

        if not SECRET:
            self.send_error(
                HTTPStatus.SERVICE_UNAVAILABLE, "webhook secret is not configured"
            )
            return

        try:
            length = int(self.headers.get("Content-Length", "0"))
        except ValueError:
            self.send_error(HTTPStatus.BAD_REQUEST, "invalid content length")
            return

        if length <= 0 or length > MAX_BODY_SIZE:
            self.send_error(HTTPStatus.REQUEST_ENTITY_TOO_LARGE)
            return

        body = self.rfile.read(length)
        try:
            payload = json.loads(body.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError):
            self.send_error(HTTPStatus.BAD_REQUEST, "invalid json")
            return

        supplied_token = self.headers.get("X-Gitee-Token", "")
        timestamp = str(
            self.headers.get("X-Gitee-Timestamp")
            or payload.get("timestamp")
            or parse_qs(request_url.query).get("timestamp", [""])[0]
        )
        password = str(payload.get("password", ""))
        signature = quote_plus(
            b64encode(
                hmac.new(
                    SECRET.encode("utf-8"),
                    f"{timestamp}\n{SECRET}".encode("utf-8"),
                    "sha256",
                ).digest()
            ).decode("ascii")
        ) if timestamp else ""
        token_valid = hmac.compare_digest(supplied_token, SECRET)
        signature_valid = timestamp and hmac.compare_digest(supplied_token, signature)
        password_valid = hmac.compare_digest(password, SECRET)
        try:
            timestamp_valid = abs(time.time() * 1000 - int(timestamp)) <= 3600000
        except ValueError:
            timestamp_valid = False
        if not (password_valid or token_valid or (signature_valid and timestamp_valid)):
            self.send_error(HTTPStatus.UNAUTHORIZED)
            return

        if payload.get("ref") != "refs/heads/main":
            self._respond(HTTPStatus.OK, b"ignored")
            return

        try:
            subprocess.Popen(
                [DEPLOY_SCRIPT],
                stdin=subprocess.DEVNULL,
                stdout=subprocess.DEVNULL,
                stderr=subprocess.DEVNULL,
                start_new_session=True,
            )
        except OSError:
            self.send_error(HTTPStatus.INTERNAL_SERVER_ERROR, "deploy script unavailable")
            return

        self._respond(HTTPStatus.ACCEPTED, b"deployment started")

    def _respond(self, status, body):
        self.send_response(status)
        self.send_header("Content-Type", "text/plain; charset=utf-8")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format, *args):
        return


if __name__ == "__main__":
    ThreadingHTTPServer((HOST, PORT), WebhookHandler).serve_forever()
