package seed

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

const maxNicknameLength = 12

type Tournament struct {
	Name         string
	Groups       int
	Rounds       int
	Qualifiers   int
	RaceTo       int
	PointsPerWin int
	MatchDays    []string
}

type Player struct {
	RegNo    int
	FullName string
	Nickname string
	Group    string
}

type Match struct {
	No      int
	Group   string
	Day     int
	Date    string
	PlayerA string
	PlayerB string
	ScoreA  int
	ScoreB  int
	Status  string
}

type Data struct {
	Tournament Tournament
	Players    []Player
	Matches    []Match
}

// Load reads tournament.csv, players.csv and matches.csv from fsys and validates them against each other.
func Load(fsys fs.FS) (Data, error) {
	t, err := loadTournament(fsys)
	if err != nil {
		return Data{}, err
	}
	players, err := loadPlayers(fsys, t)
	if err != nil {
		return Data{}, err
	}
	matches, days, err := loadMatches(fsys, t, players)
	if err != nil {
		return Data{}, err
	}
	t.MatchDays = days
	return Data{Tournament: t, Players: players, Matches: matches}, nil
}

func loadTournament(fsys fs.FS) (Tournament, error) {
	rows, err := readCSV(fsys, "tournament.csv", "name", "groups", "rounds", "qualifiers", "race_to", "points_per_win")
	if err != nil {
		return Tournament{}, err
	}
	if len(rows) != 1 {
		return Tournament{}, fmt.Errorf("tournament.csv: want exactly 1 row, got %d", len(rows))
	}
	r := rows[0]
	t := Tournament{Name: r.text("name")}
	ints := []struct {
		col string
		dst *int
	}{{"groups", &t.Groups}, {"rounds", &t.Rounds}, {"qualifiers", &t.Qualifiers}, {"race_to", &t.RaceTo}, {"points_per_win", &t.PointsPerWin}}
	for _, f := range ints {
		if *f.dst, err = r.positiveInt(f.col); err != nil {
			return Tournament{}, err
		}
	}
	if t.Name == "" {
		return Tournament{}, r.errorf("name is empty")
	}
	if t.Groups > 26 {
		return Tournament{}, r.errorf("groups must be 26 or fewer")
	}
	if n := t.Groups * t.Qualifiers; n&(n-1) != 0 {
		return Tournament{}, r.errorf("groups × qualifiers must be a power of two, got %d", n)
	}
	return t, nil
}

func loadPlayers(fsys fs.FS, t Tournament) ([]Player, error) {
	rows, err := readCSV(fsys, "players.csv", "reg_no", "full_name", "nickname", "group")
	if err != nil {
		return nil, err
	}
	players := make([]Player, 0, len(rows))
	seenReg, seenName, seenNick := map[int]bool{}, map[string]bool{}, map[string]bool{}
	for _, r := range rows {
		reg, err := r.positiveInt("reg_no")
		if err != nil {
			return nil, err
		}
		p := Player{RegNo: reg, FullName: r.text("full_name"), Nickname: r.text("nickname"), Group: r.text("group")}
		switch {
		case p.FullName == "":
			return nil, r.errorf("full_name is empty")
		case seenReg[p.RegNo]:
			return nil, r.errorf("duplicate reg_no %d", p.RegNo)
		case seenName[p.FullName]:
			return nil, r.errorf("duplicate full_name %q", p.FullName)
		case p.Nickname != "" && seenNick[p.Nickname]:
			return nil, r.errorf("duplicate nickname %q", p.Nickname)
		case len([]rune(p.Nickname)) > maxNicknameLength:
			return nil, r.errorf("nickname %q is longer than %d characters", p.Nickname, maxNicknameLength)
		case !validGroup(p.Group, t.Groups):
			return nil, r.errorf("group %q is not one of the tournament's %d groups", p.Group, t.Groups)
		}
		seenReg[p.RegNo], seenName[p.FullName], seenNick[p.Nickname] = true, true, p.Nickname != ""
		players = append(players, p)
	}
	return players, nil
}

func loadMatches(fsys fs.FS, t Tournament, players []Player) ([]Match, []string, error) {
	rows, err := readCSV(fsys, "matches.csv", "match_no", "group", "match_day", "date", "player_a", "player_b", "score_a", "score_b", "status")
	if err != nil {
		return nil, nil, err
	}
	groupOf := map[string]string{}
	for _, p := range players {
		groupOf[p.FullName] = p.Group
	}
	matches := make([]Match, 0, len(rows))
	seenNo := map[int]bool{}
	dayDates := map[int]string{}
	for _, r := range rows {
		m, err := parseMatch(r, t)
		if err != nil {
			return nil, nil, err
		}
		if err := checkMatch(r, m, groupOf, seenNo, dayDates); err != nil {
			return nil, nil, err
		}
		matches = append(matches, m)
	}
	return matches, sortedDates(dayDates), nil
}

