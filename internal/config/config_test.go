package config

import "testing"

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantPort   int
		wantDBPath string
		wantErr    bool
	}{
		{name: "unset uses defaults", wantPort: 8080, wantDBPath: "foosballcup.db"},
		{name: "valid port", env: map[string]string{"PORT": "9000"}, wantPort: 9000, wantDBPath: "foosballcup.db"},
		{name: "upper bound", env: map[string]string{"PORT": "65535"}, wantPort: 65535, wantDBPath: "foosballcup.db"},
		{name: "database path set", env: map[string]string{"DATABASE_PATH": "/data/cup.db"}, wantPort: 8080, wantDBPath: "/data/cup.db"},
		{name: "port not a number", env: map[string]string{"PORT": "abc"}, wantErr: true},
		{name: "port zero", env: map[string]string{"PORT": "0"}, wantErr: true},
		{name: "port above range", env: map[string]string{"PORT": "70000"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(func(key string) string { return tt.env[key] })
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.Port != tt.wantPort {
				t.Errorf("Port = %d, want %d", cfg.Port, tt.wantPort)
			}
			if cfg.DatabasePath != tt.wantDBPath {
				t.Errorf("DatabasePath = %q, want %q", cfg.DatabasePath, tt.wantDBPath)
			}
		})
	}
}

func TestAddr_Port8080_ReturnsColonPort(t *testing.T) {
	if got := (Config{Port: 8080}).Addr(); got != ":8080" {
		t.Errorf("Addr() = %q, want %q", got, ":8080")
	}
}
