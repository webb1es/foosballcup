# Roadmap

Each milestone is small enough for half a day to one day, ships something that works on its own, and has a check that proves it. Milestones are done in order unless noted. Rules and data model are in `architecture.md`.

Design decisions (look and feel, layout, components) are settled before milestone 4.

## A. Foundation, home page, deploy

| # | Deliverable | Done when |
|---|---|---|
| 1 | Go module, `/healthz`, `/readyz`, Dockerfile, CI build | The container starts and `/healthz` returns 200 |
| 2 | SQLite, embedded migrations, CSV seed import | A fresh start loads 12 players and all group matches, and a test asserts the counts |
| 3 | Standings engine and `GET /api/tournament` | Table-driven tests cover points, goal difference, goals scored and head-to-head, and the seed standings match a hand calculation |
| 4 | Vite, React and Tailwind shell, embedded, with the home page (`/`) | The home page renders real API data at 375px and 1280px |
| 5 | Deploy to foosballcup.webbies.dev | The public URL serves the home page over TLS |
| 6 | Fixtures and standings pages | Both pages render at 375px and 1280px, and fixtures filter by player or nickname |

## B. Admin access

| # | Deliverable | Done when |
|---|---|---|
| 7 | PIN login with argon2id, signed cookie, 3-strike lockout, seeded first admin | Tests cover a wrong PIN, lockout expiry and a 60-minute sliding session |
| 8 | Admin management (create, reset PIN, deactivate, reactivate) | The last active admin and self-deactivation are both rejected by tests |

## C. Scoring

| # | Deliverable | Done when |
|---|---|---|
| 9 | Start, goal and take-back with the rules engine and `version`, plus the phone scoring screen | Goals past `race_to` and goals on a non-live match are rejected, and every action writes an event row |
| 10 | SSE stream, 15 s polling fallback, public live-match card | A goal shows on a second device within 2 s |
| 11 | Submit, confirm, send-back and correct | A self-confirm is rejected, and only confirmed results change standings |
| 12 | Pause, restart and abandon via `pending_actions` | Each takes effect only after a second admin approves |

## D. Knockouts and projector

| # | Deliverable | Done when |
|---|---|---|
| 13 | Lock group stage, write the first knockout round, and fill later rounds as results are confirmed | Seeding gives A1 v B4, B1 v A4, A2 v B3, B2 v A3, and locking is refused while any group match is unconfirmed |
| 14 | Bracket page and `/tv` projector mode | The projector cycles through live match, standings, bracket and results on a 1080p screen |

## E. Extras

| # | Deliverable | Done when |
|---|---|---|
| 15 | Web Push with VAPID and per-device opt-in | A submitted result reaches the other admins' phones, and endpoints are dropped after three errors |
| 16 | Player edits (nickname, avatar), activity log view, CSV export | The nickname limit and uniqueness are enforced, and the export matches the database |

## Open items

- Due dates per milestone, pending the knockout start date.
- Group matches played before milestone 11 ships are entered by editing `seed/matches.csv`.
