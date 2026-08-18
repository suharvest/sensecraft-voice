#!/usr/bin/env python3
"""模拟设备完整生命周期，持续循环，验证 voice-service 的设备纳管链路。

流程：
  1. 首次注册：POST /api/v1/devices/register，带 X-Enrollment-Key，拿到 token 并持久化
  2. 心跳：每 N 秒 POST /api/v1/devices/register，带 Authorization: Bearer <token>
  3. 检测心跳响应里的 asr_config_version 与本地不一致
  4. 拉配置：GET /api/v1/devices/me/asr-config（带 token）
  5. 上报生效结果：随下次心跳带上 asr_config_version / asr_config_error

用法：
  python3 mock_device.py --base-url http://127.0.0.1:3008 \
      --mac AA:BB:CC:DD:EE:01 --enrollment-key CHANGE_ME_ENROLLMENT_KEY \
      --interval 5 --cycles 3
  --cycles 0 表示无限循环。
只依赖标准库。
"""
import argparse
import json
import os
import time
import urllib.error
import urllib.request


def http_json(method, url, body=None, headers=None, timeout=15):
    """返回 (status_code, parsed_or_raw_text)"""
    data = json.dumps(body).encode() if body is not None else None
    req = urllib.request.Request(url, data=data, method=method)
    req.add_header("Content-Type", "application/json")
    for k, v in (headers or {}).items():
        req.add_header(k, v)
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:
            raw = resp.read().decode("utf-8", "replace")
            status = resp.status
    except urllib.error.HTTPError as e:
        raw = e.read().decode("utf-8", "replace")
        status = e.code
    except Exception as e:  # 网络层错误
        return 0, {"error": str(e)}
    try:
        return status, json.loads(raw)
    except ValueError:
        return status, raw


def log(step, status, payload):
    text = json.dumps(payload, ensure_ascii=False) if not isinstance(payload, str) else payload
    if len(text) > 900:
        text = text[:900] + "...(truncated)"
    print("[%s] %-22s HTTP %s  %s" % (time.strftime("%H:%M:%S"), step, status, text), flush=True)


class MockDevice:
    def __init__(self, args):
        self.args = args
        self.token = None
        self.local_asr_version = -1
        self.asr_config = None
        self.pending_report = None  # (version, error_msg)
        self.token_file = args.token_file

        if self.token_file and os.path.exists(self.token_file):
            self.token = open(self.token_file).read().strip() or None

    # ---------- HTTP helpers ----------
    def url(self, path):
        return self.args.base_url.rstrip("/") + path

    def register(self, first=False):
        body = {
            "mac_address": self.args.mac,
            "name": self.args.name,
            "ip_address": "192.168.1.123",
            "version": "mock-1.0.0",
            "cpu_usage_percent": 12.5,
            "memory_used_bytes": 512 * 1024 * 1024,
            "memory_total_bytes": 2048 * 1024 * 1024,
            "disk_used_bytes": 4 * 1024 ** 3,
            "disk_total_bytes": 32 * 1024 ** 3,
            "swap_used_bytes": 0,
            "swap_total_bytes": 1024 * 1024 * 1024,
        }
        # 上报上一轮 ASR 配置生效结果
        if self.pending_report is not None:
            version, err = self.pending_report
            body["asr_config_version"] = version
            body["asr_config_error"] = err

        headers = {}
        if self.token:
            headers["Authorization"] = "Bearer " + self.token
        if first or not self.token:
            headers["X-Enrollment-Key"] = self.args.enrollment_key

        status, payload = http_json("POST", self.url("/api/v1/devices/register"), body, headers)
        log("register(首次)" if first else "heartbeat", status, payload)

        if isinstance(payload, dict):
            result = payload.get("result") or {}
            token = result.get("token")
            if token:
                self.token = token
                print("        -> 拿到 device token: %s...%s (仅首次返回)"
                      % (token[:8], token[-4:]), flush=True)
                if self.token_file:
                    with open(self.token_file, "w") as f:
                        f.write(token)
            if self.pending_report is not None:
                print("        -> 已上报生效结果 version=%s error=%r"
                      % self.pending_report, flush=True)
                self.pending_report = None
            return result
        return {}

    def fetch_asr_config(self):
        headers = {}
        if self.token:
            headers["Authorization"] = "Bearer " + self.token
        status, payload = http_json("GET", self.url("/api/v1/devices/me/asr-config"), None, headers)
        log("fetch asr-config", status, payload)
        if isinstance(payload, dict):
            return payload.get("result")
        return None

    # ---------- 主循环 ----------
    def run(self):
        first = self.token is None
        cycle = 0
        while True:
            cycle += 1
            result = self.register(first=first)
            first = False

            remote_version = result.get("asr_config_version")
            if remote_version is not None and remote_version != self.local_asr_version:
                print("        -> asr_config_version 变化: 本地 %s -> 远端 %s，拉取配置"
                      % (self.local_asr_version, remote_version), flush=True)
                cfg = self.fetch_asr_config()
                if cfg:
                    self.asr_config = cfg
                    self.local_asr_version = remote_version
                    # 模拟调本机 respeaker-service 的 POST /api/v1/asr/config 应用配置
                    applied_err = ""
                    print("        -> 应用配置 code=%s base_url=%s api_key=%s"
                          % (cfg.get("code"),
                             cfg.get("config_json", {}).get("base_url"),
                             "***" if cfg.get("config_json", {}).get("api_key") else "(empty)"),
                          flush=True)
                    self.pending_report = (remote_version, applied_err)
                else:
                    # 未分配服务器等情形：记录错误，下次心跳上报
                    self.local_asr_version = remote_version
                    self.pending_report = (remote_version, "fetch asr-config failed")

            if self.args.cycles and cycle >= self.args.cycles:
                print("完成 %d 轮，退出" % cycle, flush=True)
                return
            time.sleep(self.args.interval)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--base-url", default="http://127.0.0.1:3008")
    ap.add_argument("--mac", default="AA:BB:CC:DD:EE:01")
    ap.add_argument("--name", default="mock-device")
    ap.add_argument("--enrollment-key", default="")
    ap.add_argument("--interval", type=float, default=5.0, help="心跳间隔（秒）")
    ap.add_argument("--cycles", type=int, default=0, help="循环轮数，0 = 无限")
    ap.add_argument("--token-file", default="", help="持久化 token 的文件路径")
    args = ap.parse_args()
    MockDevice(args).run()


if __name__ == "__main__":
    main()
