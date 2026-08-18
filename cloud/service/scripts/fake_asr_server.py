#!/usr/bin/env python3
"""被纳管的假 ASR 服务器（OVS 契约子集），用于验证 voice-service 的探测与能力回填。

提供：
  GET /readyz            -> 200；--busy 时 503 {"reasons":["sessions_full"]}（忙但健康）
                            --down 时 503 {"reasons":["backend_not_ready"]}（真故障）
  GET /health /livez     -> 200
  GET /asr/capabilities  -> {"backend","capabilities","sample_rate"}，需要 Bearer api key（如果设了 --api-key）

用法：
  python3 fake_asr_server.py --port 18621 --api-key ovs-test-key
"""
import argparse
import json
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

ARGS = None


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _send(self, code, payload, ctype="application/json"):
        body = payload if isinstance(payload, bytes) else json.dumps(payload).encode()
        self.send_response(code)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        path = self.path.split("?")[0]

        if path in ("/health", "/livez"):
            self._send(200, {"status": "ok"})
            return

        if path == "/readyz":
            if ARGS.down:
                self._send(503, {"status": "not_ready",
                                 "reasons": ARGS.reasons.split(",")})
            elif ARGS.busy:
                # RK / orin-nano 并发上限 1：解码期间 /readyz 稳定 503 sessions_full
                self._send(503, {"status": "not_ready", "reasons": ["sessions_full"]})
            else:
                self._send(200, {"status": "ready"})
            return

        if path == "/asr/capabilities":
            if ARGS.api_key:
                auth = self.headers.get("Authorization", "")
                if auth != "Bearer " + ARGS.api_key:
                    self._send(401, {"error": "unauthorized"})
                    return
            self._send(200, {
                "backend": ARGS.backend,
                "capabilities": ARGS.capabilities.split(","),
                "sample_rate": ARGS.sample_rate,
            })
            return

        self._send(404, {"error": "not_found"})

    def log_message(self, fmt, *args):
        print("[fake-asr] " + fmt % args, flush=True)


def main():
    global ARGS
    ap = argparse.ArgumentParser()
    ap.add_argument("--host", default="127.0.0.1")
    ap.add_argument("--port", type=int, default=18621)
    ap.add_argument("--api-key", default="")
    ap.add_argument("--backend", default="fake")
    ap.add_argument("--capabilities", default="offline,multi_language")
    ap.add_argument("--sample-rate", type=int, default=16000)
    ap.add_argument("--busy", action="store_true",
                    help="/readyz 返回 503 sessions_full（应判定为 busy，不计失败）")
    ap.add_argument("--down", action="store_true",
                    help="/readyz 返回 503 且带 --reasons 的故障原因，用于验证连续失败阈值")
    ap.add_argument("--reasons", default="backend_not_ready",
                    help="--down 时返回的 reasons，逗号分隔")
    ARGS = ap.parse_args()

    srv = ThreadingHTTPServer((ARGS.host, ARGS.port), Handler)
    print("fake ASR server on http://%s:%d (api_key=%s)" % (
        ARGS.host, ARGS.port, "set" if ARGS.api_key else "none"), flush=True)
    srv.serve_forever()


if __name__ == "__main__":
    main()
