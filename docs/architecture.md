# Architecture

## 1. Purpose

A responsive web app for running a foosball tournament. The public sees live scores, fixtures, standings and the knockout bracket without logging in. Admins score matches live from the table and confirm each other's results. A projector mode runs on a big screen.

Design priority is phone, then tablet, then desktop, matching expected use. Every screen works at every size.

## 2. Actors

| Actor | Access | Capabilities |
|---|---|---|
| Public | None. Never identified. | Read everything. Filter fixtures by player or nickname. |
| Admin | Personal 5-digit PIN | Start and score matches, submit and confirm results, correct results, pause / restart / abandon with a second admin, lock the group stage, manage players, manage admins, view the activity log, export. |

All admins have equal privileges. The first admin is seeded from configuration; admins create, reset and deactivate the rest in the app. Players are data, not users.

## 3. Tournament format

Set per tournament, not fixed in code:

| Setting | Meaning | This cup |
|---|---|---|
| `groups` | Number of groups. Players are assigned alternately by registration order. | 2 |
| `rounds` | Times each pair in a group plays. | 2 |
| `qualifiers` | Players per group who advance. `groups × qualifiers` must be a power of two. | 4 |
| `race_to` | Goals to win a match. No draws. | 10 |
| `points_per_win` | Points for a win. A loss scores 0. | 1 |
| match days | Dates the group stage is spread over. | 3 |

Fixtures are generated from these settings: a round-robin schedule per group, balanced so no player exceeds two matches on a day, spread across the match days. A match may be played before its scheduled day.

Ranking: points, goal difference, goals scored, head-to-head record between the tied players, then shared rank.

Knockouts: the first round pairs the best-ranked qualifier of each group against the lowest-ranked qualifier of another, alternating groups so group-mates cannot meet before the final where the format allows. With two groups and four qualifiers that is A1 v B4, B1 v A4, A2 v B3, B2 v A3. Then semi-finals, third place, final. Same race to `race_to`.

## 4. Match lifecycle

```
not_started ─start─▶ live ─(a player reaches race_to)─▶ full time ─submit─▶ provisional ─confirm─▶ confirmed
                      │ ▲                                                        │                     │
                      │ └ resume                                       send back ┘ (to live)           │
                      ├ pause ─▶ paused                                                                │
                      ├ restart (0–0)                                                 correct ◀────────┘
                      └ abandon ─▶ abandoned                                          (to provisional)
```

Enforced server-side:

1. Goals only while `live`, by an admin, never past `race_to`.
2. At `race_to` the scoring admin submits the final score. The match becomes `provisional` and records who submitted.
3. Any other admin confirms. An admin cannot confirm their own submission.
4. The confirming admin may send it back with a note. The match returns to `live` with the score intact.
5. Any admin may correct a confirmed result. It becomes `provisional` and needs confirmation by a different admin.
6. Pause, restart and abandon are requests that take effect when a second admin approves.
7. Live and provisional scores are public immediately. Only `confirmed` results count in standings, seeding and the bracket.
8. Lock group stage is available when every group match is confirmed. It writes the first knockout round. Later rounds fill as results are confirmed.
9. Every action writes an event with the admin id.

## 5. Stack

| Layer | Choice |
|---|---|
| Server | Go standard library. `net/http` routing, `crypto` for hashing and signing, `embed` for the UI. Rules engine, fixture generation, PIN auth, activity log, SSE fan-out, push sending. |
| Dependencies | `modernc.org/sqlite` (database driver), `golang.org/x/crypto` (argon2id). Nothing else. |
| Database | SQLite. One file on a persistent volume. Embedded migrations. |
| UI | React, Vite, TypeScript, Tailwind. Built to static files and embedded in the binary. Service worker for push. |
| Live updates | Server-sent events on one endpoint. Polling every 15 s as fallback. |
| Notifications | Web Push (VAPID). Admins opt in per device after unlocking. |
| Image | Multi-stage Dockerfile, distroless nonroot, `linux/arm64`. |
| Deployment | Single container on the owner's GitOps platform: k3s, Traefik with automatic TLS, Argo CD, secrets operator. |

