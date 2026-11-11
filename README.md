# Foosball Cup

Live scores, fixtures, standings and bracket for an office foosball tournament, with a projector mode for the big screen.

https://foosballcup.webbies.dev

## What it does

- **Public**, no login: the match in progress with the score updating as goals go in, fixtures by day, both group tables, the bracket.
- **Admins**, personal PIN: start a match, tap GOAL as it happens, submit the final score at 10. A second admin confirms it before it counts. Corrections, pauses, restarts and abandons need two admins. Admins get a push notification when something needs their attention. Everything is logged.
- **Projector**: full-screen view cycling live match, standings, bracket and results.

## Rules

Race to 10, no draws, one point per win. Two groups of six, double round robin, top four qualify. Quarter-finals A1 v B4, B1 v A4, A2 v B3, B2 v A3, then semi-finals, third place, final. Tiebreak: points, goal difference, goals scored, head-to-head.

## Match states

`not_started → live → provisional → confirmed`

Live and provisional scores are public straight away. Only confirmed results count.

## Stack

Single Go binary, standard library, SQLite, embedded React UI, server-sent events, Web Push. Ships as one container.

## Repository

```
docs/architecture.md   rules, data model, API, deployment, theme
seed/players.csv       players
seed/matches.csv       group fixtures and results
```
