"""Unprivileged public invocation router. Only the control agent uses Docker."""
import re
import time
import contextlib

from flask import Flask, Response, abort, request
from runtime_common import ID, MAX_BODY, invoke, record, connection

app = Flask(__name__)
app.config['MAX_CONTENT_LENGTH'] = MAX_BODY


@app.get('/_certificate')
def certificate():
    # Caddy calls this on loopback; its public sites never proxy this endpoint.
    domain = request.args.get('domain', '')
    match = re.fullmatch(r'([0-9a-f]{32})\.slangapps\.com', domain)
    item = record(match[1]) if match else None
    return ('', 200) if item and item['public'] and item['status'] == 'running' else ('', 404)


@app.route('/', methods=['POST', 'OPTIONS', 'GET'])
def route():
    match = re.fullmatch(r'([0-9a-f]{32})\.slangapps\.com', request.host)
    if not match:
        abort(404)
    item = record(match[1])
    if not item or not item['public'] or item['mode'] != 'httpPost':
        abort(404)
    if request.method == 'OPTIONS':
        response = Response(status=204)
        response.headers.update({'Access-Control-Allow-Origin': '*',
            'Access-Control-Allow-Methods': 'POST', 'Access-Control-Allow-Headers': 'Content-Type'})
        return response
    if request.method == 'GET':
        return {'deployment_id': item['iid'], 'status': item['status'], 'method': 'POST'}
    if not request.is_json:
        abort(415)
    with contextlib.closing(connection('events.db')) as db, db:
        window = int(time.time()) // 60
        db.execute('INSERT INTO invocation_rate(iid,window,count) VALUES(?,?,1) '
            'ON CONFLICT(iid) DO UPDATE SET window=excluded.window,count=CASE '
            'WHEN invocation_rate.window=excluded.window THEN invocation_rate.count+1 ELSE 1 END', (item['iid'], window))
        count = db.execute('SELECT count FROM invocation_rate WHERE iid=?', (item['iid'],)).fetchone()[0]
    if count > 60:
        return {'message': 'Public invocation limit is 60 requests per minute'}, 429, {'Retry-After': '60'}
    status, body = invoke(item, request.get_data())
    response = Response(body, status=status, content_type='application/json')
    response.headers['Access-Control-Allow-Origin'] = '*'
    response.headers['Cache-Control'] = 'no-store'
    response.headers['X-Content-Type-Options'] = 'nosniff'
    return response
