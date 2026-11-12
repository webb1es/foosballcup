package tournament

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/webb1es/foosballcup/internal/seed"
	"github.com/webb1es/foosballcup/internal/store"
)

func seededDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	data, err := seed.Load(os.DirFS(filepath.Join("..", "..", "seed")))
	if err != nil {
		t.Fatalf("seed.Load() error = %v", err)
	}
	if _, err := seed.ImportIfEmpty(ctx, db, data, time.Now()); err != nil {
		t.Fatalf("seed.ImportIfEmpty() error = %v", err)
	}
	return db
}

// table renders a group's standings as "rank name P W L GF GA GD Pts" lines.
func table(snap Snapshot, group string) string {
	names := map[int64]string{}
	for _, p := range snap.Players {
		names[p.ID] = p.FullName
	}
	var lines []string
	for _, g := range snap.Standings {
		if g.Group != group {
			continue
		}
		for _, r := range g.Rows {
			lines = append(lines, fmt.Sprintf("%d %s %d %d %d %d %d %d %d", r.Rank, names[r.PlayerID],
				r.Played, r.Won, r.Lost, r.GoalsFor, r.GoalsAgainst, r.GoalDifference, r.Points))
		}
	}
	return strings.Join(lines, "\n")
}

func TestLoad_SeedStandings_MatchHandCalculation(t *testing.T) {
	snap, err := Load(context.Background(), seededDB(t))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Hand calculation from the 8 confirmed results in seed/matches.csv.
	wantA := strings.Join([]string{
		"1 Trust Mandinyenya 2 1 1 15 13 2 1",
		"2 Thomas Nyagwaya 2 1 1 13 13 0 1",
		"3 Benard Mwanjeya 2 1 1 13 15 -2 1",
		"4 Nation Chirara 0 0 0 0 0 0 0",
		"4 Tafadzwa Utete 0 0 0 0 0 0 0",
		"4 Taona Madzudzo 0 0 0 0 0 0 0",
	}, "\n")
	wantB := strings.Join([]string{
		"1 Tapfuma Chigwada 2 2 0 20 8 12 2",
		"2 Romeo Jerenyama 2 1 1 19 16 3 1",
		"3 Fungai Mashozhera 2 1 1 16 14 2 1",
		"4 Kudzanai Tsuro 2 1 1 14 19 -5 1",
		"5 Takudzwa Chikwanda 0 0 0 0 0 0 0",
		"6 Job Gwese 2 0 2 8 20 -12 0",
	}, "\n")
	if got := table(snap, "A"); got != wantA {
		t.Errorf("group A\n got:\n%s\nwant:\n%s", got, wantA)
	}
	if got := table(snap, "B"); got != wantB {
		t.Errorf("group B\n got:\n%s\nwant:\n%s", got, wantB)
	}
}

func TestLoad_SeedSnapshot_HasTournamentPlayersAndMatches(t *testing.T) {
	snap, err := Load(context.Background(), seededDB(t))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	info := snap.Tournament
	if info.Name != "Entelect Zim Foosball Cup" || strings.Join(info.Groups, "") != "AB" || info.RaceTo != 10 || info.Qualifiers != 4 {
		t.Errorf("tournament = %+v", info)
	}
	if len(snap.Players) != 12 || len(snap.Matches) != 60 || len(snap.Bracket) != 0 {
		t.Errorf("players = %d, matches = %d, bracket = %d", len(snap.Players), len(snap.Matches), len(snap.Bracket))
	}
}

func TestLoad_LiveScore_DoesNotChangeStandings(t *testing.T) {
	db := seededDB(t)
	ctx := context.Background()
	before, err := Load(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE matches SET status = 'provisional', score_a = 10, score_b = 0 WHERE match_no = 1`); err != nil {
		t.Fatal(err)
	}
	after, err := Load(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if table(before, "A") != table(after, "A") {
		t.Errorf("provisional result changed standings:\n%s", table(after, "A"))
	}
}

func TestLoad_SnapshotJSON_HidesPrivateFields(t *testing.T) {
	db := seededDB(t)
	if _, err := db.Exec(`UPDATE players SET email = 'secret@example.com', department = 'Finance'`); err != nil {
		t.Fatal(err)
	}
	snap, err := Load(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	body, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	for _, leaked := range []string{"secret@example.com", "Finance", "email", "department", "reg_no"} {
		if strings.Contains(string(body), leaked) {
			t.Errorf("snapshot JSON contains %q", leaked)
		}
	}
}

func TestLoad_EmptyDatabase_ReturnsErrNotFound(t *testing.T) {
	db, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := Load(context.Background(), db); !errors.Is(err, ErrNotFound) {
		t.Errorf("Load() error = %v, want ErrNotFound", err)
	}
}
