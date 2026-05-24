#!/usr/bin/env python3
"""
go-diamond Python Client Example

This example demonstrates how to use go-diamond configuration center
in Python projects. It provides both polling and long-polling modes.

Usage:
    pip install requests
    python python_example.py
"""

import time
import threading
import hashlib
from typing import Optional, Callable, Dict, List, Any
import requests


class DiamondConfigClient:
    """Python client for go-diamond configuration center."""

    def __init__(
        self,
        server_url: str = "http://127.0.0.1:8080",
        namespace: str = "default",
        group: str = "DEFAULT_GROUP",
        data_id: str = "app.json",
        poll_interval: int = 2,
    ):
        """
        Initialize the diamond config client.

        Args:
            server_url: go-diamond server address
            namespace: Configuration namespace (environment)
            group: Configuration group
            data_id: Configuration item identifier
            poll_interval: Polling interval in seconds for watch mode
        """
        self.server_url = server_url.rstrip("/")
        self.namespace = namespace
        self.group = group
        self.data_id = data_id
        self.poll_interval = poll_interval
        self._cache: Dict[str, str] = {}
        self._md5_cache: Dict[str, str] = {}
        self._listeners: List[Callable[[str], None]] = []
        self._running = False
        self._thread: Optional[threading.Thread] = None

    def get_config(self, timeout: int = 30) -> Optional[str]:
        """
        Fetch configuration from server.

        Args:
            timeout: Request timeout in seconds

        Returns:
            Configuration content or None if not found
        """
        url = f"{self.server_url}/api/v1/configs/{self.namespace}/{self.group}/{self.data_id}"
        try:
            response = requests.get(url, timeout=timeout)
            if response.status_code == 200:
                data = response.json()
                if data.get("code") == 0 and "data" in data:
                    content = data["data"]["content"]
                    md5 = data["data"]["contentMd5"]
                    key = f"{self.namespace}/{self.group}/{self.data_id}"
                    self._cache[key] = content
                    self._md5_cache[key] = md5
                    return content
            elif response.status_code == 404:
                return None
        except requests.RequestException as e:
            print(f"Error fetching config: {e}")
        return None

    def add_listener(self, callback: Callable[[str], None]) -> None:
        """
        Add a listener for configuration changes.

        Args:
            callback: Function called when config changes, receives new content
        """
        self._listeners.append(callback)

    def start_watch(self) -> None:
        """
        Start watching for configuration changes using polling.
        This runs in a background thread.
        """
        if self._running:
            return

        self._running = True
        self._thread = threading.Thread(target=self._watch_loop, daemon=True)
        self._thread.start()

    def _watch_loop(self) -> None:
        """Background polling loop for watching changes."""
        while self._running:
            config = self.get_config()
            if config is not None:
                key = f"{self.namespace}/{self.group}/{self.data_id}"
                old_md5 = self._md5_cache.get(key)
                current_md5 = hashlib.md5(config.encode()).hexdigest()

                if old_md5 and old_md5 != current_md5:
                    # Config changed, notify listeners
                    for listener in self._listeners:
                        try:
                            listener(config)
                        except Exception as e:
                            print(f"Listener error: {e}")

            time.sleep(self.poll_interval)

    def stop_watch(self) -> None:
        """Stop watching for configuration changes."""
        self._running = False
        if self._thread:
            self._thread.join(timeout=5)

    def watch_with_md5(self, timeout: int = 30) -> Optional[dict]:
        """
        Long-polling watch with MD5 check.

        Args:
            timeout: Long-poll timeout in seconds

        Returns:
            Changed config data or None if no change
        """
        key = f"{self.namespace}/{self.group}/{self.data_id}"
        current_md5 = self._md5_cache.get(key, "")

        url = f"{self.server_url}/api/v1/watch/{self.namespace}/{self.group}/{self.data_id}"
        params = {"md5": current_md5, "timeout": timeout}

        try:
            response = requests.get(url, params=params, timeout=timeout + 5)
            if response.status_code == 200:
                data = response.json()
                if data.get("code") == 0 and "data" in data:
                    return data["data"]
            elif response.status_code == 304:
                return None  # Not modified
        except requests.RequestException as e:
            print(f"Error in watch: {e}")
        return None

    def get_cached(self) -> Optional[str]:
        """Get cached configuration content."""
        key = f"{self.namespace}/{self.group}/{self.data_id}"
        return self._cache.get(key)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.stop_watch()


