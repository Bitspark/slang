"""Compatibility API for the released editor, with one Docker sandbox per visitor."""
import collections
import hashlib
import hmac
import http.client
import http.cookies
import http.server
import io
import json
import os
from pathlib import Path
import re
import secrets
import shutil
import subprocess
import threading
import time
import urllib.parse
import uuid
import zipfile
from email.parser import BytesParser
from email.policy import default

ROOT = Path(os.environ.get("SLANG_WEB_ROOT", "/srv/tryslang"))
ORIGIN = os.environ.get("SLANG_WEB_ORIGIN", "https://tryslang.com")
STATE = ROOT / "state"
MAX_BODY = 1024 * 1024
MAX_REPLY = 4 * MAX_BODY
MAX_OUTPUT = 256 * 1024
MAX_HISTORY = 512 * 1024
MAX_ACTIVE = 12
IDLE_SECONDS = 20 * 60
RETENTION_SECONDS = 30 * 86400
SID = re.compile(r"^[0-9a-f]{32}$")
HANDLE = re.compile(r"^/run/([0-9a-f]+)/$")
sessions = {}
lock = threading.RLock()
key = b""


def docker(*args):
    return subprocess.check_output(["docker", *args], text=True, stderr=subprocess.PIPE, timeout=30).strip()


def sign(sid):
    return sid + "." + hmac.new(key, sid.encode(), hashlib.sha256).hexdigest()


def verify(value):
    sid, _, signature = value.partition(".")
    return sid if SID.fullmatch(sid) and hmac.compare_digest(sign(sid), value) else None


def compatible_catalog(objects):
    """Hide library blueprints whose dependencies are unavailable in public mode."""
    available = {o["def"]["id"] for o in objects if o["type"] == "elementary"}
    for _ in objects:
        before = len(available)
        for obj in objects:
            bp = obj["def"]
            if obj["type"] == "library" and all(v["operator"] in available for v in bp.get("operators", {}).values()):
                available.add(bp["id"])
        if len(available) == before:
            break
    return [o for o in objects if o["type"] == "local" or o["def"]["id"] in available]


def record_output(history, result):
    history.append(result)
    # Bound bytes as well as item count: large results must not exhaust the host.
    while history and len(json.dumps(list(history)).encode()) > MAX_HISTORY:
        history.popleft()


class Session:
    def __init__(self, sid):
        self.sid = sid
        self.name = "tryslang-session-" + sid
        self.last = time.time()
        self.lock = threading.RLock()
        self.history = {}
        self.legacy = set()
        self.host = None

    def start(self):
        workspace = STATE / "workspaces" / self.sid
        if not workspace.exists():
            shutil.copytree(ROOT / "examples", workspace)
        workspace.touch()
        docker("run", "-d", "--name", self.name, "--label", "app=tryslang-session",
               "--network", "tryslang-sessions", "--read-only", "--cap-drop=ALL",
               "--security-opt=no-new-privileges", "--memory=192m", "--memory-swap=192m",
               "--cpus=0.5", "--pids-limit=64", "--log-opt=max-size=1m", "--log-opt=max-file=1",
               "--tmpfs", "/tmp:rw,noexec,nosuid,size=8m,mode=1777",
               "--mount", f"type=bind,src={workspace},dst=/workspace",
               "--mount", f"type=bind,src={ROOT / 'assets/lib'},dst=/library,readonly",
               "tryslang-runtime:current")
        info = json.loads(docker("inspect", self.name))[0]
        self.host = info["NetworkSettings"]["Networks"]["tryslang-sessions"]["IPAddress"]
        for _ in range(50):
            try:
                self.call("GET", "/run/")
                return
            except OSError:
                time.sleep(.1)
        raise RuntimeError("Workspace did not start")

    def call(self, method, path, body=None):
        conn = http.client.HTTPConnection(self.host, 5149, timeout=20)
        try:
            conn.request(method, path, json.dumps(body).encode() if body is not None else None,
                         {"Content-Type": "application/json"})
            response = conn.getresponse()
            limit = MAX_OUTPUT if path.startswith("/run/") else MAX_REPLY
            data = response.read(limit + 1)
            if len(data) > limit:
                raise ValueError("Program output is too large")
            result = json.loads(data) if data else None
            if response.status >= 400:
                message = (result or {}).get("error", {}).get("msg", "Program could not run")
                raise ValueError(message)
            return result
        finally:
            conn.close()

    def stop(self):
        try:
            docker("rm", "-f", self.name)
        except subprocess.CalledProcessError:
            pass


