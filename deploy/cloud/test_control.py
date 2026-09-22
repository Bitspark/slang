"""Integration checks for a fresh private control-plane deployment.

Run inside the Auth image on the private Compose network. It creates disposable
verification accounts and a blueprint; it never prints tokens or recovery codes.
No calls to external services are made.
"""
import json
import secrets
import time
import uuid

import requests
from production import db
from models import User

run = secrets.token_hex(5)
origin = {'Origin': 'https://slang.run'}
password = secrets.token_urlsafe(24)


def call(service, path, method='GET', token=None, expected=200, **kwargs):
    headers = dict(origin, **kwargs.pop('headers', {}))
    if token:
        headers['Authorization'] = 'Bearer ' + token
    response = requests.request(method, f'http://{service}:8080/api/v1/{path}', headers=headers, timeout=30, **kwargs)
    assert response.status_code == expected, f'{service} {method} {path}: expected {expected}, got {response.status_code}'
    return response


def signup(suffix):
    response = call('auth', 'auth/signup', 'POST', expected=201,
        json=dict(username='verify_' + run + suffix, password=password))
    for cookie in response.cookies:
        assert cookie.secure and cookie.path == '/' and cookie.has_nonstandard_attr('HttpOnly')
    assert len(response.cookies) == 2
    return response.json(), response


def eventually(predicate):
    end = time.monotonic() + 20
    while time.monotonic() < end:
        if predicate():
            return
        time.sleep(.3)
    raise AssertionError('Event delivery did not complete within 20 seconds')


first, first_response = signup('a')
second, _ = signup('b')
a, b = first['access_token'], second['access_token']
call('auth', 'auth/user', token=a)
call('meta', 'telstar/stream', token=a, expected=403)
call('meta', 'telstar/stream', expected=401)
call('auth', 'auth/login', 'POST', expected=403, headers={'Origin': 'https://other.example'}, json={})
call('auth', 'auth/login', 'POST', expected=403,
     json=dict(username=first['user']['username'], password='incorrect-password'))

bid = str(uuid.uuid4())
bundle = {'main': bid, 'blueprints': {bid: {'id': bid,
    'meta': {'name': 'Deployment verification'},
    'services': {'main': {'in': {'type': 'string'}, 'out': {'type': 'string'}}},
    'operators': {}, 'connections': {'(': [')']},
    'geometry': {'size': {'width': 440, 'height': 320}}}}}
call('repo', 'bundle', 'POST', token=a, expected=201, json=bundle)
call('repo', f'bundle/{bid}', token=b, expected=404)
call('repo', f'bundle/{bid}', 'PUT', token=b, expected=404, json=bundle)
call('repo', f'bundle/{bid}', 'DELETE', token=b, expected=404)
call('repo', f'bundle/{bid}/clone', 'POST', token=b, expected=403, json={})
call('repo', f'bundle/{bid}', 'PUT', token=a, expected=400, json={'main': str(uuid.uuid4())})
cloned = call('repo', f'bundle/{bid}/clone', 'POST', token=a, json={}).json()
assert call('repo', f'bundle/{bid}', token=a).json()['operator_id'] == bid
assert cloned['operator_id'] != bid and not cloned['library']

with db.connection_context():
    User.update(groups='admin').where(User.username == first['user']['username']).execute()
a = call('auth', 'auth/login', 'POST', json=dict(username=first['user']['username'], password=password)).json()['access_token']
eventually(lambda: any(c['identifier'] == first['user']['identifier'] for c in call('customer', 'customers', token=a).json()))
streams = call('meta', 'telstar/stream', token=a).json()
assert any(s['name'] == 'userSignedUp' and s['length'] >= 2 for s in streams)
assert any(s['name'] == 'userSignedUp' and any(g['group_name'] == 'customer.internal' for g in s['groups']) for s in streams)
call('meta', 'telstar/stream/missing/group/missing', 'DELETE', token=a, expected=404)
response = requests.get('http://narrow:8080/verification', timeout=10)
assert response.status_code == 200
eventually(lambda: any(s['name'] == 'narrowRequestReceived' and s['length'] > 0 for s in call('meta', 'telstar/stream', token=a).json()))
call('repo', f'bundle/{bid}', 'DELETE', token=a, expected=204)
call('repo', f"bundle/{cloned['operator_id']}", 'DELETE', token=a, expected=204)

# Recovery rotates the one-time code and invalidates existing tokens at every API.
old_b = b
recovered = call('auth', 'auth/recover', 'POST', json=dict(username=second['user']['username'],
    recovery_code=second['recovery_code'], password=password + 'new'))
b = recovered.json()['access_token']
assert recovered.json()['recovery_code'] != second['recovery_code']
call('repo', 'bundle', token=old_b, expected=401)
call('auth', 'auth/recover', 'POST', expected=403, json=dict(username=second['user']['username'],
    recovery_code=second['recovery_code'], password=password + 'again'))
cookie = '; '.join(c.name + '=' + c.value for c in recovered.cookies)
refreshed = call('auth', 'auth/token', 'POST', headers={'Cookie': cookie}, json={})
call('auth', 'auth/token', 'DELETE', headers={'Cookie': cookie})
call('repo', 'bundle', token=refreshed.json()['access_token'], expected=401)
print(json.dumps({'result': 'passed', 'checks': ['signup', 'login', 'cookie_flags', 'csrf',
    'repo_ownership', 'clone_isolation', 'bundle_validation', 'meta_admin_only',
    'telstar_streams', 'customer_delivery', 'narrow_events', 'recovery_rotation',
    'refresh', 'cross_service_logout'], 'verification_prefix': 'verify_' + run}))
