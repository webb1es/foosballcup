package standings

import (
	"fmt"
	"strings"
	"testing"
)

// summary renders rows as "rank:name:points:gd:gf" for compact comparison.
func summary(rows []Row) string {
	parts := make([]string, len(rows))
	for i, r := range rows {
		parts[i] = fmt.Sprintf("%d:%s:%d:%d:%d", r.Rank, r.name, r.Points, r.GoalDifference, r.GoalsFor)
	}
	return strings.Join(parts, " ")
}

func players(names ...string) []Player {
	ps := make([]Player, len(names))
	for i, n := range names {
		ps[i] = Player{ID: int64(i + 1), Name: n}
	}
	return ps
}

func TestCompute(t *testing.T) {
	const ann, ben, cal, dee = 1, 2, 3, 4

	tests := []struct {
		name         string
		players      []Player
		results      []Result
		pointsPerWin int
		want         string
	}{
		{
			name:    "no results: everyone shares first",
			players: players("Ann", "Ben", "Cal"),
			want:    "1:Ann:0:0:0 1:Ben:0:0:0 1:Cal:0:0:0",
		},
		{
			name:    "points decide",
			players: players("Ann", "Ben", "Cal"),
			results: []Result{{ben, ann, 10, 9}, {ben, cal, 10, 9}, {ann, cal, 10, 9}},
			want:    "1:Ben:2:2:20 2:Ann:1:0:19 3:Cal:0:-2:18",
		},
		{
			name:    "goal difference breaks a points tie",
			players: players("Ann", "Ben", "Cal", "Dee"),
			results: []Result{{ann, cal, 10, 2}, {ben, dee, 10, 8}},
			want:    "1:Ann:1:8:10 2:Ben:1:2:10 3:Dee:0:-2:8 4:Cal:0:-8:2",
		},
		{
			name:    "goals scored breaks a goal difference tie",
			players: players("Ann", "Ben", "Cal", "Dee"),
			results: []Result{{ann, cal, 10, 5}, {ann, dee, 3, 10}, {ben, cal, 10, 6}, {ben, dee, 4, 10}},
			want:    "1:Dee:2:13:20 2:Ben:1:-2:14 3:Ann:1:-2:13 4:Cal:0:-9:11",
		},
		{
			name:    "head-to-head breaks a full tie ahead of name order",
			players: players("Ann", "Ben", "Cal", "Dee"),
			results: []Result{{ben, ann, 10, 8}, {cal, ben, 10, 8}, {ann, dee, 10, 8}},
			want:    "1:Cal:1:2:10 2:Ben:1:0:18 3:Ann:1:0:18 4:Dee:0:-2:8",
		},
		{
			name:    "split head-to-head shares the rank",
			players: players("Ann", "Ben", "Cal"),
			results: []Result{{ann, ben, 10, 7}, {ben, ann, 10, 7}, {ann, cal, 10, 7}, {ben, cal, 10, 7}},
			want:    "1:Ann:2:3:27 1:Ben:2:3:27 3:Cal:0:-6:14",
		},
		{
			name:    "three-way cycle shares the rank and the next rank skips",
			players: players("Ann", "Ben", "Cal", "Dee"),
			results: []Result{{ann, ben, 10, 8}, {ben, cal, 10, 8}, {cal, ann, 10, 8}},
			want:    "1:Ann:1:0:18 1:Ben:1:0:18 1:Cal:1:0:18 4:Dee:0:0:0",
		},
		{
			name:         "points per win scales points only",
			players:      players("Ann", "Ben"),
			results:      []Result{{ann, ben, 10, 4}},
			pointsPerWin: 3,
			want:         "1:Ann:3:6:10 2:Ben:0:-6:4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ppw := tt.pointsPerWin
			if ppw == 0 {
				ppw = 1
			}
			if got := summary(Compute(tt.players, tt.results, ppw)); got != tt.want {
				t.Errorf("\n got  %s\n want %s", got, tt.want)
			}
		})
	}
}

func TestCompute_ResultAgainstOutsider_IsIgnored(t *testing.T) {
	rows := Compute(players("Ann"), []Result{{1, 99, 10, 0}}, 1)
	if rows[0].Played != 0 {
		t.Errorf("Played = %d, want 0", rows[0].Played)
	}
}
