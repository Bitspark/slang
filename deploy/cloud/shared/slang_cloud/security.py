"""Same-origin cookie authentication and signed JWTs for separate services.

Only Auth receives the signing key. Every other API receives its public key
and checks the current session generation in Redis, so logout/password changes
take effect across APIs rather than only at the token issuer.
"""
import os
import secrets
from functools import wraps
from pathlib import Path

import redis
from flask import abort, jsonify, request
from flask_jwt_extended import JWTManager, get_jwt, verify_jwt_in_request
from werkzeug.middleware.proxy_fix import ProxyFix


def redis_link():
    return redis.from_url(os.environ['REDIS'])


def configure(app, signing=False):
    app.url_map.strict_slashes = False
    app.wsgi_app = ProxyFix(app.wsgi_app, x_for=1, x_proto=1, x_host=1)
    app.config.update(
        MAX_CONTENT_LENGTH=8 * 1024 * 1024,
        JWT_TOKEN_LOCATION=['headers', 'cookies'],
        JWT_ALGORITHM='RS256',
        JWT_PUBLIC_KEY=Path(os.environ['JWT_PUBLIC_KEY_FILE']).read_text(),
        JWT_IDENTITY_CLAIM='sub',
        JWT_COOKIE_SECURE=os.environ.get('COOKIE_SECURE', 'true') == 'true',
        JWT_COOKIE_SAMESITE='Lax',
        JWT_ACCESS_COOKIE_NAME='__Host-slang_access',
        JWT_REFRESH_COOKIE_NAME='__Host-slang_refresh',
        JWT_ACCESS_COOKIE_PATH='/', JWT_REFRESH_COOKIE_PATH='/',
        JWT_COOKIE_CSRF_PROTECT=False,  # Origin + JSON checks below cover mutations.
        JWT_ERROR_MESSAGE_KEY='message',
    )
    if signing:
        app.config['JWT_PRIVATE_KEY'] = Path(os.environ['JWT_PRIVATE_KEY_FILE']).read_text()
    jwt = JWTManager(app)

    @jwt.token_in_blocklist_loader
    def revoked(header, claims):
        version = redis_link().get('auth:version:' + claims['sub'])
        return not version or not secrets.compare_digest(version.decode(), claims.get('version', ''))

    @app.before_request
    def same_origin():
        if request.method not in ('GET', 'HEAD', 'OPTIONS'):
            origin = request.headers.get('Origin')
            if origin and origin != os.environ.get('PUBLIC_ORIGIN', 'https://slang.run'):
                abort(403, 'Cross-origin request rejected')
            if request.headers.get('Sec-Fetch-Site') == 'cross-site':
                abort(403, 'Cross-site request rejected')
            if request.method in ('POST', 'PUT', 'PATCH') and not request.is_json:
                abort(415, 'Use application/json')

    @app.after_request
    def headers(response):
        response.headers['Cache-Control'] = 'no-store'
        response.headers['X-Content-Type-Options'] = 'nosniff'
        return response

    @app.errorhandler(400)
    @app.errorhandler(403)
    @app.errorhandler(404)
    @app.errorhandler(409)
    @app.errorhandler(413)
    @app.errorhandler(415)
    @app.errorhandler(429)
    def request_error(error):
        return jsonify(message=error.description), error.code

    return jwt


def admin_required(fn):
    @wraps(fn)
    def wrapped(*args, **kwargs):
        verify_jwt_in_request()
        if 'admin' not in get_jwt().get('groups', []):
            abort(403, 'Administrator access required')
        return fn(*args, **kwargs)
    return wrapped


def rate_limit(namespace, key, limit, seconds):
    r = redis_link()
    name = f'rate:{namespace}:{key}'
    # Atomic increment/expiry; an interrupted process cannot leave immortal keys.
    count = r.eval("local n=redis.call('INCR',KEYS[1]); if n==1 then redis.call('EXPIRE',KEYS[1],ARGV[1]) end; return n", 1, name, seconds)
    if count > limit:
        abort(429, 'Too many requests; please try again later')
