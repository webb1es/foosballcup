# Foosball Cup — Architecture

Status: **locked** 11 Nov 2026. Wireframes and flows: https://claude.ai/artifact/WvZsJHzx6Ht3UqzafPGT5T
Source of current state: `docs/Foosball_Tournament_12_Players_v3.xlsx` (the spreadsheet this replaces).
Live at: `https://fooscup.webbies.dev`

## 1. What it is

A phone-first website for the Entelect Zimbabwe Foosball Cup. Public pages show live scores, fixtures, standings and the knockout bracket with no login. Five admins, each with a personal PIN, score matches live from the table and confirm each other's results. A projector mode runs on the office big screen.

It replaces a protected Excel workbook that one person had to keep up to date.

## 2. Actors

| Actor | Access | Can do |
|---|---|---|
| Public | No login, never identified | View everything. Filter fixtures by player or nickname. |
| Admin (×5) | Personal 5-digit PIN, pre-set by the organiser | Start and score matches live, submit and confirm results, correct results, pause / restart / abandon (with a second admin), lock the group stage, manage players, view the activity log, export. |

There is no player role. Players are data, not users.

## 3. Rules of the cup

- 12 players, two groups of 6, assigned alternately by registration order.
- Group stage: double round robin, 60 matches, three match days (Wed 11 Nov, Fri 13 Nov, Wed 18 Nov). Nobody plays more than 2 matches a day.
- **Race to 10. No draws.** Win = 1 point, loss = 0.
- Ranking: points, then goal difference, then goals scored. Players level on all three share a rank.
- Top 4 per group qualify. Quarter-finals Fri 20 Nov: A1 v B4, B1 v A4, A2 v B3, B2 v A3. Semis, third place and final Fri 4 Dec.
- Knockouts: same race to 10. Winner advances, loser of a semi plays the third-place match.

## 4. Match lifecycle

```
not_started ──start──▶ live ──(a player reaches 10)──▶ full time ──submit──▶ provisional ──confirm──▶ confirmed
                        │ ▲                                                       │                      │
                        │ └── resume                                   send back ─┘ (back to live,       │
                        ├── pause ──▶ paused                                       score editable)       │
                        ├── restart (score to 0–0)                                                       │
                        └── abandon ──▶ abandoned                                       correct ─────────┘
                                                                                   (back to provisional)
```

Rules enforced by the server, not the UI:

1. A goal can only be recorded while `live`, by an admin, and never takes either score past 10.
2. At 10 the match is at full time. No more goals. The scoring admin **submits** the final score → `provisional`. The submitting admin is recorded.
3. **Any other admin** confirms → `confirmed`. An admin cannot confirm their own submission.
4. Confirming admin may instead **send back** with a note → match returns to `live` with the score intact for correction and re-submission.
5. Any admin can **correct** a confirmed result → `provisional` again, needs confirmation by an admin other than the corrector.
6. **Pause, restart, abandon** are requests. They take effect only when a second admin approves.
7. Live and provisional scores are **public immediately**. Only `confirmed` results count in standings, seeding and the bracket.
8. **Lock group stage** is available only when all 60 group matches are `confirmed`. It writes the eight QF player slots from the standings. Semi and final slots fill automatically as knockout results are confirmed.
9. Every one of the above writes an event to the activity log with the admin's id.

## 5. Stack

One Go binary with everything inside it. Deployed through the existing Project Loom platform on the Oracle VM.

| Layer | Choice | Why |
|---|---|---|
| Server | Go, `net/http` (stdlib) or chi | Rules engine, PIN auth, audit log, SSE fan-out. One language, one process, compile-time safety on the rules. |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no C toolchain). One file on a volume. Migrations embedded. | Scale is 12 players and ~70 matches. Same shape as Invoicer on loom, so the deployment pattern already exists. |
| UI | React + Vite + TypeScript + Tailwind, built to static files and compiled into the binary with `embed` | Live scoring and confirm screens want client-side state and instant feedback. Served from the same process and port. |
| Live updates | Server-sent events, one endpoint | Every open page and the projector subscribe. In-memory fan-out is fine with exactly one pod. 15 s polling as fallback. |
| Image | Multi-stage Dockerfile: Node builds the UI, Go builds the binary, final stage distroless `nonroot` (uid 65532), `linux/arm64` | Matches loom's security context and the node's architecture. |
| Platform | Project Loom: k3s, Traefik edge, cert-manager TLS, Argo CD GitOps, Infisical secrets, GHCR images | Already running. Traefik owns 80/443 on the VM, so nothing public can bypass it anyway. |

Design system for the UI is decided in the UI/UX phase.

## 6. Data model

All tables carry `tournament_id` so a future cup reuses the schema.

**tournaments**
`id, name, race_to (10), points_per_win (1), group_stage_locked_at, created_at`

**players**
`id, tournament_id, reg_no, full_name, nickname (≤10 chars, unique per tournament), avatar_url, group ('A'|'B'), email, department`
Email and department are never shown publicly. Nicknames and avatars are set by admins only.

**admins**
`id, tournament_id, label ('Admin 1'…'Admin 5'), display_name, pin_hash, active, failed_attempts, locked_until`
PINs are stored hashed (argon2id). Nobody can read a PIN back. Hashes are seeded from the Infisical secret on first start (see §9).

**matches**
`id, tournament_id, stage ('group'|'qf'|'sf'|'third'|'final'), group, match_no, match_day, scheduled_date, slot_a, slot_b, player_a_id, player_b_id, score_a, score_b, status, started_at, full_time_at, submitted_by, submitted_at, confirmed_by, confirmed_at, scorer_admin_id, send_back_note, version`
`status ∈ not_started | live | paused | provisional | confirmed | abandoned`
`slot_a/slot_b` hold seeding labels for knockouts ("A1", "W-QF1", "L-SF2") so the bracket renders before players are known. `version` is bumped on every write for optimistic concurrency.

**match_events** (the activity log and the live feed)
`id, match_id, admin_id, type, payload jsonb, created_at`
`type ∈ start | goal | take_back | pause | resume | restart | full_time | submit | confirm | send_back | correct | abandon | request | approve | reject`

**pending_actions**
`id, match_id, action ('pause'|'restart'|'abandon'), requested_by, approved_by, status ('open'|'approved'|'rejected'), created_at, resolved_at`

**standings** is a SQL view over `matches where status = 'confirmed'`. Never stored. Rank ties share a rank.

## 7. Pages and routes

Public (no auth):

| Route | Screen |
|---|---|
| `/` | Home: live match, up next, latest results, standings snapshot |
| `/fixtures` | By match day, filter by player or nickname |
| `/standings` | Both groups, qualifying line, tiebreak footnote |
| `/bracket` | QF → SF → Final, third place, champion |
| `/tv` | Projector mode, 1920×1080, no chrome, cycles Live / Standings / Bracket / Results every 20 s |

Admin (PIN cookie required):

| Route | Screen |
|---|---|
| `/admin` | PIN entry |
| `/admin/matches` | Live now · Needs confirming · Not started · Confirmed |
| `/admin/matches/[id]` | Live scoring: Goal / take back / full time / pause / restart / abandon |
| `/admin/matches/[id]/confirm` | Confirm result or send back with a note |
| `/admin/more` | Results to confirm, group stage lock, correct a result, projector link, players, dates, activity log, export |

API (JSON, same binary):

```
GET  /api/tournament            players, matches, standings, bracket in one payload
GET  /api/events                SSE stream: match updates and pending actions
POST /api/admin/login           PIN → cookie        POST /api/admin/logout
POST /api/matches/:id/start     POST /api/matches/:id/goal         { player: 'a'|'b' }
POST /api/matches/:id/take-back POST /api/matches/:id/submit
POST /api/matches/:id/confirm   POST /api/matches/:id/send-back    { note }
POST /api/matches/:id/correct   { score_a, score_b }
POST /api/matches/:id/request   { action: 'pause'|'restart'|'abandon' }
POST /api/actions/:id/approve   POST /api/actions/:id/reject
POST /api/tournament/lock-groups
PATCH /api/players/:id          nickname, avatar_url
GET  /api/export                xlsx of players, fixtures, results
GET  /healthz  (process alive)  GET /readyz  (SQLite ping)
```

Every write route re-checks the rules in §4 server-side and writes a `match_events` row. The React app is served for every non-API path.

## 8. Auth and security

- Admin enters a PIN. Server compares against the hash, issues a signed HttpOnly, Secure, SameSite=Lax cookie carrying `admin_id` and expiry. Every admin request extends expiry to now + 60 min. Lock clears the cookie.
- 3 wrong PINs lock that admin for 10 minutes. Attempts are logged.
- Public routes are read-only. All writes require the cookie and go through the rules engine.
- Public pages show full name, nickname, avatar. Never email or department.
- Nickname: max 10 characters, assigned and changed by admins only.
- Traefik adds the platform's security headers and compression middleware, as Invoicer does.

## 9. Configuration and secrets

Follows loom's convention: Git holds what must match the manifests, Infisical holds values.

Infisical project `fooscup`, one environment `prod`, synced by the operator into Kubernetes Secret `fooscup-env`:

| Key | What |
|---|---|
| `ADMIN_PINS` | Five `label:argon2id-hash` pairs, comma-separated. Seeded into `admins` on first start; re-read on restart to rotate. |
| `COOKIE_SECRET` | HMAC key for the session cookie. |
| `LITESTREAM_*` | Object storage credentials for backups. |

Set in the manifest (not Infisical): `BASE_URL=https://fooscup.webbies.dev`, `DATABASE_PATH=/data/fooscup.db`, `PORT=8080`, `GOMEMLIMIT`.

## 10. Deployment (Project Loom)

