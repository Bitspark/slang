# Hosted Slang operations

This directory owns the restored production system. It runs the original service
repositories and Telstar contracts on Docker Compose, with a separate Go execution
host. The local studio adapter and `deploy/public` are separate configurations.
The private [ecosystem overview](https://github.com/Bitspark/slang-ecosystem)
owns cross-repository relationships and historical evidence.

## Hosts and public routes

| Host | Address | Role |
| --- | --- | --- |
| `slang-app` | `159.69.123.88`, private `10.78.1.1` | Caddy, studio, product site, Meta UI, Compose APIs/workers, PostgreSQL 17, Redis 7.2 |
| `slang-runtime` | `159.69.127.87`, private `10.78.1.2` | Caddy, private mTLS agent, unprivileged invocation router, bounded Go containers |

`slang.run` is the account-based studio. `/api/<service>/` strips its prefix and
forwards to the original service's `/api/v1/...` routes. `/admin/meta/` is the
original Meta UI; its API requires an admin account. `slang.bitspark.com` serves
the product site. `tryslang.com`, `www.tryslang.com`, and
`playground.tryslang.com` redirect to the product site.

Public programs use `https://<32-hex-deployment-id>.slangapps.com/`. DNS A records
are DNS-only in Cloudflare; the wildcard points at the runtime. Caddy issues a
certificate only after its loopback admission endpoint confirms a running public
deployment. Studio tests use authenticated same-origin invocation and have no
public route. Background processes require a trigger input, receive one initial
trigger, and can continue from schedules/delegates. They have no HTTP endpoint.

The Hetzner SSH firewall is separate from public ports 80/443. Management uses
dedicated Slang SSH keys and pinned host keys. The former bn2 account and root
SSH login are disabled. Network filters are installed before Docker restarts
programs. Shared cloud-account administration is not a separate IAM boundary.

## Service boundaries

| Component | Listener | Persistent responsibility |
| --- | --- | --- |
| Auth | loopback 8201 | accounts, password/recovery hashes, Telstar outbox |
| Repo | loopback 8202 | blueprints, library/templates, outbox |
| Meta | loopback 8203 | Redis Streams administration |
| Usage | loopback 8204 | owner-scoped events and logs |
| Customer | loopback 8205 | internal account/event records |
| Narrow | loopback 8206 | internal request-event API/outbox |
| Deployment | loopback 8207 | account limits, deployment definitions/outbox |
| Search | loopback 8208 | operator documentation index |
| Vision | loopback 8209 | owner-checked previews |
| Transit | no public listener | lifecycle routing and runtime telemetry to Telstar |
| Runtime agent | private 10.78.1.2:9443 | Docker lifecycle, immutable program snapshots |
| Runtime router | loopback 9010 | public invocation and certificate admission |

Every database service has its own database/role. Auth alone mounts the private
JWT signing key. Other APIs verify its public key and the session generation in
Redis, so logout and password/recovery changes revoke sessions across services.
Cookies are Secure, HttpOnly, host-only and SameSite=Lax. Mutations require JSON
and reject foreign Origin/Fetch-Site headers. A recovery code replaces outbound
email recovery. `auth.bitspark.de/bootstrap_admin.py` creates the initial operator
account; keep its output only in a private controller file.

Customer, Narrow and Usage retain internal events. Airtable, Mailchimp, Segment,
Slack and Sentry integrations are absent from active code; frontend tracking and
the historical support webhook are removed. Historical OAuth registrations,
databases and credential values are not reused. These are fresh accounts/data.

Vision's original embedded editor runs in a separate browser container with no
application secrets or external network. Program containers have no credentials,
host files beyond their own immutable bundle, or Docker socket. They run as uid
65532 with a read-only filesystem, no capabilities, no-new-privileges, 192 MiB,
0.25 CPU, 64 processes, 32 MiB temporary storage and bounded logs. Egress to the
control/runtime hosts, private ranges and metadata is denied; public Internet
access is allowed. This is container isolation on a shared Linux kernel.

## Build and activate

Prerequisites: Ubuntu 24.04, Docker with Compose v2/Buildx, Caddy, Python 3.12 plus
venv, OpenSSL, age, rsync, and the private host network. No Kubernetes is used.
Check out every component at its intended release revision. Preserve a previous
Compose file, image tags, static release directories and manifests for rollback.

1. On the controller, create a JSON mapping from service names to checkouts using
   `components.example.json`. Run `python deploy/cloud/prepare_context.py
   --components components.json --output <new-directory>`. It excludes environment
   files, private keys, legacy Now configuration and dependency trees. It records
   each source revision and refuses to reuse an existing context directory.
2. On a fresh app host, create `/srv/slang/secrets` mode 0700 and a random
   `postgres-password` file mode 0600. Install `compose.data.yml`, start it, wait
   for PostgreSQL/Redis health, and run `provision_control.py` as root. It creates
   service credentials, roles and new signing keys without replacing existing
   data. Never run historical migrations or use old `.env` files.
3. Archive the prepared context, copy it to a new `/srv/slang/build/<release>`
   directory, and run `sh <context>/cloud/build_control.sh <context> <release-tag>`.
   Non-`working` releases require clean source trees. The script builds each
   service, initializes missing tables, activates Compose, installs retention
   timers and records actual image IDs in `/srv/slang/control-manifest.json`.
   Do not reuse a release tag for changed bytes.
4. Build Linux/amd64 `cmd/slang` from the recorded Go source revision. Copy the
   binary and cloud sources to the runtime and run `install_runtime.sh
   <binary> <cloud-directory> <release-tag> <source-revision>` as root. Keep the
   runtime's server private key there. On first installation, copy only its CSR
   to `/tmp/slang-runtime.csr` on the app host and run `provision_runtime_tls.sh`.
   Install the returned CA/server certificates on runtime, then enable/start
   `slang-runtime-agent` and `slang-runtime-router`. Client credentials stay on
   app and are mounted only into Deployment and Transit.
5. Generate standard-library bundles using `go run ./cmd/makebundles -libdir
   <slang-lib/slang> -outdir <bundles>`. Install them at `/srv/slang/library`, then
   run `seed_library.py` in a Repo image with that directory mounted read-only.
   Seeding preserves user records and includes Blank/Echo templates.
6. Build the studio (`npm ci`, typecheck, tests, `npm run build`), Meta frontend
   (`npm ci`, `npm run build` in `frontend`) and the released product website.
   Unpack each into its own `/srv/slang/releases/<component>-<revision>` directory.
   Atomically update `studio`, `meta`, and `website` symlinks. Meta serves the
   repository `index.html` and its built `frontend/main.js`/`main.css`; the website
   release contains `site/`. Record revisions/artifact SHA-256s in a static
   manifest. Never copy `.env` or historical keys into browser artifacts.
7. Validate/install the respective Caddyfiles, update the exact DNS A records,
   and verify real TLS. Preserve mail/DKIM and unrelated DNS records. Allow DNS
   propagation before interpreting an initial ACME challenge failure. The app
   and runtime need ports 80/443; databases and management ports stay private.

Use the combined Compose files for operations:
`docker compose -f /srv/slang/compose.data.yml -f
/srv/slang/compose.control.yml <command>`. Do not remove data volumes or use
`--remove-orphans` with an incomplete file set.

## Verification and limits

`test_control.py` runs inside an Auth image on the Compose network. It exercises
signup/login, cookies, CSRF, ownership, cloning, Meta authorization, Customer and
Narrow events, recovery rotation and cross-service revocation. `test_execution.py`
runs inside an API image and covers actual Go execution, lifecycle, ownership,
preview rendering, debug reuse, background triggers and logs. Set `TEST_PUBLIC=1`
to include real runtime DNS/TLS invocation. These create disposable verification
accounts; retain their printed identifiers for scoped cleanup. Never run broad
deletion against real users. Telstar has its own supervision/Redis regressions;
Meta's retention test must use an isolated Redis database.

Current small-host limits: 200 accounts; 200 active blueprints per account;
1 MiB saved bundle bodies and 8 MiB assembled runtime definitions; 60 blueprint
writes/hour/account and 120/hour globally; 5 deployment records/account and 12
globally, including stopped/debug records. Repeated tests reuse the debug record
for the same blueprint. Public invocation allows 60 requests/minute/deployment,
one concurrent invocation, 2 MiB input/output and a 15-second response timeout.
A timed-out program is stopped by the telemetry worker and can be restarted.
No automatic scale-out or high availability is configured.

Every five minutes, maintenance trims acknowledged Telstar history older than
15 minutes, bounded by **every** group's checkpoint and oldest pending message.
Pending/unconsumed work is never discarded. Telstar deduplication expires after
14 days. Usage keeps at most 31 days, 10,000 events and 1,000 logs per deployment;
the API shows the most recent 200 logs. Saved blueprints/accounts have no timed
deletion policy. Check Redis memory, pending consumers, disk space and systemd
failures when capacity errors appear. No external notification integration is
enabled.

## Backups and recovery

`backup.py` and `slang-backup.timer` take daily age-encrypted snapshots at 03:15
UTC, with up to three minutes of jitter. Each host keeps seven snapshots and
transfers encrypted copies to the other host. The SSH receiving account has
`restrict` and a forced `/usr/bin/rrsync -wo /srv/slang/backups/peer` command;
its dedicated key cannot execute a shell or administer either server. Configure
the host, peer private address and **public** age recipient in
`/srv/slang/backup.json`. Pin the peer SSH host key in `secrets/backup-known-hosts`.
The private age identity stays on the controller, outside Git and both servers.

App backups briefly stop API/workers (not PostgreSQL/Redis), dump all databases,
snapshot Redis and include host-side secrets/configuration/manifests. The initial
verified run took about 45 seconds including stop/start and encryption. Runtime
backups briefly stop its agent/router, snapshot both SQLite databases and copy
immutable instance bundles and TLS keys; program containers continue running.
Expect a short daily API/invocation interruption. These two hosts share a cloud
account/location; peer copies do not protect against losing both hosts/account.
Keep an additional encrypted controller/offsite copy and the age identity.

Run `python deploy/cloud/verify_backup.py --identity <private-age-file>` on the
controller (optionally `--age <binary>`). It decrypts in memory, restores all
PostgreSQL databases and Redis into temporary network-disabled Docker containers,
checks runtime SQLite integrity and every instance snapshot, then removes only
its named test containers. The private identity is never sent to a server.

For a real replacement, provision the hosts/networks, install the recorded
release, stop writers, then restore the app SQL and Redis snapshot into **fresh**
data volumes. Restore matching service credentials and the signing key from the
same backup. Do not overwrite a running database or merge unrelated snapshots.
Restore runtime SQLite, immutable bundles and keys; run `initialize_runtime.py`
to fix state permissions/lock directories. Recreate each previously running
instance through the private agent using its saved owner, mode and bundle, then
call its `route` endpoint for public HTTP instances. Recreate lock files through
`start`, which also discovers each new loopback port. Resume the original Telstar
groups/outboxes and test login, blueprint loading and public invocation. Renew
runtime TLS if the host/private IP or trust boundary changes.

The historical H3 `/srv/tryslang/state` was preserved on retirement, and an
encrypted archive was copied to controller and both Slang hosts. It is not
automatically imported into the new account-based studio.

## Rollback

Restore the prior Compose file and static symlinks, then start the previous
images; keep current data volumes. Restore the prior runtime source/image only
when its state schema remains compatible. Database restoration is a separate
maintenance operation using a tested backup, not a routine code rollback.
Inspect `systemctl --failed`, service/container logs, `/api/<service>/api/__status`,
private mTLS health, pending Telstar messages and the actual public program URL.
Renew the internal server/client certificates before their 825-day expiry;
the private CA has a ten-year lifetime. Caddy renews public certificates.
