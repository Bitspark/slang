"""Authenticated private runtime transport shared by Deployment and Transit."""
import os
import requests


class RuntimeErrorResponse(Exception):
    def __init__(self, status, message):
        self.status, self.message = status, message
        super().__init__(message)


def runtime(path, method='GET', **kwargs):
    try:
        response = requests.request(method, os.environ.get('RUNTIME_URL', 'https://10.78.1.2:9443') + path,
            cert=('/run/runtime-tls/client.crt', '/run/runtime-tls/client.key'),
            verify='/run/runtime-tls/ca.crt', timeout=(5, 60), **kwargs)
    except requests.RequestException as error:
        raise RuntimeErrorResponse(503, 'Runtime management is unavailable') from error
    if not response.ok:
        raise RuntimeErrorResponse(response.status_code, 'Runtime rejected the operation')
    return response