## 6. Data model

All tables carry `tournament_id`.

| Table | Columns |
|---|---|
| `tournaments` | id, name, group_count, rounds, qualifiers, race_to, points_per_win, match_days (json), group_stage_locked_at, created_at |
| `players` | id, tournament_id, reg_no, full_name, nickname (≤12, unique), avatar_url, group_label, email, department |
| `admins` | id, tournament_id, label, display_name, pin_hash (argon2id), active, failed_attempts, locked_until, created_by, created_at |
| `matches` | id, tournament_id, stage (group / r16 / qf / sf / third / final), group_label, match_no, match_day, scheduled_date, slot_a, slot_b, player_a_id, player_b_id, score_a, score_b, status, started_at, full_time_at, submitted_by, submitted_at, confirmed_by, confirmed_at, scorer_admin_id, send_back_note, version |
| `match_events` | id, match_id, admin_id, type, payload, created_at |
| `admin_events` | id, admin_id, target_admin_id, type (create / reset_pin / deactivate / reactivate), created_at |
| `pending_actions` | id, match_id, action (pause / restart / abandon), requested_by, approved_by, status, created_at, resolved_at |
| `push_subscriptions` | id, admin_id, endpoint, p256dh, auth, created_at, last_error |

`status` ∈ not_started, live, paused, provisional, confirmed, abandoned. `slot_a` / `slot_b` hold seeding labels ("A1",
"W-QF1") so the bracket renders before players are known. `version` increments on every write. Standings are computed in
Go from confirmed group matches on each read, not stored. Email and department are never served publicly.

## 7. Routes

Public pages: `/`, `/fixtures`, `/standings`, `/bracket`, `/tv`.
Admin pages: `/admin` (PIN), `/admin/matches`, `/admin/matches/:id` (live scoring), `/admin/matches/:id/confirm`, `/admin/more`, `/admin/admins`.

```
GET    /api/tournament                players, matches, standings, bracket
GET    /api/events                    SSE stream
POST   /api/admin/login               PIN → cookie
POST   /api/admin/logout
POST   /api/admin/push                register this device for push
DELETE /api/admin/push                unregister
GET    /api/admins                    list (labels, names, active)
POST   /api/admins                    create { label, display_name, pin }
POST   /api/admins/:id/reset-pin      { pin }
POST   /api/admins/:id/deactivate
POST   /api/admins/:id/reactivate
POST   /api/matches/:id/start
POST   /api/matches/:id/goal          { player: "a" | "b" }
POST   /api/matches/:id/take-back
POST   /api/matches/:id/submit
POST   /api/matches/:id/confirm
POST   /api/matches/:id/send-back     { note }
POST   /api/matches/:id/correct       { score_a, score_b }
POST   /api/matches/:id/request       { action: "pause" | "restart" | "abandon" }
POST   /api/actions/:id/approve
POST   /api/actions/:id/reject
POST   /api/tournament/lock-groups
PATCH  /api/players/:id               nickname, avatar_url
GET    /api/export                    csv
GET    /healthz                       process alive
GET    /readyz                        database reachable
```

Every write re-checks §4 and writes an event row. An admin cannot deactivate themselves, and the last active admin cannot be deactivated. Non-API paths serve the UI.

## 8. Auth and security

- PIN checked against an argon2id hash. On success a signed HttpOnly, Secure, SameSite=Lax cookie carries the admin id. Expiry extends to 60 minutes after each request. Lock clears it.
- Three wrong PINs lock the admin for 10 minutes.
- Public routes are read-only. Writes require the cookie and pass the rules engine.
- PINs are never displayed after creation. Resetting issues a new one.
- Nicknames are set by admins only, max 12 characters.

## 9. Notifications

Sent to every admin except the one who caused the event:

