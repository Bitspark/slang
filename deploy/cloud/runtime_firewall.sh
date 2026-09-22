#!/bin/sh
set -eu
# Also runs as Docker ExecStartPre, before restart policies can launch programs.
iptables -N DOCKER-USER 2>/dev/null || true
iptables -N SLANG-PROGRAMS 2>/dev/null || true
for destination in 0.0.0.0/8 10.0.0.0/8 100.64.0.0/10 127.0.0.0/8 \
  169.254.0.0/16 172.16.0.0/12 192.168.0.0/16 192.0.0.0/24 \
  198.18.0.0/15 224.0.0.0/4 240.0.0.0/4 159.69.123.88/32 159.69.127.87/32; do
  iptables -C SLANG-PROGRAMS -d "$destination" -j REJECT 2>/dev/null || \
    iptables -I SLANG-PROGRAMS 1 -d "$destination" -j REJECT
done
iptables -C DOCKER-USER -i slangprog0 -j SLANG-PROGRAMS 2>/dev/null || \
  iptables -I DOCKER-USER 1 -i slangprog0 -j SLANG-PROGRAMS
iptables -C INPUT -i slangprog0 -j DROP 2>/dev/null || \
  iptables -I INPUT 1 -i slangprog0 -j DROP
# Permit replies to host-originated invocation connections. New tenant-originated
# connections still reach the destination restrictions above.
iptables -C SLANG-PROGRAMS -m conntrack --ctstate ESTABLISHED,RELATED -j RETURN 2>/dev/null || \
  iptables -I SLANG-PROGRAMS 1 -m conntrack --ctstate ESTABLISHED,RELATED -j RETURN
iptables -C INPUT -i slangprog0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT 2>/dev/null || \
  iptables -I INPUT 1 -i slangprog0 -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT
