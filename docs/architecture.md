# Architecture

Status: locked 2026-11-11. Design canvas: https://claude.ai/artifact/WvZsJHzx6Ht3UqzafPGT5T

## 1. Purpose

A phone-first web app for an office foosball tournament. The public sees live scores, fixtures, standings and the knockout bracket without logging in. Admins score matches live from the table and confirm each other's results. A projector mode runs on a big screen.

## 2. Actors

| Actor | Access | Capabilities |
|---|---|---|
| Public | None. Never identified. | Read everything. Filter fixtures by player or nickname. |
| Admin | Personal 5-digit PIN | Start and score matches, submit and confirm results, correct results, pause / restart / abandon with a second admin, lock the group stage, manage players, view the activity log, export. |

Players are data, not users.

## 3. Rules

- 12 players, two groups of 6, assigned alternately by registration order.
- Group stage: double round robin, 60 matches over three match days.
- Race to 10. No draws. Win 1 point, loss 0.
- Ranking: points, goal difference, goals scored. Players level on all three share a rank.
- Top 4 per group qualify. Quarter-finals: A1 v B4, B1 v A4, A2 v B3, B2 v A3. Then semi-finals, third place, final.

## 4. Match lifecycle

```
not_started ─start─▶ live ─(a player reaches 10)─▶ full time ─submit─▶ provisional ─confirm─▶ confirmed
                      │ ▲                                                   │                     │
                      │ └ resume                                  send back ┘ (to live)           │
                      ├ pause ─▶ paused                                                           │
                      ├ restart (0–0)                                            correct ◀────────┘
                      └ abandon ─▶ abandoned                                     (to provisional)
```

Enforced server-side:

1. Goals only while `live`, by an admin, never past 10.
2. At 10 the scoring admin submits the final score. The match becomes `provisional` and records who submitted.
3. Any other admin confirms. An admin cannot confirm their own submission.
4. The confirming admin may send it back with a note. The match returns to `live` with the score intact.
5. Any admin may correct a confirmed result. It becomes `provisional` and needs confirmation by a different admin.
6. Pause, restart and abandon are requests that take effect when a second admin approves.
7. Live and provisional scores are public immediately. Only `confirmed` results count in standings, seeding and the bracket.
8. Lock group stage is available when all 60 group matches are confirmed. It writes the eight quarter-final slots. Later rounds fill as knockout results are confirmed.
9. Every action writes an event with the admin id.

## 5. Stack

| Layer | Choice |
|---|---|
| Server | Go, `net/http`. Rules engine, PIN auth, activity log, SSE fan-out. |
| Database | SQLite via `modernc.org/sqlite`. One file on a persistent volume. Embedded migrations. |
| UI | React, Vite, TypeScript, Tailwind. Built to static files and embedded in the binary. |
| Live updates | Server-sent events on one endpoint. Polling every 15 s as fallback. |
| Image | Multi-stage Dockerfile, distroless nonroot, `linux/arm64`. |
| Deployment | Single container on the owner's GitOps platform: k3s, Traefik with automatic TLS, Argo CD, secrets operator. |

## 6. Data model

All tables carry `tournament_id`.

| Table | Columns |
|---|---|
| `tournaments` | id, name, race_to, points_per_win, group_stage_locked_at, created_at |
| `players` | id, tournament_id, reg_no, full_name, nickname (≤10, unique), avatar_url, group, email, department |
| `admins` | id, tournament_id, label, display_name, pin_hash (argon2id), active, failed_attempts, locked_until |
| `matches` | id, tournament_id, stage (group / qf / sf / third / final), group, match_no, match_day, scheduled_date, slot_a, slot_b, player_a_id, player_b_id, score_a, score_b, status, started_at, full_time_at, submitted_by, submitted_at, confirmed_by, confirmed_at, scorer_admin_id, send_back_note, version |
| `match_events` | id, match_id, admin_id, type, payload, created_at |
| `pending_actions` | id, match_id, action (pause / restart / abandon), requested_by, approved_by, status, created_at, resolved_at |

`status` ∈ not_started, live, paused, provisional, confirmed, abandoned. `slot_a` / `slot_b` hold seeding labels ("A1", "W-QF1") so the bracket renders before players are known. `version` increments on every write. `standings` is a view over confirmed matches. Email and department are never served publicly.

## 7. Routes

Public pages: `/`, `/fixtures`, `/standings`, `/bracket`, `/tv`.
Admin pages: `/admin` (PIN), `/admin/matches`, `/admin/matches/:id` (live scoring), `/admin/matches/:id/confirm`, `/admin/more`.

```
GET   /api/tournament                 players, matches, standings, bracket
GET   /api/events                     SSE stream
POST  /api/admin/login                PIN → cookie
POST  /api/admin/logout
POST  /api/matches/:id/start
POST  /api/matches/:id/goal           { player: "a" | "b" }
POST  /api/matches/:id/take-back
POST  /api/matches/:id/submit
POST  /api/matches/:id/confirm
POST  /api/matches/:id/send-back      { note }
POST  /api/matches/:id/correct        { score_a, score_b }
POST  /api/matches/:id/request        { action: "pause" | "restart" | "abandon" }
POST  /api/actions/:id/approve
POST  /api/actions/:id/reject
POST  /api/tournament/lock-groups
PATCH /api/players/:id                nickname, avatar_url
GET   /api/export                     xlsx
GET   /healthz                        process alive
GET   /readyz                         database reachable
```

