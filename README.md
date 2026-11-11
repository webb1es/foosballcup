# Foosball Cup

Website for an office foosball tournament: live scores, fixtures, standings and the knockout bracket, with a projector mode for the big screen.

Live site: https://fooscup.webbies.dev

## What it does

- **Public**, no login: see the match in progress with the score updating as goals go in, today's fixtures, both group tables and the bracket.
- **Admins**, personal PIN: start a match, tap GOAL as it happens, submit the final score at 10. A second admin confirms it before it counts. Corrections, pauses, restarts and abandons all need two admins. Everything is logged.
- **Projector**: a full-screen view that cycles live match, standings, bracket and results.

## Rules

Race to 10, no draws, one point per win. Two groups of six, double round robin, top four qualify. Quarter-finals seeded A1 v B4, B1 v A4, A2 v B3, B2 v A3, then semi-finals, third place and final. Tiebreak: points, goal difference, goals scored.

## Match states

`not_started → live → provisional → confirmed`

Live and provisional scores are visible to everyone straight away. Only confirmed results count in the standings.

## Stack

Single Go binary with SQLite and an embedded React UI. Server-sent events push updates to open pages. Ships as one container.

## Repository

```
docs/architecture.md   rules, data model, API, deployment, visual theme
seed/players.csv       the 12 players
seed/matches.csv       the 60 group fixtures and results so far
```

Application code lands in `server/` (Go) and `web/` (React) as it is built.

## Status

Design locked. Build in progress.