def get_session(cookie):
    with lock:
        sid = verify(cookie)
        if sid in sessions:
            session = sessions[sid]
            session.last = time.time()
            return session
        if len(sessions) >= MAX_ACTIVE:
            raise RuntimeError("All playground workspaces are busy. Please try again in a few minutes.")
        if not sid:
            if len(list((STATE / "workspaces").iterdir())) >= 1024:
                raise RuntimeError("Workspace storage is full. Please try again later.")
            sid = secrets.token_hex(16)
        session = Session(sid)
        try:
            session.start()
        except Exception:
            session.stop()
            raise
        sessions[sid] = session
        return session


def validate_bundle(body, current_ids):
    if not isinstance(body, dict):
        raise ValueError("Expected a blueprint or bundle")
    if "id" in body:
        body = {"main": body["id"], "blueprints": {body["id"]: body}}
    blueprints = body.get("blueprints")
    if not isinstance(blueprints, dict) or not blueprints:
        raise ValueError("The file contains no blueprints")
    if len(set(current_ids) | set(blueprints)) > 100:
        raise ValueError("A workspace can contain at most 100 blueprints")
    for bp_id, bp in blueprints.items():
        if str(uuid.UUID(bp_id)) != bp_id or bp.get("id") != bp_id:
            raise ValueError("Invalid blueprint ID")
        if not isinstance(bp.get("services"), dict) or "main" not in bp["services"]:
            raise ValueError("A blueprint needs a main service")
        if len(bp.get("operators", {})) > 100 or len(bp.get("connections", {})) > 300:
            raise ValueError("This blueprint is too large for the public playground")
    return body


def unpack_upload(raw, content_type):
    if content_type.startswith("multipart/form-data"):
        msg = BytesParser(policy=default).parsebytes(b"Content-Type: " + content_type.encode() + b"\r\n\r\n" + raw)
        raw = next((part.get_payload(decode=True) for part in msg.iter_parts() if part.get_filename()), b"")
    if raw.startswith(b"PK"):
        with zipfile.ZipFile(io.BytesIO(raw)) as archive:
            if len(archive.infolist()) > 200 or sum(f.file_size for f in archive.infolist()) > MAX_BODY:
                raise ValueError("Archive is too large")
            entries = [f for f in archive.namelist() if f.endswith(".slang")]
            if len(entries) != 1:
                raise ValueError("Upload an exported Slang workspace ZIP or .slang bundle")
            raw = archive.read(entries[0])
    return json.loads(raw)


