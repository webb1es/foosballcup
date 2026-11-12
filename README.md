# Entelect Zim Foosball Cup

Live scores, fixtures, standings and bracket for a foosball tournament, with a projector mode for the big screen. Works on phone, tablet and desktop.

https://foosballcup.webbies.dev

## What it does

- **Public**, no login: the match in progress with the score updating as goals go in, fixtures by day, both group tables, the bracket.
- **Admins**, personal PIN: start a match, tap GOAL as it happens, submit the final score. A second admin confirms it before it counts. Corrections, pauses, restarts and abandons need two admins. Admins get a push notification when something needs their attention. Admins manage players and other admins. Everything is logged.
- **Projector**: full-screen view cycling live match, standings, bracket and results.

## Rules

Groups, round robin, then knockouts. Group count, rounds, qualifiers per group, race-to score and points per win are settings. No draws. Tiebreak: points, goal difference, goals scored, head-to-head.

This cup: two groups of six, double round robin, race to 10, top four qualify.

## Match states

`not_started → live → provisional → confirmed`

Live and provisional scores are public straight away. Only confirmed results count.

## Stack

Single Go binary, standard library, SQLite, embedded React UI, server-sent events, Web Push. Ships as one container.

## Repository

```
docs/architecture.md   rules, data model, API, deployment, theme
seed/tournament.csv    name and format settings
seed/players.csv       players
seed/matches.csv       group fixtures and results
main.go                server entry point
internal/              server, rules, database, standings
web/                   React UI
brand/                 logo files
```

## Run locally

```
cd web && npm install && npm run build && cd ..
go run .
```

Serves the app and API on `localhost:8080`. The first start creates the database and loads `seed/`.

For UI work, keep `go run .` running and start Vite in a second terminal. It reloads on save and forwards `/api` to the
Go server:

```
cd web && npm run dev
```

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `DATABASE_PATH` | `foosballcup.db` | SQLite file. The container uses `/data/foosballcup.db`. |