| Event | Message |
|---|---|
| Result submitted | "Tap-In 10–4 Big Job needs confirming" |
| Result sent back | "Match #20 sent back: <note>" (to the submitter only) |
| Pause / restart / abandon requested | "Admin 4 wants to pause match #3" |
| Group stage ready to lock | "All group results confirmed" |

Tapping a notification opens the relevant admin screen. Failed endpoints are dropped after three consecutive errors.

## 10. Configuration

Secrets, injected from the platform's secret store:

| Key | Purpose |
|---|---|
| `SEED_ADMIN_PIN_HASH` | argon2id hash for the first admin. Used only when the admins table is empty. |
| `COOKIE_SECRET` | HMAC key for the session cookie. |
| `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY` | Web Push signing keys. |
| `LITESTREAM_*` | Object storage credentials for backups. |

Plain env: `BASE_URL=https://foosballcup.webbies.dev`, `DATABASE_PATH`, `PORT`.

## 11. Deployment

- Two environments on the loom platform (`webb1es/loom`, `workloads/foosballcup/`): staging at
  `test.foosballcup.webbies.dev` (namespace `foosballcup-staging`) and prod at `foosballcup.webbies.dev`
  (`foosballcup-prod`). Each has its own database.
- One pod per environment, `Recreate` strategy, 1 Gi persistent volume mounted at `/data`. SQLite allows one writer.
- Push to `staging`: CI checks, builds `ghcr.io/webb1es/foosballcup:sha-<sha>` and writes that tag to loom's staging
  overlay. Argo CD syncs.
- Promote by rebase-merging `staging` into `main` (fast-forward). CI writes the same tag to the prod overlay without
  rebuilding.
- Deploys cause a few seconds of downtime. Never promote to prod during a live match.
- Backups: Litestream sidecar replicating the database to object storage. Not set up yet.
- Rollback: set the previous tag in the overlay.

## 12. Seed data

`seed/players.csv` and `seed/matches.csv` hold this cup's players and group fixtures, including results already played. Together with `seed/tournament.csv` (name and format settings) they are embedded in the binary and imported on the first start against an empty database, inserting played matches as `confirmed` with a system admin "Import". A new tournament without a fixtures file gets its fixtures generated from the settings in §3.

## 13. Live updates

Write → SQLite → in-process broadcaster → SSE to every open page. Clients refetch `/api/tournament` on reconnect. The projector cycles Live / Standings / Bracket / Results every 20 s and pins Live while a match is on.

## 14. Visual theme

One fixed theme. No dark mode.

| Token | Value | Use |
|---|---|---|
| bg | #EEF0F4 | Page |
| surface | #FFFFFF | Cards, lists, tables |
| surface-2 | #DCE1EA | Row separators, table headers, disabled |
| border | #C5CCD9 | Borders |
| text | #141B2B | Body, admin header bars, active group switch |
| body | #3E4A66 | Secondary text |
| muted | #5B6880 | Labels, full names, metadata |
| faint | #9AA5BB | Placeholders, disabled text, ranks outside qualifying |
| brand | #2B5BC9 | Primary buttons, links, confirmed, qualifying ranks and line, arc motif |
| live | #9BD34A | GOAL button, Live chip, live borders, player A ring, active nav, current day |
| live-tint | #EEF7DF | Live strip background |
| live-text | #3F7A0B | LIVE label on light surfaces, rank 1 |
| on-live | #0B1220 | Text on live fill |
| night | #0A1430 | Live card, player B half, floating nav |
| night-2 | #1D3F8F | Live card player A half |
| night-line | #22335F | Ring tracks on night |
| ring-b | #4C7DF0 | Player B ring and avatar outline |
| on-night | #FFFFFF | Names and leading score on night |
| on-night-muted | #A8B3CC | Full names, trailing score, metadata on night |

Lime means happening now. Blue means confirmed or primary action. Dashed blue means provisional. Navy is reserved for the live match and the nav.

Type: Chakra Petch 700 for scores, ranks, points and short labels (match numbers, NEXT, DAY 2), tabular numerals. Space Grotesk 700 for headings and nicknames. Inter for body. Nickname bold with full name beneath in muted 11 to 12px.