Shape copied from `workloads/invoicer`, minus Gotenberg and the sign-in rate-limit ingress. **Prod only**: one overlay, one ApplicationSet row, namespace `fooscup-prod`.

```
fooscup repo                       loom repo                           cluster
push to main ─▶ CI builds UI+Go ─▶ workloads/fooscup/overlays/prod ─▶ Argo CD syncs ─▶ Deployment (1 pod)
                image linux/arm64   kustomization images.newTag         Traefik + TLS    PVC /data (1Gi, local-path)
                ghcr.io/webb1es/    = sha-<sha>                         fooscup.webbies  Infisical → fooscup-env
                fooscup:sha-<sha>                                       .dev
```

- **Deployment, not Rollout.** SQLite allows one writer, the volume is `ReadWriteOnce`, so `replicas: 1`, `strategy: Recreate`. A deploy costs a few seconds of "no available server". Never deploy during a live match.
- **Pod**: distroless nonroot, `runAsUser/fsGroup 65532`, `/data` mounted from PVC `fooscup-data`, `preStop` sleep 5 s, readiness `/readyz`, liveness `/healthz`. Requests 20m CPU / 32Mi, limit 256Mi.
- **Ingress**: host `fooscup.webbies.dev`, `letsencrypt-prod` issuer, `websecure` entrypoint, security-headers and compress middleware. DNS A record on the node's public IP must exist before the cert can issue.
- **CI** (in this repo): on push to `main`, build and push `ghcr.io/webb1es/fooscup:sha-<sha>` for `linux/arm64`, then bump `images.newTag` in loom's prod overlay. The GHCR package must be public or the pod needs an `imagePullSecret`.
- **Backups**: Litestream sidecar replicating `/data/fooscup.db` continuously to the Oracle object storage bucket used for the Postgres backups. Restore test before match day 2.
- **Rollback**: revert the tag in the prod overlay. Argo redeploys the previous image.
- **Config change without a rebuild**: bump the `fooscup.app/redeploy` annotation on the pod template.

Loom changes are made in that repo with its owner's explicit go, per its own rules.

## 11. Seeding from the spreadsheet

A one-off command in the binary (`fooscup import <xlsx>`) reads the workbook and writes players, the 60 group fixtures with dates, the 8 knockout placeholders, and the results already played. Imported results are inserted as `confirmed` with a system admin id "Import" so the audit trail is honest. Run once against the live volume after first deploy. After that the spreadsheet retires. `/api/export` produces a sheet in the same shape for anyone who wants a copy.

## 12. Live updates

- Admin phone → POST → rules check → SQLite write → in-process broadcaster → SSE to every open public page and the projector.
- Clients subscribe to `/api/events`. On disconnect they reconnect with backoff and refetch `/api/tournament`. If SSE is unavailable they poll every 15 s.
- The projector cycles screens on a 20 s timer and takes data from the same stream. If a match is live, the Live screen stays pinned.

## 13. Non-goals for this cup

- Public sign-ups. Players are managed by admins.
- Player accounts, "my matches", notifications to players.
- Staging environment. Add a second overlay later if the cup becomes annual.
- Multiple concurrent tournaments in the UI (the schema supports it, the screens assume one).
- Photo upload UI. Avatars are URLs set by admins; initials fallback.

## 14. Decisions log

| Date | Decision |
|---|---|
| 11 Nov | Two actors only: public and admin. |
| 11 Nov | 5 admins, personal pre-set PINs, everything logged against a PIN. |
| 11 Nov | Race to 10, no draws, win 1 point. |
| 11 Nov | Live scores public instantly. Final score needs submit by one admin and confirm by another. |
| 11 Nov | Any admin can correct; correction goes back to provisional. |
| 11 Nov | Pause / restart / abandon need a second admin. |
| 11 Nov | Public sees provisional scores, tagged. Only confirmed results count. |
| 11 Nov | 60 min sliding admin session. |
| 11 Nov | Projector mode: yes. |
| 11 Nov | Nicknames, max 10 chars, admin-assigned, full name shown underneath. Dummy avatars for now. |
| 11 Nov | Plain sporting language in the UI: Start, Goal, Full time, Submit, Confirm, Send back, Correct, Live, Provisional, Confirmed. |
| 11 Nov | Stack: single Go binary, SQLite, embedded React + Vite UI, SSE. |
| 11 Nov | Hosting: Project Loom on the Oracle VM. Host `fooscup.webbies.dev`. Prod only. Infisical for secrets. |

## 15. Still open (not blocking)

- Exact word for a submitted-but-unconfirmed score. "Provisional" is used for now.
- Nudge to other admins when a result needs confirming: none, push, or WhatsApp link.
- Can a match start before its scheduled day? Default: yes, admins decide.
- Head-to-head as a further tiebreak, or keep shared ranks as the sheet does. Default: shared ranks.

## 16. Next

UI/UX design phase on the same canvas: colour, type, components, states. Then build.
