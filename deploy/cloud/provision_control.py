"""Initialize fresh service credentials on the control host; never print secrets.

Run as root after compose.data.yml is healthy. Existing credentials are reused.
This creates databases/roles but never replaces existing service data.
"""
import os
import secrets
import subprocess
from pathlib import Path
from urllib.parse import quote

ROOT = Path('/srv/slang')
SECRET = ROOT / 'secrets'
SERVICES = ('auth', 'repo', 'deployment', 'usage', 'vision', 'customer', 'narrow')


def private_file(path, value, owner=0):
    try:
        fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)
    except FileExistsError:
        return path.read_text()
    with os.fdopen(fd, 'w') as output:
        output.write(value)
    os.chown(path, owner, owner)
    return value


def sql(statement):
    return subprocess.run(['docker', 'exec', '-i', 'slang-postgres-1', 'psql',
                           '-U', 'slang', '-d', 'slang', '-At', '-v', 'ON_ERROR_STOP=1'],
                          input=statement, text=True, capture_output=True, check=True).stdout.strip()


def main():
    os.umask(0o077)
    SECRET.mkdir(exist_ok=True, mode=0o700)
    private = SECRET / 'jwt-private.pem'
    public = SECRET / 'jwt-public.pem'
    if not private.exists():
        subprocess.run(['openssl', 'genpkey', '-algorithm', 'RSA', '-pkeyopt',
                        'rsa_keygen_bits:3072', '-out', str(private)], check=True, capture_output=True)
    subprocess.run(['openssl', 'pkey', '-in', str(private), '-pubout', '-out', str(public)],
                   check=True, capture_output=True)
    os.chown(private, 10001, 10001)
    private.chmod(0o400)
    public.chmod(0o444)
    common = 'REDIS=redis://redis:6379/0\nJWT_PUBLIC_KEY_FILE=/run/keys/jwt-public.pem\nPUBLIC_ORIGIN=https://slang.run\n'
    for service in SERVICES:
        password = private_file(SECRET / f'{service}-database-password', secrets.token_urlsafe(36))
        # All identifiers come from the constant list; passwords use URL-safe alphabet.
        role = 'slang_' + service
        if not sql(f"SELECT 1 FROM pg_roles WHERE rolname='{role}';"):
            sql(f"CREATE ROLE {role} LOGIN PASSWORD '{password}';")
        if not sql(f"SELECT 1 FROM pg_database WHERE datname='{role}';"):
            sql(f'CREATE DATABASE {role} OWNER {role};')
        sql(f'REVOKE CONNECT ON DATABASE {role} FROM PUBLIC; GRANT CONNECT ON DATABASE {role} TO {role};')
        value = common + f'DATABASE=postgresql://{role}:{quote(password)}@postgres:5432/{role}\n'
        value += 'CONSUMER_NAME=' + service + '-1\nEXTERNAL_INTEGRATIONS=false\n'
        if service == 'auth':
            value += 'JWT_PRIVATE_KEY_FILE=/run/signing/jwt-private.pem\n'
        private_file(ROOT / f'{service}.env', value)
    for service in ('meta', 'search'):
        private_file(ROOT / f'{service}.env', common + 'EXTERNAL_INTEGRATIONS=false\n')
    print('Service databases, roles and signing keys are ready; values kept on host.')


if __name__ == '__main__':
    main()
