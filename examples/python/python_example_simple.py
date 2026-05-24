#!/usr/bin/env python3
"""
go-diamond Python Client Example (stdlib only)

This example uses only Python standard library (urllib)
to demonstrate how to use go-diamond configuration center.

Usage:
    python3 python_example_simple.py
"""

import json
import hashlib
import time
from typing import Optional, Callable, List, Dict, Any
from http.client import HTTPConnection, HTTPException
from urllib.parse import urlencode


class DiamondConfigClient:
    """Python client for go-diamond using only stdlib."""

    def __init__(
        self,
        server_url: str = "http://127.0.0.1:8080",
        namespace: str = "default",
        group: str = "DEFAULT_GROUP",
        data_id: str = "app.json",
        poll_interval: int = 2,
    ):
        self.server_url = server_url.rstrip("/").replace("http://", "")
        self.namespace = namespace
        self.group = group
        self.data_id = data_id
        self.poll_interval = poll_interval
        self._cache: Dict[str, str] = {}
        self._md5_cache: Dict[str, str] = {}
        self._listeners: List[Callable[[str], None]] = []
        self._running = False

    def _make_request(self, host: str, port: int, path: str, method: str = "GET",
                      body: Optional[bytes] = None, headers: Optional[Dict] = None) -> tuple:
        """Make HTTP request using stdlib."""
        conn = HTTPConnection(host, port, timeout=30)
        try:
            headers = headers or {}
            if body:
                headers["Content-Type"] = "application/json"
            conn.request(method, path, body=body, headers=headers)
            response = conn.getresponse()
            return response.status, response.read().decode("utf-8")
        except HTTPException as e:
            return 0, str(e)
        finally:
            conn.close()

    def get_config(self) -> Optional[str]:
        """Fetch configuration from server."""
        host, port = self.server_url.split(":")
        port = int(port) if port else 80
        path = f"/api/v1/configs/{self.namespace}/{self.group}/{self.data_id}"

        status, body = self._make_request(host, port, path)

        if status == 200:
            try:
                data = json.loads(body)
                if data.get("code") == 0 and "data" in data:
                    content = data["data"]["content"]
                    md5 = data["data"]["contentMd5"]
                    key = f"{self.namespace}/{self.group}/{self.data_id}"
                    self._cache[key] = content
                    self._md5_cache[key] = md5
                    return content
            except (json.JSONDecodeError, KeyError):
                pass
        return None

    def add_listener(self, callback: Callable[[str], None]) -> None:
        """Add a listener for configuration changes."""
        self._listeners.append(callback)

    def watch_loop(self) -> None:
        """Watch loop for configuration changes."""
        while self._running:
            config = self.get_config()
            if config is not None:
                key = f"{self.namespace}/{self.group}/{self.data_id}"
                old_md5 = self._md5_cache.get(key)
                current_md5 = hashlib.md5(config.encode()).hexdigest()

                if old_md5 and old_md5 != current_md5:
                    for listener in self._listeners:
                        try:
                            listener(config)
                        except Exception as e:
                            print(f"Listener error: {e}")

            time.sleep(self.poll_interval)

    def batch_watch(self, items: List[Dict], timeout: int = 30) -> Dict:
        """Batch watch multiple configurations."""
        host, port = self.server_url.split(":")
        port = int(port) if port else 80
        path = "/api/v1/watch/batch"

        payload = json.dumps({"watchItems": items, "timeout": timeout}).encode("utf-8")
        headers = {"Content-Type": "application/json"}

        status, body = self._make_request(host, port, path, method="POST", body=payload, headers=headers)

        if status == 200:
            try:
                return json.loads(body)
            except json.JSONDecodeError:
                pass
        return {"code": -1, "message": "request failed"}

    def get_cached(self) -> Optional[str]:
        """Get cached configuration."""
        key = f"{self.namespace}/{self.group}/{self.data_id}"
        return self._cache.get(key)


def demo_basic():
    """Basic configuration retrieval demo."""
    print("=" * 50)
    print("Demo: Basic Configuration Retrieval")
    print("=" * 50)

    client = DiamondConfigClient(
        server_url="http://127.0.0.1:8080",
        namespace="default",
        group="DEFAULT_GROUP",
        data_id="app.json",
    )

    config = client.get_config()
    if config:
        print(f"Config content: {config}")
    else:
        print("Config not found or server unavailable")


def demo_cached():
    """Configuration with cache demo."""
    print("\n" + "=" * 50)
    print("Demo: Configuration with Cache")
    print("=" * 50)

    client = DiamondConfigClient(
        server_url="http://127.0.0.1:8080",
        namespace="default",
        group="DEFAULT_GROUP",
        data_id="app.json",
    )

    config = client.get_config()
    print(f"First fetch: {config}")

    cached = client.get_cached()
    print(f"Cached value: {cached}")


def demo_batch():
    """Batch watch demo."""
    print("\n" + "=" * 50)
    print("Demo: Batch Watch Multiple Configs")
    print("=" * 50)

    client = DiamondConfigClient(
        server_url="http://127.0.0.1:8080",
        namespace="default",
        group="DEFAULT_GROUP",
        data_id="app.json",
    )

    items = [
        {"namespace": "default", "group": "DEFAULT_GROUP", "dataId": "app.json", "md5": ""},
        {"namespace": "default", "group": "DEFAULT_GROUP", "dataId": "db.json", "md5": ""},
    ]

    response = client.batch_watch(items, timeout=5)
    print(f"Batch watch response: {response}")


if __name__ == "__main__":
    print("go-diamond Python Client Demo (stdlib only)")
    print("=" * 50)
    print()
    print("Note: Make sure go-diamond server is running on http://127.0.0.1:8080")
    print()

    demo_basic()
    demo_cached()
    demo_batch()