Shape: cards 24px radius with an 8px bottom-left corner, taken from the logo's open curve. The live card uses 28px with the same 8px corner. Inner lists 14px. Pills for chips, day selector, group switch and tabs. 1px borders. Live card 1.5px lime, 3px on projector. Provisional 1.5px dashed brand. No shadows or gradients. Touch targets at least 44px.

Live card: split in two halves, player A on night-2, player B on night, each with avatar, nickname and full name on their own side. Between them sit two concentric progress rings: outer ring player A in live, inner ring player B in ring-b, each filled to score ÷ `race_to` on a night-line track. The score sits inside the rings. A footer strip on night carries the LIVE label, group and match number. Ring diameter: 128px phone, 168px tablet, 220px desktop, 420px projector. Score inside: 34px, 46px, 60px, 120px.

Components: two concentric arcs (brand outer, live inner) sit partly off-screen top-right behind the header on every screen, in place of a watermark. Day selector as pills under the title: past days on text, current day on live, future days outlined. Phone has a floating 4-tab nav on night with the active tab on live; tablet and desktop put the same tabs as pills in the header. Up next rows show an overlapping avatar pair, "A v B" in Space Grotesk, group and match number in muted, and a Chakra Petch tag at right. Standings use a pill group switch on phone and show both groups on tablet and desktop. The qualifying line is a dashed brand rule under the last qualifying place.

Breakpoints: 390 reference. 768 and up moves nav to the header; Home puts the live card full width, then two columns: up next and latest results on the left, both group tables on the right. 1280 and up caps content at 1344px; Home has a left column with the live card, up next as a row of three cards and latest results, and a right column with both group tables.

## 15. Out of scope

Public sign-ups, player accounts, notifications to players, photo upload, running several tournaments at once in the UI.

## 16. Decisions

| Date       | Decision                                                                                                                                  |
|------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| 2026-11-11 | Two actors: public and admin. All admins equal. First admin seeded; the rest managed in the app. Every action logged.                     |
| 2026-11-11 | Tournament format is configuration: groups, rounds, qualifiers, race to, points per win, match days.                                      |
| 2026-11-11 | Race to 10, no draws, win 1 point for this cup. Tiebreak ends with head-to-head, then shared rank.                                        |
| 2026-11-11 | Matches may start before their scheduled day.                                                                                             |
| 2026-11-11 | Live scores public instantly. Final score needs submit by one admin and confirm by another.                                               |
| 2026-11-11 | Any admin can correct. Corrections return to provisional.                                                                                 |
| 2026-11-11 | Pause, restart, abandon need a second admin.                                                                                              |
| 2026-11-11 | Public sees provisional scores, labelled "Provisional". Only confirmed results count.                                                     |
| 2026-11-11 | 60 minute sliding admin session.                                                                                                          |
| 2026-11-11 | Web Push to all admins when something needs their attention.                                                                              |
| 2026-11-11 | Projector mode included.                                                                                                                  |
| 2026-11-11 | Nicknames, max 10 characters, admin-assigned, full name shown beneath.                                                                    |
| 2026-11-11 | Go standard library, SQLite, embedded React UI, SSE. Prod only.                                                                           |
| 2026-11-11 | Fully responsive, designed phone first. Theme: Paper, one fixed theme.                                                                    |
| 2026-11-12 | Theme replaced by Chrome Arc: Paper base, navy split live card with progress rings, arc motif, cut-corner cards, Chakra Petch for scores. |
| 2026-11-12 | Public name "Entelect Zim Foosball Cup" on every user-facing surface. Module, repo, binary and image are `foosballcup`.                   |
| 2026-11-12 | Nickname limit raised to 12 characters to fit the seeded nicknames.                                                                       |
| 2026-11-12 | Seed data imports automatically on first start instead of through an import command.                                                      |
| 2026-11-12 | Staging and prod on loom, replacing prod only. Staging is `test.foosballcup.webbies.dev`.                                                 |
