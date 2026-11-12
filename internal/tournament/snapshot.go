package tournament

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/webb1es/foosballcup/internal/standings"
)

// ErrNotFound means the database holds no tournament yet.
var ErrNotFound = errors.New("no tournament")

// Snapshot is the public view of the tournament served by GET /api/tournament.
// It never contains player email or department.
type Snapshot struct {
	Tournament Info            `json:"tournament"`
	Players    []Player        `json:"players"`
	Matches    []Match         `json:"matches"`
	Standings  []GroupStanding `json:"standings"`
	Bracket    []Match         `json:"bracket"`
}

type Info struct {
	Name             string   `json:"name"`
	Groups           []string `json:"groups"`
	Rounds           int      `json:"rounds"`
	Qualifiers       int      `json:"qualifiers"`
	RaceTo           int      `json:"race_to"`
	PointsPerWin     int      `json:"points_per_win"`
	MatchDays        []string `json:"match_days"`
	GroupStageLocked bool     `json:"group_stage_locked"`

	id int64
}

type Player struct {
	ID        int64   `json:"id"`
	FullName  string  `json:"full_name"`
	Nickname  *string `json:"nickname"`
	AvatarURL *string `json:"avatar_url"`
	Group     string  `json:"group"`
}

type Match struct {
	ID            int64   `json:"id"`
	Stage         string  `json:"stage"`
	Group         *string `json:"group"`
	MatchNo       int     `json:"match_no"`
	MatchDay      *int    `json:"match_day"`
	ScheduledDate *string `json:"scheduled_date"`
	SlotA         *string `json:"slot_a"`
	SlotB         *string `json:"slot_b"`
	PlayerAID     *int64  `json:"player_a_id"`
	PlayerBID     *int64  `json:"player_b_id"`
	ScoreA        int     `json:"score_a"`
	ScoreB        int     `json:"score_b"`
	Status        string  `json:"status"`
	StartedAt     *string `json:"started_at"`
	FullTimeAt    *string `json:"full_time_at"`
	ConfirmedAt   *string `json:"confirmed_at"`
	Version       int     `json:"version"`
}

type GroupStanding struct {
	Group string          `json:"group"`
	Rows  []standings.Row `json:"rows"`
}

// Load reads the tournament and computes standings from confirmed group matches.
func Load(ctx context.Context, db *sql.DB) (Snapshot, error) {
	info, err := loadInfo(ctx, db)
	if err != nil {
		return Snapshot{}, err
	}
	players, err := loadPlayers(ctx, db, info.id)
	if err != nil {
		return Snapshot{}, err
	}
	matches, err := loadMatches(ctx, db, info.id)
	if err != nil {
		return Snapshot{}, err
	}

	var group, bracket []Match
	for _, m := range matches {
		if m.Stage == "group" {
			group = append(group, m)
		} else {
			bracket = append(bracket, m)
		}
	}
	return Snapshot{
		Tournament: info,
		Players:    players,
		Matches:    nonNil(group),
		Standings:  groupStandings(info, players, group),
		Bracket:    nonNil(bracket),
	}, nil
}

func loadInfo(ctx context.Context, db *sql.DB) (Info, error) {
	var (
		info      Info
		groups    int
		matchDays string
		lockedAt  sql.NullString
	)
	err := db.QueryRowContext(ctx, `SELECT id, name, group_count, rounds, qualifiers, race_to, points_per_win,
		match_days, group_stage_locked_at FROM tournaments ORDER BY id LIMIT 1`).
		Scan(&info.id, &info.Name, &groups, &info.Rounds, &info.Qualifiers, &info.RaceTo, &info.PointsPerWin, &matchDays, &lockedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Info{}, ErrNotFound
	}
	if err != nil {
		return Info{}, fmt.Errorf("read tournament: %w", err)
	}
	if err := json.Unmarshal([]byte(matchDays), &info.MatchDays); err != nil {
		return Info{}, fmt.Errorf("decode match days: %w", err)
	}
	info.Groups = make([]string, groups)
	for i := range info.Groups {
		info.Groups[i] = string(rune('A' + i))
	}
	info.GroupStageLocked = lockedAt.Valid
	return info, nil
}

func loadPlayers(ctx context.Context, db *sql.DB, tournamentID int64) ([]Player, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, full_name, nickname, avatar_url, group_label
		FROM players WHERE tournament_id = ? ORDER BY group_label, reg_no`, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("read players: %w", err)
	}
	defer rows.Close()
	players := []Player{}
	for rows.Next() {
		var p Player
		var nickname, avatar sql.NullString
		if err := rows.Scan(&p.ID, &p.FullName, &nickname, &avatar, &p.Group); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		p.Nickname, p.AvatarURL = stringPtr(nickname), stringPtr(avatar)
		players = append(players, p)
	}
	return players, rows.Err()
}

func loadMatches(ctx context.Context, db *sql.DB, tournamentID int64) ([]Match, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, stage, group_label, match_no, match_day, scheduled_date,
		slot_a, slot_b, player_a_id, player_b_id, score_a, score_b, status, started_at, full_time_at,
		confirmed_at, version
		FROM matches WHERE tournament_id = ? ORDER BY match_no`, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("read matches: %w", err)
	}
	defer rows.Close()
	var matches []Match
	for rows.Next() {
		var m Match
		var group, date, slotA, slotB, started, fullTime, confirmed sql.NullString
		var day sql.NullInt64
		var playerA, playerB sql.NullInt64
		if err := rows.Scan(&m.ID, &m.Stage, &group, &m.MatchNo, &day, &date, &slotA, &slotB, &playerA, &playerB,
			&m.ScoreA, &m.ScoreB, &m.Status, &started, &fullTime, &confirmed, &m.Version); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		m.Group, m.ScheduledDate, m.SlotA, m.SlotB = stringPtr(group), stringPtr(date), stringPtr(slotA), stringPtr(slotB)
		m.StartedAt, m.FullTimeAt, m.ConfirmedAt = stringPtr(started), stringPtr(fullTime), stringPtr(confirmed)
		m.PlayerAID, m.PlayerBID = int64Ptr(playerA), int64Ptr(playerB)
		if day.Valid {
			d := int(day.Int64)
			m.MatchDay = &d
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}

// groupStandings ranks each group using confirmed results only. Live and provisional scores never count.
func groupStandings(info Info, players []Player, matches []Match) []GroupStanding {
	out := make([]GroupStanding, 0, len(info.Groups))
	for _, g := range info.Groups {
		var members []standings.Player
		for _, p := range players {
			if p.Group == g {
				members = append(members, standings.Player{ID: p.ID, Name: p.FullName})
			}
		}
		var results []standings.Result
		for _, m := range matches {
			if m.Status != "confirmed" || m.Group == nil || *m.Group != g || m.PlayerAID == nil || m.PlayerBID == nil {
				continue
			}
			results = append(results, standings.Result{PlayerA: *m.PlayerAID, PlayerB: *m.PlayerBID, ScoreA: m.ScoreA, ScoreB: m.ScoreB})
		}
		out = append(out, GroupStanding{Group: g, Rows: standings.Compute(members, results, info.PointsPerWin)})
	}
	return out
}

func stringPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

func int64Ptr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

func nonNil(ms []Match) []Match {
	if ms == nil {
		return []Match{}
	}
	return ms
}
