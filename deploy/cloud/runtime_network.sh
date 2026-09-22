#!/bin/sh
set -eu
# Docker's published ports bypass UFW; enforce tenant egress in DOCKER-USER.
docker network inspect slang-programs >/dev/null 2>&1 || docker network create \
  --subnet 172.30.0.0/24 --opt com.docker.network.bridge.name=slangprog0 \
  --opt com.docker.network.bridge.enable_icc=false slang-programs
/bin/sh /srv/slang/runtime/runtime_firewall.sh