class Handler(http.server.BaseHTTPRequestHandler):
    def setup(self):
        super().setup()
        self.connection.settimeout(30)

    def log_message(self, fmt, *args):
        pass  # Never log programs, cookies, or request bodies.

    def reply(self, status, value, content_type="application/json", filename=None):
        payload = value if isinstance(value, bytes) else json.dumps(value).encode()
        self.send_response(status)
        self.send_header("Content-Type", content_type)
        self.send_header("Content-Length", str(len(payload)))
        self.send_header("Cache-Control", "no-store")
        if getattr(self, "session_cookie", None):
            self.send_header("Set-Cookie", f"slang_session={self.session_cookie}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age={RETENTION_SECONDS}")
        if filename:
            self.send_header("Content-Disposition", f'attachment; filename="{filename}"')
        self.end_headers()
        self.wfile.write(payload)

    def handle_api(self):
        self.session_cookie = None
        parsed = urllib.parse.urlsplit(self.path)
        path = parsed.path
        if path == "/healthz":
            return self.reply(200, {"status": "ok", "active_workspaces": len(sessions)})
        if self.command not in ("GET", "POST", "DELETE"):
            return self.reply(405, {"error": "Method not allowed"})
        origin = self.headers.get("Origin")
        if origin and origin != ORIGIN:
            return self.reply(403, {"error": "Cross-origin requests are not supported"})
        if self.headers.get("Sec-Fetch-Site") == "cross-site":
            return self.reply(403, {"error": "Cross-site requests are not supported"})
        size = int(self.headers.get("Content-Length", 0))
        if size < 0 or size > MAX_BODY or self.headers.get("Transfer-Encoding"):
            return self.reply(413, {"error": "Upload limit is 1 MB"})
        raw = self.rfile.read(size)
        if len(raw) != size:
            raise ValueError("Incomplete request")
        cookie = http.cookies.SimpleCookie()
        cookie.load(self.headers.get("Cookie", ""))
        session = get_session(cookie["slang_session"].value if "slang_session" in cookie else "")
        self.session_cookie = sign(session.sid)
        with session.lock:
            session.last = time.time()
            (STATE / "workspaces" / session.sid).touch()
            body = json.loads(raw) if raw and path != "/share/import" else None
            method = self.command
            if path == "/operator/" and method == "GET":
                result = session.call("GET", path)
                result["objects"] = compatible_catalog(result["objects"])
                return self.reply(200, result)
            if (path == "/operator/def/" and method == "POST") or (path == "/share/import" and method == "POST"):
                if path == "/share/import":
                    body = unpack_upload(raw, self.headers.get("Content-Type", ""))
                catalog = session.call("GET", "/operator/")["objects"]
                ids = [o["def"]["id"] for o in catalog if o["type"] == "local"]
                reserved = {o["def"]["id"] for o in catalog if o["type"] != "local"}
                if isinstance(body, dict) and isinstance(body.get("blueprints"), dict):
                    body = dict(body, blueprints={k: v for k, v in body["blueprints"].items() if k not in reserved})
                bundle = validate_bundle(body, ids)
                bundle["blueprints"] = {k: v for k, v in bundle["blueprints"].items() if k not in reserved}
                saved = {o["def"]["id"]: o["def"] for o in catalog if o["type"] == "local"}
                saved.update(bundle["blueprints"])
                if len(json.dumps(saved)) > 2 * MAX_BODY:
                    raise ValueError("Workspace storage limit is 2 MB. Export a backup and simplify your programs.")
                session.call("POST", "/operator/def/", bundle)
                return self.reply(200, {"status": "success"})
            if path in ("/share/export", "/workspace/export") and method == "GET":
                catalog = compatible_catalog(session.call("GET", "/operator/")["objects"])
                bps = {o["def"]["id"]: o["def"] for o in catalog if o["type"] != "elementary"}
                local = [o["def"]["id"] for o in catalog if o["type"] == "local"]
                selected = urllib.parse.parse_qs(parsed.query).get("fqop", local[:1])
                main = selected[0] if selected and selected[0] in bps else local[0]
                output = io.BytesIO()
                with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
                    archive.writestr("workspace.slang", json.dumps({"main": main, "blueprints": bps}))
                return self.reply(200, output.getvalue(), "application/zip", "slang-workspace.zip")
            if path == "/run/" and method == "POST":
                if len(session.history) >= 4:
                    raise ValueError("Stop a running program before starting another (maximum 4)")
                legacy = "id" in body
                result = session.call("POST", path, {"blueprint": body.get("blueprint", body.get("id")), "gens": body.get("gens", {}), "props": body.get("props", {})})
                obj = result["object"]
                session.history[obj["handle"]] = collections.deque(maxlen=50)
                if legacy:
                    session.legacy.add(obj["handle"])
                    return self.reply(200, dict(obj, status="success"))
                return self.reply(200, result)
            if path == "/run/" and method == "DELETE":
                handle = body.get("handle", "")
                if handle not in session.history:
                    raise ValueError("Unknown running program")
                session.call("DELETE", "/run/" + handle + "/")
                session.history.pop(handle, None)
                session.legacy.discard(handle)
                return self.reply(200, {"status": "success"})
            match = HANDLE.fullmatch(path)
            if match and match[1] in session.history:
                handle = match[1]
                if method == "GET":
                    return self.reply(200, list(session.history[handle]))
                if method == "POST":
                    result = session.call("POST", path, body)
                    record_output(session.history[handle], result)
                    return self.reply(200, {"status": "success"} if handle in session.legacy else result)
                if method == "DELETE":
                    session.call("DELETE", path)
                    session.history.pop(handle, None)
                    session.legacy.discard(handle)
                    return self.reply(200, {"status": "success"})
            if path == "/run/" and method == "GET":
                return self.reply(200, session.call("GET", path))
            return self.reply(404, {"status": "error", "error": {"code": "NOT_FOUND", "msg": "Unknown endpoint or expired program. Start the program again."}})

    def dispatch(self):
        try:
            self.handle_api()
        except (ValueError, KeyError, TypeError, zipfile.BadZipFile, StopIteration) as exc:
            self.reply(400, {"status": "error", "error": {"code": "INVALID", "msg": str(exc)}})
        except (OSError, http.client.HTTPException):
            # Crashed or timed-out programs must not leave a visitor stuck forever.
            cookie = http.cookies.SimpleCookie(self.headers.get("Cookie", ""))
            sid = verify(cookie["slang_session"].value) if "slang_session" in cookie else None
            with lock:
                if sid in sessions:
                    sessions.pop(sid).stop()
            self.reply(503, {"status": "error", "error": {"code": "RESTART", "msg": "The program exceeded its limits. Reload to restart your workspace; saved programs are preserved."}})
        except Exception:
            self.reply(503, {"status": "error", "error": {"code": "BUSY", "msg": "The workspace is busy or the program exceeded its limits. Please reload in a moment."}})

    do_GET = dispatch
    do_POST = dispatch
    do_DELETE = dispatch


