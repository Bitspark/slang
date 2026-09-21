# Public TrySlang deployment

The public site runs the released Angular editor against the current Go daemon.
The released editor speaks the older save/start/poll/stop API, so `gateway.py`
adapts that API while preserving the native daemon run endpoints.

## Production

- Site: https://tryslang.com/ (editor: `/app/`)
- Server: `bn2-space-h3`, Hetzner ID `165375950`, IPv4 `159.69.122.161`
- SSH: the existing `bn2-space-h3` SSH alias, user `bn2`, passwordless sudo
- Files: `/srv/tryslang`; source checkout/archive: `/srv/tryslang/src`
- Services: `caddy`, `docker`, `tryslang` (enabled at boot)
- DNS: Cloudflare zone `b1e7c265c3639dc705fb872c1035c952`; apex and `www`
  A records point to the server, **DNS only**, TTL 120. Preserve mail and other records.
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

### Website branding

The landing-page wordmark and website/editor favicon use the approved SVGs from
[Slang Design v0.2.1](https://github.com/Bitspark/slang-design/tree/v0.2.1/assets/logo).
The dark-surface wordmark uses the richer raspberry and blue palette. The assets
and outlined Roboto lettering's license are vendored in `site/brand/` and copied
by `prepare.py`; the website makes no font request for its logo.

For a branding-only update, back up the current files and copy `site/index.html`
and `site/brand/` into `/srv/tryslang/site/`. Keep the source archive's matching
files current as well. An editor favicon change also updates the icon link in
`/srv/tryslang/assets/ui/index.html`. Caddy serves these static files directly;
no runtime rebuild or restart is needed for a logo change.

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
