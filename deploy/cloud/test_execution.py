"""Exercise the real private service chain and Go runner through HTTP APIs."""
import json
import os
import secrets
import time
import uuid
import requests

password = secrets.token_urlsafe(24)
prefix = 'runtime_' + secrets.token_hex(5)
headers = {'Origin': 'https://slang.run'}

def call(service, path, method='GET', token=None, expected=200, **kwargs):
    extra = dict(headers)
    if token:
        extra['Authorization'] = 'Bearer ' + token
    result = requests.request(method, f'http://{service}:8080/api/v1/{path}', headers=extra, timeout=70, **kwargs)
    assert result.status_code == expected, f'{service} {path}: expected {expected}, got {result.status_code}'
    return result

def account(suffix):
    return call('auth', 'auth/signup', 'POST', expected=201, json=dict(username=prefix + suffix, password=password)).json()

one, two = account('a'), account('b')
a, b = one['access_token'], two['access_token']
assert len(call('repo', 'bundle', token=a).json()) >= 122
echo = call('repo', 'bundle/00000000-0000-4000-8000-000000000002/clone', 'POST', token=a, json={}).json()
created = call('deployment', 'instance', 'POST', token=a, expected=201, json={'operator_id': echo['operator_id'], 'mode': 'httpPost'}).json()
iid = created['instance_id']
assert created['started'] and len(iid) == 32
call('deployment', f'instance/{iid}', token=b, expected=404)
call('deployment', f'instance/{iid}', 'PATCH', token=b, expected=404, json={'status': 'stop'})
call('deployment', f'instance/{iid}', 'DELETE', token=b, expected=404)
call('deployment', f'instance/{iid}/invoke', 'POST', token=b, expected=404, json='hidden')
assert call('deployment', f'instance/{iid}/invoke', 'POST', token=a, json='real Go execution').json() == 'real Go execution'
assert call('deployment', f'instance/{iid}', token=a).json()['status'] == 'running'
if os.environ.get('TEST_PUBLIC') == '1':
    # Transit must publish the route before on-demand TLS can issue a certificate.
    time.sleep(3)
    result = requests.post(f'https://{iid}.slangapps.com/', json='public HTTPS execution', timeout=45)
    assert result.status_code == 200 and result.json() == 'public HTTPS execution'

deadline = time.monotonic() + 45
while time.monotonic() < deadline:
    preview = requests.get(f"http://vision:8080/api/v1/bundle/{echo['operator_id']}.png", headers=dict(headers, Authorization='Bearer ' + a), timeout=10)
    if preview.status_code == 200:
        assert preview.content.startswith(b'\x89PNG\r\n\x1a\n')
        break
    time.sleep(.5)
else:
    raise AssertionError('Owned preview did not render')
call('vision', f"bundle/{echo['operator_id']}.png", token=b, expected=404)

deadline = time.monotonic() + 30
while time.monotonic() < deadline:
    response = requests.get(f'http://usage:8080/api/v1/instance/{iid}/events', headers=dict(headers, Authorization='Bearer ' + a), timeout=10)
    if response.status_code == 200 and response.json()['events']['executions']:
        break
    time.sleep(.5)
else:
    raise AssertionError('Transit invocation did not reach Usage through Telstar')
call('usage', f'instance/{iid}/events', token=b, expected=404)
call('usage', f'instance/{iid}/logs', token=b, expected=404)
call('deployment', f'instance/{iid}', 'PATCH', token=a, json={'status': 'stop'})
assert call('deployment', f'instance/{iid}', token=a).json()['status'] == 'stopped'
call('deployment', f'instance/{iid}', 'PATCH', token=a, json={'status': 'start'})
call('deployment', f'instance/{iid}', 'PATCH', token=a, json={'status': 'restart'})
assert call('deployment', f'instance/{iid}/invoke', 'POST', token=a, json='after restart').json() == 'after restart'
call('deployment', f'instance/{iid}', 'DELETE', token=a, expected=204)
call('deployment', f'instance/{iid}', token=a, expected=404)

debug = call('deployment', 'instance', 'POST', token=a, expected=201, json={'operator_id': echo['operator_id'], 'mode': 'debug'}).json()['instance_id']
reused = call('deployment', 'instance', 'POST', token=a, expected=201, json={'operator_id': echo['operator_id'], 'mode': 'debug'}).json()['instance_id']
assert debug == reused
call('deployment', f'instance/{debug}', 'DELETE', token=a, expected=204)

call('deployment', 'instance', 'POST', token=a, expected=400, json={'operator_id': echo['operator_id'], 'mode': 'process'})
definition = call('repo', f"bundle/{echo['operator_id']}", token=a).json()['definition']
definition['blueprints'][echo['operator_id']]['services']['main'] = {'in': {'type': 'trigger'}, 'out': {'type': 'trigger'}}
call('repo', f"bundle/{echo['operator_id']}", 'PUT', token=a, expected=204, json=definition)
process = call('deployment', 'instance', 'POST', token=a, expected=201, json={'operator_id': echo['operator_id'], 'mode': 'process'}).json()['instance_id']
time.sleep(3)
assert call('deployment', f'instance/{process}', token=a).json()['status'] == 'running'
deadline = time.monotonic() + 20
while time.monotonic() < deadline:
    logs = call('usage', f'instance/{process}/logs', token=a).json()['logs']
    if any(json.loads(line).get('msg') == 'null' for line in logs):
        break
    time.sleep(.5)
else:
    raise AssertionError('Background process initial trigger did not reach logs')
call('deployment', f'instance/{process}', 'DELETE', token=a, expected=204)
call('repo', f"bundle/{echo['operator_id']}", 'DELETE', token=a, expected=204)
print(json.dumps({'result': 'passed', 'checks': ['122_library_bundles', 'deployment_owner_checks',
    'real_go_echo', 'transit_usage_events', 'usage_owner_checks', 'stop', 'start', 'restart', 'delete',
    'preview_png_and_ownership', 'private_debug_reuse', 'background_trigger_and_logs'],
    'verification_prefix': prefix}))
