package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestOpen_FreshFile_CreatesSchema(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	for _, table := range []string{"tournaments", "players", "admins", "matches", "match_events", "admin_events", "pending_actions", "push_subscriptions"} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %s missing: %v", table, err)
		}
	}
}

func TestOpen_Twice_AppliesMigrationsOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	for i := 0; i < 2; i++ {
		db, err := Open(context.Background(), path)
		if err != nil {
			t.Fatalf("Open() #%d error = %v", i+1, err)
		}
		db.Close()
	}

	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()
	var applied int
	if err := db.QueryRow(`SELECT count(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if applied != 1 {
		t.Errorf("schema_migrations rows = %d, want 1", applied)
	}
}

func TestOpen_ForeignKeys_AreEnforced(t *testing.T) {
	db, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`INSERT INTO players (tournament_id, reg_no, full_name, group_label) VALUES (999, 1, 'Nobody', 'A')`)
	if err == nil {
		t.Error("insert with unknown tournament_id succeeded, want foreign key error")
	}
}