def reap():
    while True:
        time.sleep(30)
        with lock:
            for sid, session in list(sessions.items()):
                if time.time() - session.last > IDLE_SECONDS and session.lock.acquire(blocking=False):
                    try:
                        session.stop()
                        del sessions[sid]
                    finally:
                        session.lock.release()
            for workspace in (STATE / "workspaces").iterdir():
                if SID.fullmatch(workspace.name) and workspace.name not in sessions and time.time() - workspace.stat().st_mtime > RETENTION_SECONDS:
                    shutil.rmtree(workspace)


class Server(http.server.ThreadingHTTPServer):
    daemon_threads = True
    slots = threading.BoundedSemaphore(48)

    def process_request(self, request, address):
        if not self.slots.acquire(blocking=False):
            self.shutdown_request(request)
            return
        super().process_request(request, address)

    def process_request_thread(self, request, address):
        try:
            super().process_request_thread(request, address)
        finally:
            self.slots.release()


def main():
    global key
    STATE.mkdir(exist_ok=True)
    (STATE / "workspaces").mkdir(exist_ok=True)
    keyfile = STATE / "cookie-key"
    if not keyfile.exists():
        keyfile.write_bytes(secrets.token_bytes(32))
        keyfile.chmod(0o600)
    key = keyfile.read_bytes()
    stale = docker("ps", "-aq", "--filter", "label=app=tryslang-session").split()
    if stale:
        docker("rm", "-f", *stale)
    threading.Thread(target=reap, daemon=True).start()
    Server(("127.0.0.1", 5150), Handler).serve_forever()


if __name__ == "__main__":
    main()
