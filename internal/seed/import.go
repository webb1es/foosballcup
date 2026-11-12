package seed

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

const importAdminLabel = "import"

// ImportIfEmpty writes d into db when the database holds no tournament yet.
// It reports whether anything was imported.
func ImportIfEmpty(ctx context.Context, db *sql.DB, d Data, now time.Time) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin import: %w", err)
	}
	defer tx.Rollback()

	var existing int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM tournaments`).Scan(&existing); err != nil {
		return false, fmt.Errorf("count tournaments: %w", err)
	}
	if existing > 0 {
		return false, nil
	}

	ts := now.UTC().Format(time.RFC3339)
	tournamentID, err := insertTournament(ctx, tx, d.Tournament, ts)
	if err != nil {
		return false, err
	}
	// Business rule: imported results count as confirmed, attributed to a system admin that can never log in.
	adminID, err := insertID(ctx, tx, `INSERT INTO admins (tournament_id, label, display_name, active, created_at)
		VALUES (?, ?, 'Import', 0, ?)`, tournamentID, importAdminLabel, ts)
	if err != nil {
		return false, fmt.Errorf("insert import admin: %w", err)
	}
	playerIDs, err := insertPlayers(ctx, tx, tournamentID, d.Players)
	if err != nil {
		return false, err
	}
	if err := insertMatches(ctx, tx, tournamentID, adminID, playerIDs, d.Matches, ts); err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit import: %w", err)
	}
	return true, nil
}

func insertTournament(ctx context.Context, tx *sql.Tx, t Tournament, ts string) (int64, error) {
	days, err := json.Marshal(t.MatchDays)
	if err != nil {
		return 0, fmt.Errorf("encode match days: %w", err)
	}
	id, err := insertID(ctx, tx, `INSERT INTO tournaments
		(name, group_count, rounds, qualifiers, race_to, points_per_win, match_days, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.Name, t.Groups, t.Rounds, t.Qualifiers, t.RaceTo, t.PointsPerWin, string(days), ts)
	if err != nil {
		return 0, fmt.Errorf("insert tournament: %w", err)
	}
	return id, nil
}

func insertPlayers(ctx context.Context, tx *sql.Tx, tournamentID int64, players []Player) (map[string]int64, error) {
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO players (tournament_id, reg_no, full_name, nickname, group_label)
		VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("prepare players: %w", err)
	}
	defer stmt.Close()
	ids := make(map[string]int64, len(players))
	for _, p := range players {
		res, err := stmt.ExecContext(ctx, tournamentID, p.RegNo, p.FullName, nullIfEmpty(p.Nickname), p.Group)
		if err != nil {
			return nil, fmt.Errorf("insert player %q: %w", p.FullName, err)
		}
		if ids[p.FullName], err = res.LastInsertId(); err != nil {
			return nil, fmt.Errorf("player id %q: %w", p.FullName, err)
		}
	}
	return ids, nil
}

func insertMatches(ctx context.Context, tx *sql.Tx, tournamentID, adminID int64, playerIDs map[string]int64, matches []Match, ts string) error {
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO matches
		(tournament_id, stage, group_label, match_no, match_day, scheduled_date, player_a_id, player_b_id,
		 score_a, score_b, status, submitted_by, submitted_at, confirmed_by, confirmed_at)
		VALUES (?, 'group', ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare matches: %w", err)
	}
	defer stmt.Close()
	event, err := tx.PrepareContext(ctx, `INSERT INTO match_events (tournament_id, match_id, admin_id, type, payload, created_at)
		VALUES (?, ?, ?, 'import', ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare match events: %w", err)
	}
	defer event.Close()

	for _, m := range matches {
		var by, at any
		if m.Status == "confirmed" {
			by, at = adminID, ts
		}
		res, err := stmt.ExecContext(ctx, tournamentID, m.Group, m.No, m.Day, m.Date,
			playerIDs[m.PlayerA], playerIDs[m.PlayerB], m.ScoreA, m.ScoreB, m.Status, by, at, by, at)
		if err != nil {
			return fmt.Errorf("insert match %d: %w", m.No, err)
		}
		matchID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("match id %d: %w", m.No, err)
		}
		payload, err := json.Marshal(map[string]any{"status": m.Status, "score_a": m.ScoreA, "score_b": m.ScoreB})
		if err != nil {
			return fmt.Errorf("encode event %d: %w", m.No, err)
		}
		if _, err := event.ExecContext(ctx, tournamentID, matchID, adminID, string(payload), ts); err != nil {
			return fmt.Errorf("insert event for match %d: %w", m.No, err)
		}
	}
	return nil
}

func insertID(ctx context.Context, tx *sql.Tx, query string, args ...any) (int64, error) {
	res, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
