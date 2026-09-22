"""Run against the gateway (locally or over HTTPS). Creates only private test data."""
import copy
import http.client
import io
import json
from pathlib import Path
import sys
import time
import urllib.parse
import uuid
import zipfile


class Client:
    def __init__(self, base):
        self.base = urllib.parse.urlsplit(base)
        self.cookie = ""

    def request(self, method, path, body=None, kind="application/json", status=200):
        cls = http.client.HTTPSConnection if self.base.scheme == "https" else http.client.HTTPConnection
        conn = cls(self.base.hostname, self.base.port, timeout=35)
        headers = {"Content-Type": kind}
        if self.cookie:
            headers["Cookie"] = self.cookie
        data = body if isinstance(body, bytes) else json.dumps(body).encode() if body is not None else None
        try:
            conn.request(method, path, data, headers)
            r = conn.getresponse()
            payload = r.read()
            assert r.status == status, (method, path, r.status, payload[:500])
            if r.getheader("Set-Cookie"):
                self.cookie = r.getheader("Set-Cookie").split(";")[0]
            return json.loads(payload) if r.getheader("Content-Type", "").startswith("application/json") else payload
        finally:
            conn.close()


def main():
    base = sys.argv[1]
    a, b = Client(base), Client(base)
    # systemd reports started before startup has finished retiring old containers.
    deadline = time.monotonic() + 30
    while True:
        try:
            assert a.request("GET", "/healthz")["status"] == "ok"
            break
        except (OSError, http.client.HTTPException):
            if time.monotonic() >= deadline:
                raise
            time.sleep(1)
    catalog = a.request("GET", "/operator/")["objects"]
    counts = {kind: len([o for o in catalog if o["type"] == kind]) for kind in ("local", "library", "elementary")}
    assert counts["local"] >= 3 and counts["library"] >= 20 and counts["elementary"] >= 30, counts
    names = {o["def"]["meta"]["name"] for o in catalog if o["type"] == "elementary"}
    assert not names.intersection({"HTTP client", "HTTP server", "read file", "DB execute", "send email"})
    print("Catalog and public operator restrictions:", counts)
    example = json.loads((Path(__file__).parent / "examples/364c6631-57f8-4e02-a96c-a8c05906c470.json").read_text())
    saved = copy.deepcopy(example)
    saved["id"] = str(uuid.uuid4())
    saved["meta"]["name"] = "Smoke test " + saved["id"][:8]
    assert a.request("POST", "/operator/def/", saved)["status"] == "success"
    own = a.request("GET", "/operator/")["objects"]
    other = b.request("GET", "/operator/")["objects"]
    assert saved["id"] in {o["def"]["id"] for o in own}
    assert saved["id"] not in {o["def"]["id"] for o in other}
    print("Save/reload and visitor isolation: passed")
    run = a.request("POST", "/run/", {"id": saved["id"], "props": {}, "gens": {}, "stream": True})
    assert run["status"] == "success"
    a.request("POST", run["url"], 21)
    assert a.request("GET", run["url"]) == [42]
    b.request("GET", run["url"], status=404)
    a.request("DELETE", "/run/", {"handle": run["handle"]})
    assert a.request("GET", "/run/")["objects"] == []
    print("Editor-compatible start / input 21 / output 42 / stop: passed")
    modern = a.request("POST", "/run/", {"blueprint": example["id"], "props": {}, "gens": {}})["object"]
    assert a.request("POST", modern["url"], 12) == 24
    a.request("DELETE", modern["url"])
    print("Native daemon API: passed")
    archive = a.request("GET", "/workspace/export")
    with zipfile.ZipFile(io.BytesIO(archive)) as z:
        assert saved["id"] in json.loads(z.read("workspace.slang"))["blueprints"]
    assert b.request("POST", "/share/import", archive, "application/zip")["status"] == "success"
    assert saved["id"] in {o["def"]["id"] for o in b.request("GET", "/operator/")["objects"]}
    print("Workspace export and import into another session: passed")
    if base.startswith("https:"):
        assert b"Connect ideas" in a.request("GET", "/")
        assert b"app-root" in a.request("GET", "/app/")
        assert b"Welcome back" in a.request("GET", "/slang-app/index-1/")
        print("Public HTTPS homepage, editor and help: passed")


if __name__ == "__main__":
    main()