Every write re-checks §4 and writes a `match_events` row. Non-API paths serve the UI.

## 8. Auth and security

- PIN checked against an argon2id hash. On success a signed HttpOnly, Secure, SameSite=Lax cookie carries the admin id. Expiry extends to 60 minutes after each request. Lock clears it.
- Three wrong PINs lock the admin for 10 minutes.
- Public routes are read-only. Writes require the cookie and pass the rules engine.
- Nicknames are set by admins only, max 10 characters.

## 9. Configuration

Secrets, injected from the platform's secret store:

| Key | Purpose |
|---|---|
| `ADMIN_PINS` | Comma-separated `label:argon2id-hash`. Seeded into `admins` on start. |
| `COOKIE_SECRET` | HMAC key for the session cookie. |
| `LITESTREAM_*` | Object storage credentials for backups. |

Plain env: `BASE_URL`, `DATABASE_PATH`, `PORT`.

## 10. Deployment

- One pod, `Recreate` strategy, 1 Gi persistent volume mounted at `/data`. SQLite allows one writer.
- Deploys cause a few seconds of downtime. Never deploy during a live match.
- CI builds the image on push to `main` and bumps the image tag in the platform repo. Argo CD syncs.
- Backups: Litestream sidecar replicating the database continuously to object storage.
- Rollback: revert the image tag.

## 11. Seed data

`seed/players.csv` and `seed/matches.csv` hold the 12 players and 60 group fixtures, including results already played. `fooscup import` loads them once into a fresh database, inserting played matches as `confirmed` with a system admin "Import". Knockout placeholders are generated.

## 12. Live updates

Write → SQLite → in-process broadcaster → SSE to every open page. Clients refetch `/api/tournament` on reconnect. The projector cycles Live / Standings / Bracket / Results every 20 s and pins Live while a match is on.

## 13. Visual theme

One fixed theme. No dark mode.

| Token | Value | Use |
|---|---|---|
| bg | #E6E9EF | Page |
| surface | #F5F7FA | Cards, header, nav |
| surface-2 | #DCE1EA | Confirmed rows, table headers, disabled |
| border | #C5CCD9 | Borders |
| text | #141B2B | Body, admin header bars, score tiles |
| body | #3E4A66 | Secondary text |
| muted | #5B6880 | Labels, full names, metadata |
| faint | #9AA5BB | Placeholders, disabled text |
| brand | #2B5BC9 | Primary buttons, links, confirmed, active nav, qualifying line |
| live | #9BD34A | GOAL button, Live chip, live borders |
| live-tint | #EEF7DF | Live strip background |
| live-text | #3F7A0B | LIVE label, live dot, rank 1 |
| on-live | #0B1220 | Text on live fill |

Lime means happening now. Blue means confirmed or primary action. Dashed blue means provisional.

Type: Space Grotesk 700 for scores, headings and nicknames, tabular numerals. Inter for body. Scores: 52px phone, 76px tablet, 112px desktop, 180px projector. Nickname bold with full name beneath in muted 11 to 12px.

Shape: 12px radius on cards and inputs, 10px on buttons, pills for chips and tabs. 1px borders. Live card 1.5px lime, 3px on projector. Provisional 1.5px dashed brand. No shadows or gradients. Touch targets at least 44px.

Components: public header on surface; admin header on ink. Phone has a 4-tab bottom nav; tablet and desktop use header pill tabs. Match rows show an overlapping avatar pair, nickname line, full names in muted, score at right. Standings tables mark the qualifying line with a dashed rule under 4th place. Logo watermark bottom-right on every screen at 5 to 6 percent opacity.

Breakpoints: 390 reference. 768 and up moves nav to the header and Home to two columns. 1280 and up caps content at 1344px.

## 14. Out of scope

Public sign-ups, player accounts, notifications, staging environment, photo upload, multiple tournaments in the UI.

## 15. Decisions

| Date | Decision |
|---|---|
| 2026-11-11 | Two actors: public and admin. Admins have personal PINs; every action is logged. |
| 2026-11-11 | Race to 10, no draws, win 1 point. |
| 2026-11-11 | Live scores public instantly. Final score needs submit by one admin and confirm by another. |
| 2026-11-11 | Any admin can correct. Corrections return to provisional. |
| 2026-11-11 | Pause, restart, abandon need a second admin. |
| 2026-11-11 | Public sees provisional scores, tagged. Only confirmed results count. |
| 2026-11-11 | 60 minute sliding admin session. |
| 2026-11-11 | Projector mode included. |
| 2026-11-11 | Nicknames, max 10 characters, admin-assigned, full name shown beneath. |
| 2026-11-11 | Single Go binary, SQLite, embedded React UI, SSE. Prod only. |
| 2026-11-11 | Theme: Paper. One fixed theme, no toggle. |

## 16. Open

- Word for a submitted-but-unconfirmed score. Currently "provisional".
- Whether admins get a nudge when a result needs confirming.
- Whether a match may start before its scheduled day. Default: yes.
- Head-to-head as a further tiebreak. Default: shared ranks.
