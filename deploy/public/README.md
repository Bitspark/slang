# Previous TrySlang playground deployment

This arrangement was deployed on 21–22 September 2026 and replaced later on
22 September when the hosting plan changed. All three tryslang hostnames now
redirect to the product website. The current studio is [slang.run](https://slang.run/);
use the [hosted operations guide](../cloud/README.md) for its backend deployment.
The H3 service is stopped/disabled and its saved state is retained and backed up.
The configuration and commands below describe the previous arrangement.

This configuration runs the released Angular editor against the Go daemon.
The released editor speaks the older save/start/poll/stop API, so `gateway.py`
adapts that API while preserving the native daemon run endpoints.

## Previous production arrangement

- Product website: https://slang.bitspark.com/
- Playground at the time: `https://tryslang.com/app/` (now redirects)
- `tryslang.com/` redirects to the product website; editor/API paths retain their origin.
- The Vue frontend `slang.run` was not yet deployed in this arrangement; it is now the current studio.
- Server: `bn2-space-h3`, Hetzner ID `165375950`, IPv4 `159.69.122.161`
- SSH: the existing `bn2-space-h3` SSH alias, user `bn2`, passwordless sudo
- Files: `/srv/tryslang`; source checkout/archive: `/srv/tryslang/src`
- Services: `caddy`, `docker`, `tryslang` (enabled at boot)
- DNS: Cloudflare zone `b1e7c265c3639dc705fb872c1035c952`; apex and `www`
  A records point to the server, **DNS only**, TTL 120. Preserve mail and other records.
- Product DNS: `slang.bitspark.com` A points to the same server, DNS only, TTL 120.
  Preserve the existing bitspark.com apex and www records.
- TLS: Caddy obtains and renews public Let's Encrypt certificates. HTTP redirects
  to HTTPS, and `www` redirects to the apex.
- Firewall: `tryslang-web` (ID `11658638`) permits TCP 80/443 on this server only.
  The existing shared SSH firewall remains attached and unchanged.

Cloudflare's newer redirect to bitspark.com was disabled. An older Page Rule
still redirects proxied apex traffic to bitspark.de/slang. Cloudflare rejects
account-owned tokens on the legacy Page Rules API (error 1011), even with Page
Rules Write permission. Keep these two records DNS-only unless that legacy
rule is removed with a supported user credential. Cloudflare still provides
authoritative DNS. Do not turn the proxy back on without testing redirects.

The original DNS and redirect snapshots are in the deploying machine's
`~/.codex/deployments/tryslang-20260921/`; they contain no API tokens.
The original Caddy configuration is `/srv/tryslang/backups/Caddyfile.original`.

## Isolation and storage

Each visitor gets a random, HMAC-signed, HttpOnly, Secure, SameSite=Lax cookie.
Each active workspace runs a separate daemon container as UID 65532 with:

- A read-only root filesystem, all capabilities dropped, no new privileges.
- An internal Docker network, no published ports, no cloud credentials or Docker socket.
- Only that visitor's writable workspace and the read-only standard library mounted.
- 192 MiB memory, 0.5 CPU, 64 processes, bounded logs and an 8 MiB temporary filesystem.
- `--public`, which explicitly allowlists computation operators. Network, filesystem,
  SQL, email, shell, and unknown future operators are excluded. The older `--safe`
  flag is insufficient for this use because it permits network and file reads.

The gateway listens on loopback port 5150. Only Caddy is exposed publicly.
It serializes calls within a session, rejects cross-origin browser requests,
caps requests and program outputs, limits each workspace to 100 blueprints / 2 MB
of serialized definitions and four running programs, and allows 12 active workspaces.
Each program response is limited to 256 KiB and its recent output history to
512 KiB / 50 items, so output polling cannot grow gateway memory without bound.
Inactive containers stop after 20 minutes. Saved workspaces survive service and
machine restarts and expire after 30 days of inactivity. Visitors can export and
import ZIP bundles; archives are read in memory, never extracted to disk.

The service user belongs to the Docker group, so treat edits to the gateway as
privileged deployment changes. The daemon containers do not inherit that group.
This is a modest public playground, not an unlimited hosted execution service.

## Install or update

On Ubuntu 24.04, put the repository at `/srv/tryslang/src` and run:

```sh
sudo sh /srv/tryslang/src/deploy/public/install.sh
```

`prepare.py` verifies SHA-256 hashes of UI v0.2.5a and library v0.1.9, installs
the static assets and starter examples, and replaces obsolete in-app help with
a local guide. No Node 8 build is needed. The runtime is built from this checkout.
The installer does not modify DNS, provider firewalls, or existing workspace data.
Updates restart active programs; saved definitions remain on disk.

### Independently released website

Website source now lives in [Bitspark/slang.bitspark.com](https://github.com/Bitspark/slang.bitspark.com).
It consumes the pinned Slang Design package for tokens, native control recipes,
fonts, SVG logos, and type components. The product site, welcome guide, and
Angular presentation adapter share those assets. The editor engine is still the
checksum-verified slang-ui release; the adapter does not replace its behavior.

`website.lock.json` pins the independently built website release and SHA-256.
Update that lock after publishing a reviewed website version. Then back up
`/srv/tryslang/site`, `/srv/tryslang/assets/ui`, and the matching source/config
files, and run:

```sh
sudo python3 /srv/tryslang/src/deploy/public/prepare.py /srv/tryslang --frontend-only
```

This updates static files only. It does not restart programs or touch visitor
state. `--website-archive /path/to/archive.zip` can install a local copy of the
same artifact; checksum verification still applies. `/website-source.json`
reports the installed release. `/design/0.2.1/source.json` reports the upstream
design dependency. Licenses ship inside the versioned design assets.

Keep `tryslang.com/app/`, `/operator/*`, `/run/*`, `/share/*`, and `/workspace/*`
on the existing origin until a deliberate cloud migration. Workspace cookies
are host-specific: redirecting the editor to a different domain would not carry
saved workspaces with it. The product site's call to action therefore continues
to use the working playground until the separate slang.run frontend is ready.

Roll back a frontend-only update by restoring the backed-up site, editor files,
and Caddy configuration (reload Caddy after restoring its configuration).
Do not restore or remove visitor state as part of a presentation rollback.

Before updating, retain the current image and gateway/config files for rollback:

```sh
sudo docker tag tryslang-runtime:current tryslang-runtime:previous
sudo cp /srv/tryslang/gateway.py /srv/tryslang/backups/gateway.previous.py
```

To roll back the application, restore those files, retag the previous image as
`tryslang-runtime:current`, and restart `tryslang`. To restore the former website
behavior, restore the recorded DNS records and redirect rule through Cloudflare.
Do not delete `/srv/tryslang/state` as part of rollback.

## Verification and operations

```sh
go test ./... -timeout 3m
go vet ./...
python3 -m unittest discover -s deploy/public -p 'test_*.py' -v
python3 deploy/public/smoke.py https://tryslang.com
sudo systemctl status caddy tryslang docker
sudo journalctl -u tryslang -u caddy --since '10 minutes ago'
curl -fsS https://tryslang.com/healthz
```

The smoke test checks the catalog, capability filtering, legacy save/run/poll/stop,
native run API, two-visitor isolation, and ZIP export/import. It creates private
test workspaces that expire normally. Also open `/app/` in a browser, select
**Double a number**, click play, enter **21**, click **Send**, verify **42**, and stop.
Check creation, saving, reload, and import/export when changing the UI adapter.

For backups, archive `/srv/tryslang/state` as root with restrictive permissions;
it contains saved user programs and the cookie signing key. Keep Caddy's state
under `/var/lib/caddy` to preserve certificate account and renewal state.
