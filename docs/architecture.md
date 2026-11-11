# Foosball Cup — Architecture

Status: **locked** 11 Nov 2026. Wireframes and flows: https://claude.ai/artifact/WvZsJHzx6Ht3UqzafPGT5T
Source of current state: `docs/Foosball_Tournament_12_Players_v3.xlsx` (the spreadsheet this replaces).

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

| Layer | Choice | Why |
|---|---|---|
| App | Next.js (App Router), TypeScript | One repo, one deploy, server routes and pages together. |
| UI | Tailwind CSS | Fast to build phone-first. Design system decided in the UI/UX phase. |
| Database | Supabase Postgres (free tier) | Hosted SQL, row-level security, realtime change feed, dashboard the admins can peek at. |
| Realtime | Supabase Realtime on `matches` and `match_events` | Live scores reach phones and the projector in under a second. Public pages also poll every 15 s as a fallback. |
| Hosting | Vercel (free tier), auto-deploy from `main` | Zero ops. |
| Auth | PIN → signed HttpOnly cookie, 60 min sliding expiry | No accounts to manage. See §8. |

Scale is trivial: 12 players, about 70 matches, a few hundred page views a day. Free tiers are more than enough.

## 6. Data model

All tables carry `tournament_id` so a future cup reuses the schema.

**tournaments**
`id, name, race_to (10), points_per_win (1), group_stage_locked_at, created_at`

**players**
`id, tournament_id, reg_no, full_name, nickname (≤10 chars, unique per tournament), avatar_url, group ('A'|'B'), email, department`
Email and department are never shown publicly.

**admins**
`id, tournament_id, label ('Admin 1'…'Admin 5'), display_name, pin_hash, active, failed_attempts, locked_until`
PINs are stored hashed (argon2id). Nobody can read a PIN back.

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

Public (no auth, Supabase anon key, read-only through RLS):

| Route | Screen |
|---|---|
| `/` | Home: live match, up next, latest results, standings snapshot |
| `/fixtures` | By match day, filter by player or nickname |
| `/standings` | Both groups, qualifying line, tiebreak footnote |
| `/bracket` | QF → SF → Final, third place, champion |
| `/tv` | Projector mode, 1920×1080, no chrome, cycles Live / Standings / Bracket / Results every 20 s |

Admin (PIN cookie required, all writes go through server routes using the service role; the browser never writes to the database directly):

| Route | Screen |
|---|---|
| `/admin` | PIN entry |
| `/admin/matches` | Live now · Needs confirming · Not started · Confirmed |
| `/admin/matches/[id]` | Live scoring: Goal / take back / full time / pause / restart / abandon |
| `/admin/matches/[id]/confirm` | Confirm result or send back with a note |
| `/admin/more` | Results to confirm, group stage lock, correct a result, projector link, players, dates, activity log, export |

Server routes (POST unless stated):

```
/api/admin/login            PIN → cookie        /api/admin/logout
/api/matches/:id/start      /api/matches/:id/goal          { player: 'a'|'b' }
/api/matches/:id/take-back  /api/matches/:id/submit
/api/matches/:id/confirm    /api/matches/:id/send-back     { note }
/api/matches/:id/correct    { score_a, score_b }
/api/matches/:id/request    { action: 'pause'|'restart'|'abandon' }
/api/actions/:id/approve    /api/actions/:id/reject
/api/tournament/lock-groups
/api/players/:id            PATCH nickname, avatar
/api/export                 GET  xlsx of players, fixtures, results
```

Every route re-checks the rules in §4 server-side and writes a `match_events` row.

## 8. Auth and security

- Admin enters a PIN. Server compares against the hash, issues a signed HttpOnly, SameSite=Lax cookie carrying `admin_id` and expiry. Every admin request extends expiry to now + 60 min. Lock clears the cookie.
- 3 wrong PINs lock that admin for 10 minutes. Attempts are logged.
- Row-level security: anon role can `select` from `players` (public columns only via a view), `matches`, `match_events`, `standings`. Anon cannot write anything. `admins` is never readable from the client.
- All writes use the service role key, server-side only, after the cookie is validated.
- Public pages show full name, nickname, avatar. Never email or department.
- Nickname: max 10 characters, assigned and changed by admins only.

## 9. Seeding from the spreadsheet

A one-off script (`scripts/import-xlsx.ts`) reads the workbook and writes players, the 60 group fixtures with dates, the 8 knockout placeholders, and the results already played. Imported results are inserted as `confirmed` with a system admin id "Import" so the audit trail is honest. After import the spreadsheet retires. The export route produces a sheet in the same shape for anyone who wants a copy.

## 10. Live updates

- Admin phone → server route → Postgres write → Supabase Realtime broadcast → every open public page and the projector update in place.
- Public pages subscribe to `matches` changes for the current tournament. If the socket drops they fall back to polling every 15 s.
- The projector cycles screens on a 20 s timer and refreshes data on the same realtime feed. If a match is live, the Live screen stays pinned.

## 11. Non-goals for this cup

- Public sign-ups. Players are managed by admins.
- Player accounts, "my matches", notifications to players.
- Multiple concurrent tournaments in the UI (the schema supports it, the screens assume one).
- Photos upload UI. Avatars are URLs set by admins; initials fallback.

## 12. Decisions log

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

## 13. Still open (not blocking)

- Exact word for a submitted-but-unconfirmed score. "Provisional" is used for now.
- Nudge to other admins when a result needs confirming: none, push, or WhatsApp link.
- Can a match start before its scheduled day? Default: yes, admins decide.
- Head-to-head as a further tiebreak, or keep shared ranks as the sheet does. Default: shared ranks.

## 14. Next

UI/UX design phase on the same canvas: colour, type, components, states, using the existing design system or a new one. Then build.