def demo_basic_usage():
    """Demo: Basic configuration retrieval."""
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


def demo_with_cache():
    """Demo: Configuration with local cache."""
    print("\n" + "=" * 50)
    print("Demo: Configuration with Cache")
    print("=" * 50)

    client = DiamondConfigClient(
        server_url="http://127.0.0.1:8080",
        namespace="default",
        group="DEFAULT_GROUP",
        data_id="app.json",
    )

    # First call fetches and caches
    config = client.get_config()
    print(f"First fetch: {config}")

    # Second call returns cached value
    cached = client.get_cached()
    print(f"Cached value: {cached}")


def demo_listener():
    """Demo: Configuration change listener."""
    print("\n" + "=" * 50)
    print("Demo: Configuration Change Listener")
    print("=" * 50)

    def on_config_changed(new_content: str):
        print(f"Config changed! New content: {new_content}")

    client = DiamondConfigClient(
        server_url="http://127.0.0.1:8080",
        namespace="default",
        group="DEFAULT_GROUP",
        data_id="app.json",
    )

    client.add_listener(on_config_changed)

    # Fetch initial config
    config = client.get_config()
    print(f"Initial config: {config}")

    # Start watching in background (for demo, just show it works)
    client.start_watch()
    print("Watching for changes... (Ctrl+C to stop)")

    try:
        # Keep main thread alive for a short demo
        time.sleep(10)
    except KeyboardInterrupt:
        pass
    finally:
        client.stop_watch()
        print("Stopped watching")


def demo_batch_watch():
    """Demo: Batch watching multiple configs."""
    print("\n" + "=" * 50)
    print("Demo: Batch Watching Multiple Configs")
    print("=" * 50)

    url = "http://127.0.0.1:8080/api/v1/watch/batch"

    items = [
        {"namespace": "default", "group": "DEFAULT_GROUP", "dataId": "app.json", "md5": ""},
        {"namespace": "default", "group": "DEFAULT_GROUP", "dataId": "db.json", "md5": ""},
        {"namespace": "default", "group": "DEFAULT_GROUP", "dataId": "redis.json", "md5": ""},
    ]

    payload = {"watchItems": items, "timeout": 30}

    try:
        response = requests.post(url, json=payload, timeout=35)
        if response.status_code == 200:
            data = response.json()
            changed = data.get("data", {}).get("changed", [])
            print(f"Changed configs: {changed}")
        else:
            print(f"Batch watch failed: {response.status_code}")
    except requests.RequestException as e:
        print(f"Error: {e}")


def demo_with_context_manager():
    """Demo: Using client as context manager."""
    print("\n" + "=" * 50)
    print("Demo: Using Client as Context Manager")
    print("=" * 50)

    with DiamondConfigClient(
        server_url="http://127.0.0.1:8080",
        namespace="default",
        group="DEFAULT_GROUP",
        data_id="app.json",
    ) as client:
        config = client.get_config()
        print(f"Config via context manager: {config}")


if __name__ == "__main__":
    print("go-diamond Python Client Demo")
    print("=" * 50)
    print()
    print("Note: Make sure go-diamond server is running on http://127.0.0.1:8080")
    print()

    # Run all demos
    demo_basic_usage()
    demo_with_cache()
    demo_with_context_manager()
    demo_batch_watch()

    # Listener demo runs for 10 seconds
    demo_listener()