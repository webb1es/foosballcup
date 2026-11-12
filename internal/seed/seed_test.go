package seed

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/webb1es/foosballcup/internal/store"
)

func loadRepoSeed(t *testing.T) Data {
	t.Helper()
	d, err := Load(os.DirFS(filepath.Join("..", "..", "seed")))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	return d
}

func TestLoad_RepoSeed_ParsesTournament(t *testing.T) {
	d := loadRepoSeed(t)

	if d.Tournament.Name != "Entelect Zim Foosball Cup" {
		t.Errorf("Name = %q", d.Tournament.Name)
	}
	if got := strings.Join(d.Tournament.MatchDays, ","); got != "2026-11-11,2026-11-13,2026-11-18" {
		t.Errorf("MatchDays = %s", got)
	}
	if len(d.Players) != 12 || len(d.Matches) != 60 {
		t.Errorf("players = %d, matches = %d, want 12 and 60", len(d.Players), len(d.Matches))
	}
}

func TestImportIfEmpty_FreshDatabase_LoadsSeed(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()

	imported, err := ImportIfEmpty(ctx, db, loadRepoSeed(t), time.Date(2026, 11, 12, 9, 0, 0, 0, time.UTC))
	if err != nil || !imported {
		t.Fatalf("ImportIfEmpty() = %v, %v, want true, nil", imported, err)
	}

	counts := []struct {
		query string
		want  int
	}{
		{`SELECT count(*) FROM tournaments`, 1},
		{`SELECT count(*) FROM players`, 12},
		{`SELECT count(*) FROM players WHERE group_label = 'A'`, 6},
		{`SELECT count(*) FROM matches WHERE stage = 'group'`, 60},
		{`SELECT count(*) FROM matches WHERE status = 'confirmed' AND confirmed_by IS NOT NULL`, 8},
		{`SELECT count(*) FROM match_events WHERE type = 'import'`, 60},
		{`SELECT count(*) FROM admins WHERE label = 'import' AND active = 0`, 1},
	}
	for _, c := range counts {
		var got int
		if err := db.QueryRow(c.query).Scan(&got); err != nil {
			t.Fatalf("%s: %v", c.query, err)
		}
		if got != c.want {
			t.Errorf("%s = %d, want %d", c.query, got, c.want)
		}
	}
}

func TestImportIfEmpty_SecondRun_SkipsImport(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	defer db.Close()
	d := loadRepoSeed(t)

	if _, err := ImportIfEmpty(ctx, db, d, time.Now()); err != nil {
		t.Fatalf("first import: %v", err)
	}
	imported, err := ImportIfEmpty(ctx, db, d, time.Now())
	if err != nil || imported {
		t.Fatalf("second ImportIfEmpty() = %v, %v, want false, nil", imported, err)
	}
	var players int
	if err := db.QueryRow(`SELECT count(*) FROM players`).Scan(&players); err != nil {
		t.Fatal(err)
	}
	if players != 12 {
		t.Errorf("players = %d after second run, want 12", players)
	}
}

func TestLoad_InvalidSeed(t *testing.T) {
	const tournament = "name,groups,rounds,qualifiers,race_to,points_per_win\nCup,2,2,4,10,1\n"
	const players = "reg_no,full_name,nickname,group\n1,Ann,,A\n2,Ben,,A\n3,Cal,,B\n"
	const matchHeader = "match_no,group,match_day,date,player_a,player_b,score_a,score_b,status\n"

	tests := []struct {
		name    string
		files   map[string]string
		wantErr string
	}{
		{name: "qualifiers not a power of two", files: map[string]string{"tournament.csv": "name,groups,rounds,qualifiers,race_to,points_per_win\nCup,2,2,3,10,1\n"}, wantErr: "power of two"},
		{name: "player in unknown group", files: map[string]string{"players.csv": "reg_no,full_name,nickname,group\n1,Ann,,C\n"}, wantErr: "group \"C\""},
		{name: "duplicate reg_no", files: map[string]string{"players.csv": "reg_no,full_name,nickname,group\n1,Ann,,A\n1,Ben,,A\n"}, wantErr: "duplicate reg_no"},
		{name: "nickname too long", files: map[string]string{"players.csv": "reg_no,full_name,nickname,group\n1,Ann,Thirteenchars,A\n"}, wantErr: "longer than 12"},
		{name: "unknown player", files: map[string]string{"matches.csv": matchHeader + "1,A,1,2026-11-11,Ann,Zed,,,not_started\n"}, wantErr: "\"Zed\" is not in players.csv"},
		{name: "players from different groups", files: map[string]string{"matches.csv": matchHeader + "1,A,1,2026-11-11,Ann,Cal,,,not_started\n"}, wantErr: "group A"},
		{name: "confirmed without a winner", files: map[string]string{"matches.csv": matchHeader + "1,A,1,2026-11-11,Ann,Ben,9,8,confirmed\n"}, wantErr: "not a finished race"},
		{name: "not started with a score", files: map[string]string{"matches.csv": matchHeader + "1,A,1,2026-11-11,Ann,Ben,1,0,not_started\n"}, wantErr: "empty scores"},
		{name: "unsupported status", files: map[string]string{"matches.csv": matchHeader + "1,A,1,2026-11-11,Ann,Ben,,,live\n"}, wantErr: "status \"live\""},
		{name: "match day with two dates", files: map[string]string{"matches.csv": matchHeader + "1,A,1,2026-11-11,Ann,Ben,,,not_started\n2,A,1,2026-11-12,Ben,Ann,,,not_started\n"}, wantErr: "two dates"},
		{name: "missing column", files: map[string]string{"matches.csv": "match_no,group\n1,A\n"}, wantErr: "missing columns"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"tournament.csv": {Data: []byte(tournament)},
				"players.csv":    {Data: []byte(players)},
				"matches.csv":    {Data: []byte(matchHeader)},
			}
			for name, body := range tt.files {
				fsys[name] = &fstest.MapFile{Data: []byte(body)}
			}
			_, err := Load(fsys)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Load() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
