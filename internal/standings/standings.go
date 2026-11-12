package standings

import "sort"

type Player struct {
	ID   int64
	Name string
}

// Result is a confirmed match between two players in the same group.
type Result struct {
	PlayerA int64
	PlayerB int64
	ScoreA  int
	ScoreB  int
}

type Row struct {
	PlayerID       int64 `json:"player_id"`
	Rank           int   `json:"rank"`
	Played         int   `json:"played"`
	Won            int   `json:"won"`
	Lost           int   `json:"lost"`
	GoalsFor       int   `json:"goals_for"`
	GoalsAgainst   int   `json:"goals_against"`
	GoalDifference int   `json:"goal_difference"`
	Points         int   `json:"points"`

	name string
}

// Compute ranks every player by points, goal difference, goals scored, then
// head-to-head points among the players still tied. Players level after all
// four share a rank. Results involving players outside the list are ignored.
func Compute(players []Player, results []Result, pointsPerWin int) []Row {
	rows := tally(players, results, pointsPerWin)
	sort.SliceStable(rows, func(i, j int) bool { return before(rows[i], rows[j]) })

	for start := 0; start < len(rows); {
		end := start + 1
		for end < len(rows) && level(rows[start], rows[end]) {
			end++
		}
		rankBlock(rows[start:end], start+1, results, pointsPerWin)
		start = end
	}
	return rows
}

func tally(players []Player, results []Result, pointsPerWin int) []Row {
	rows := make([]Row, len(players))
	index := make(map[int64]int, len(players))
	for i, p := range players {
		rows[i] = Row{PlayerID: p.ID, name: p.Name}
		index[p.ID] = i
	}
	for _, r := range results {
		a, okA := index[r.PlayerA]
		b, okB := index[r.PlayerB]
		if !okA || !okB {
			continue
		}
		record(&rows[a], r.ScoreA, r.ScoreB, pointsPerWin)
		record(&rows[b], r.ScoreB, r.ScoreA, pointsPerWin)
	}
	return rows
}

func record(row *Row, scored, conceded, pointsPerWin int) {
	row.Played++
	row.GoalsFor += scored
	row.GoalsAgainst += conceded
	row.GoalDifference = row.GoalsFor - row.GoalsAgainst
	if scored > conceded {
		row.Won++
		row.Points += pointsPerWin
	} else {
		row.Lost++
	}
}

func before(a, b Row) bool {
	if !level(a, b) {
		if a.Points != b.Points {
			return a.Points > b.Points
		}
		if a.GoalDifference != b.GoalDifference {
			return a.GoalDifference > b.GoalDifference
		}
		return a.GoalsFor > b.GoalsFor
	}
	return a.name < b.name
}

func level(a, b Row) bool {
	return a.Points == b.Points && a.GoalDifference == b.GoalDifference && a.GoalsFor == b.GoalsFor
}

// rankBlock orders players level on points, goal difference and goals scored
// by their head-to-head points against each other and assigns ranks from firstRank.
func rankBlock(block []Row, firstRank int, results []Result, pointsPerWin int) {
	if len(block) == 1 {
		block[0].Rank = firstRank
		return
	}
	members := make([]Player, len(block))
	for i, r := range block {
		members[i] = Player{ID: r.PlayerID, Name: r.name}
	}
	h2h := map[int64]int{}
	for _, r := range tally(members, results, pointsPerWin) {
		h2h[r.PlayerID] = r.Points
	}
	sort.SliceStable(block, func(i, j int) bool {
		pi, pj := h2h[block[i].PlayerID], h2h[block[j].PlayerID]
		if pi != pj {
			return pi > pj
		}
		return block[i].name < block[j].name
	})
	for i := range block {
		if i > 0 && h2h[block[i].PlayerID] == h2h[block[i-1].PlayerID] {
			block[i].Rank = block[i-1].Rank
		} else {
			block[i].Rank = firstRank + i
		}
	}
}