func parseMatch(r row, t Tournament) (Match, error) {
	m := Match{Group: r.text("group"), Date: r.text("date"), PlayerA: r.text("player_a"), PlayerB: r.text("player_b"), Status: r.text("status")}
	var err error
	if m.No, err = r.positiveInt("match_no"); err != nil {
		return Match{}, err
	}
	if m.Day, err = r.positiveInt("match_day"); err != nil {
		return Match{}, err
	}
	switch m.Status {
	case "not_started":
		if r.text("score_a") != "" || r.text("score_b") != "" {
			return Match{}, r.errorf("not_started match must have empty scores")
		}
	case "confirmed":
		if m.ScoreA, err = r.score("score_a"); err != nil {
			return Match{}, err
		}
		if m.ScoreB, err = r.score("score_b"); err != nil {
			return Match{}, err
		}
		if !validResult(m.ScoreA, m.ScoreB, t.RaceTo) {
			return Match{}, r.errorf("score %d-%d is not a finished race to %d", m.ScoreA, m.ScoreB, t.RaceTo)
		}
	default:
		return Match{}, r.errorf("status %q must be not_started or confirmed", m.Status)
	}
	return m, nil
}

func checkMatch(r row, m Match, groupOf map[string]string, seenNo map[int]bool, dayDates map[int]string) error {
	ga, okA := groupOf[m.PlayerA]
	gb, okB := groupOf[m.PlayerB]
	switch {
	case seenNo[m.No]:
		return r.errorf("duplicate match_no %d", m.No)
	case !okA:
		return r.errorf("player_a %q is not in players.csv", m.PlayerA)
	case !okB:
		return r.errorf("player_b %q is not in players.csv", m.PlayerB)
	case m.PlayerA == m.PlayerB:
		return r.errorf("a player cannot play themselves")
	case ga != m.Group || gb != m.Group:
		return r.errorf("both players must be in group %s", m.Group)
	case m.Date == "":
		return r.errorf("date is empty")
	case dayDates[m.Day] != "" && dayDates[m.Day] != m.Date:
		return r.errorf("match_day %d has two dates: %s and %s", m.Day, dayDates[m.Day], m.Date)
	}
	seenNo[m.No] = true
	dayDates[m.Day] = m.Date
	return nil
}

// validResult reports whether exactly one player reached raceTo. There are no draws.
func validResult(a, b, raceTo int) bool {
	return (a == raceTo && b < raceTo) || (b == raceTo && a < raceTo)
}

func validGroup(g string, groups int) bool {
	return len(g) == 1 && g[0] >= 'A' && int(g[0]-'A') < groups
}

func sortedDates(dayDates map[int]string) []string {
	days := make([]int, 0, len(dayDates))
	for d := range dayDates {
		days = append(days, d)
	}
	sort.Ints(days)
	dates := make([]string, len(days))
	for i, d := range days {
		dates[i] = dayDates[d]
	}
	return dates
}

type row struct {
	file string
	line int
	cols map[string]string
}

func (r row) text(col string) string { return strings.TrimSpace(r.cols[col]) }

func (r row) errorf(format string, args ...any) error {
	return fmt.Errorf("%s line %d: %s", r.file, r.line, fmt.Sprintf(format, args...))
}

func (r row) positiveInt(col string) (int, error) {
	n, err := strconv.Atoi(r.text(col))
	if err != nil || n < 1 {
		return 0, r.errorf("%s must be a whole number above 0, got %q", col, r.text(col))
	}
	return n, nil
}

func (r row) score(col string) (int, error) {
	n, err := strconv.Atoi(r.text(col))
	if err != nil || n < 0 {
		return 0, r.errorf("%s must be a whole number, got %q", col, r.text(col))
	}
	return n, nil
}

func readCSV(fsys fs.FS, name string, required ...string) ([]row, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", name, err)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s: missing header", name)
	}
	index := map[string]int{}
	for i, h := range records[0] {
		index[strings.TrimSpace(strings.TrimPrefix(h, "\ufeff"))] = i
	}
	var missing []string
	for _, col := range required {
		if _, ok := index[col]; !ok {
			missing = append(missing, col)
		}
	}
	if len(missing) > 0 {
		return nil, errors.New(name + ": missing columns " + strings.Join(missing, ", "))
	}
	rows := make([]row, 0, len(records)-1)
	for i, rec := range records[1:] {
		cols := make(map[string]string, len(required))
		for _, col := range required {
			if idx := index[col]; idx < len(rec) {
				cols[col] = rec[idx]
			}
		}
		rows = append(rows, row{file: name, line: i + 2, cols: cols})
	}
	return rows, nil
}
