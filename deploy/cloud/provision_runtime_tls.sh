#!/bin/sh
# Run on slang-app with the runtime's public CSR at /tmp/slang-runtime.csr.
set -eu
umask 077
directory=/srv/slang/secrets/runtime-tls
mkdir -p "$directory"
if [ ! -f "$directory/ca.key" ]; then
  openssl req -x509 -newkey rsa:3072 -nodes -days 3650 \
    -keyout "$directory/ca.key" -out "$directory/ca.crt" \
    -subj '/CN=Slang internal runtime CA' >"$directory/setup.log" 2>&1
fi
printf '%s\n' 'subjectAltName=IP:10.78.1.2' 'extendedKeyUsage=serverAuth' >"$directory/server.ext"
openssl x509 -req -in /tmp/slang-runtime.csr -CA "$directory/ca.crt" -CAkey "$directory/ca.key" \
  -CAcreateserial -days 825 -sha256 -extfile "$directory/server.ext" -out "$directory/server.crt" 2>>"$directory/setup.log"
if [ ! -f "$directory/client.key" ]; then
  openssl req -new -newkey rsa:3072 -nodes -keyout "$directory/client.key" \
    -out "$directory/client.csr" -subj '/CN=slang-control' 2>>"$directory/setup.log"
fi
printf '%s\n' 'extendedKeyUsage=clientAuth' >"$directory/client.ext"
openssl x509 -req -in "$directory/client.csr" -CA "$directory/ca.crt" -CAkey "$directory/ca.key" \
  -CAcreateserial -days 825 -sha256 -extfile "$directory/client.ext" -out "$directory/client.crt" 2>>"$directory/setup.log"
chown 10001:10001 "$directory/client.key"
chmod 400 "$directory/client.key"
chmod 444 "$directory/ca.crt" "$directory/client.crt" "$directory/server.crt"
# Export public certificates only; no private keys leave their generating host.
install -m 644 "$directory/ca.crt" /tmp/slang-runtime-ca.crt
install -m 644 "$directory/server.crt" /tmp/slang-runtime-server.crt
echo 'Private runtime TLS certificates issued; private keys remain on their hosts.'
