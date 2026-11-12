CREATE TABLE tournaments (
  id                    INTEGER PRIMARY KEY,
  name                  TEXT    NOT NULL,
  group_count           INTEGER NOT NULL CHECK (group_count >= 1),
  rounds                INTEGER NOT NULL CHECK (rounds >= 1),
  qualifiers            INTEGER NOT NULL CHECK (qualifiers >= 1),
  race_to               INTEGER NOT NULL CHECK (race_to >= 1),
  points_per_win        INTEGER NOT NULL CHECK (points_per_win >= 1),
  match_days            TEXT    NOT NULL DEFAULT '[]',
  group_stage_locked_at TEXT,
  created_at            TEXT    NOT NULL
);

CREATE TABLE players (
  id            INTEGER PRIMARY KEY,
  tournament_id INTEGER NOT NULL REFERENCES tournaments (id),
  reg_no        INTEGER NOT NULL,
  full_name     TEXT    NOT NULL,
  nickname      TEXT    CHECK (nickname IS NULL OR length(nickname) BETWEEN 1 AND 12),
  avatar_url    TEXT,
  group_label   TEXT    NOT NULL,
  email         TEXT,
  department    TEXT,
  UNIQUE (tournament_id, reg_no),
  UNIQUE (tournament_id, full_name),
  UNIQUE (tournament_id, nickname)
);

CREATE TABLE admins (
  id              INTEGER PRIMARY KEY,
  tournament_id   INTEGER NOT NULL REFERENCES tournaments (id),
  label           TEXT    NOT NULL,
  display_name    TEXT    NOT NULL,
  pin_hash        TEXT    NOT NULL DEFAULT '',
  active          INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
  failed_attempts INTEGER NOT NULL DEFAULT 0,
  locked_until    TEXT,
  created_by      INTEGER REFERENCES admins (id),
  created_at      TEXT    NOT NULL,
  UNIQUE (tournament_id, label)
);

CREATE TABLE matches (
  id              INTEGER PRIMARY KEY,
  tournament_id   INTEGER NOT NULL REFERENCES tournaments (id),
  stage           TEXT    NOT NULL CHECK (stage IN ('group', 'r16', 'qf', 'sf', 'third', 'final')),
  group_label     TEXT,
  match_no        INTEGER NOT NULL,
  match_day       INTEGER,
  scheduled_date  TEXT,
  slot_a          TEXT,
  slot_b          TEXT,
  player_a_id     INTEGER REFERENCES players (id),
  player_b_id     INTEGER REFERENCES players (id),
  score_a         INTEGER NOT NULL DEFAULT 0 CHECK (score_a >= 0),
  score_b         INTEGER NOT NULL DEFAULT 0 CHECK (score_b >= 0),
  status          TEXT    NOT NULL DEFAULT 'not_started'
                  CHECK (status IN ('not_started', 'live', 'paused', 'provisional', 'confirmed', 'abandoned')),
  started_at      TEXT,
  full_time_at    TEXT,
  submitted_by    INTEGER REFERENCES admins (id),
  submitted_at    TEXT,
  confirmed_by    INTEGER REFERENCES admins (id),
  confirmed_at    TEXT,
  scorer_admin_id INTEGER REFERENCES admins (id),
  send_back_note  TEXT,
  version         INTEGER NOT NULL DEFAULT 0,
  UNIQUE (tournament_id, match_no),
  CHECK (player_a_id IS NULL OR player_b_id IS NULL OR player_a_id <> player_b_id)
);

CREATE INDEX matches_by_status ON matches (tournament_id, status);
CREATE INDEX matches_by_player_a ON matches (player_a_id);
CREATE INDEX matches_by_player_b ON matches (player_b_id);

CREATE TABLE match_events (
  id            INTEGER PRIMARY KEY,
  tournament_id INTEGER NOT NULL REFERENCES tournaments (id),
  match_id      INTEGER NOT NULL REFERENCES matches (id),
  admin_id      INTEGER NOT NULL REFERENCES admins (id),
  type          TEXT    NOT NULL,
  payload       TEXT    NOT NULL DEFAULT '{}',
  created_at    TEXT    NOT NULL
);

CREATE INDEX match_events_by_match ON match_events (match_id, created_at);

CREATE TABLE admin_events (
  id              INTEGER PRIMARY KEY,
  tournament_id   INTEGER NOT NULL REFERENCES tournaments (id),
  admin_id        INTEGER NOT NULL REFERENCES admins (id),
  target_admin_id INTEGER NOT NULL REFERENCES admins (id),
  type            TEXT    NOT NULL CHECK (type IN ('create', 'reset_pin', 'deactivate', 'reactivate')),
  created_at      TEXT    NOT NULL
);

CREATE INDEX admin_events_by_target ON admin_events (target_admin_id, created_at);

CREATE TABLE pending_actions (
  id            INTEGER PRIMARY KEY,
  tournament_id INTEGER NOT NULL REFERENCES tournaments (id),
  match_id      INTEGER NOT NULL REFERENCES matches (id),
  action        TEXT    NOT NULL CHECK (action IN ('pause', 'restart', 'abandon')),
  requested_by  INTEGER NOT NULL REFERENCES admins (id),
  approved_by   INTEGER REFERENCES admins (id),
  status        TEXT    NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
  created_at    TEXT    NOT NULL,
  resolved_at   TEXT
);

CREATE INDEX pending_actions_by_match ON pending_actions (match_id, status);

CREATE TABLE push_subscriptions (
  id            INTEGER PRIMARY KEY,
  tournament_id INTEGER NOT NULL REFERENCES tournaments (id),
  admin_id      INTEGER NOT NULL REFERENCES admins (id),
  endpoint      TEXT    NOT NULL UNIQUE,
  p256dh        TEXT    NOT NULL,
  auth          TEXT    NOT NULL,
  created_at    TEXT    NOT NULL,
  last_error    TEXT
);

CREATE INDEX push_subscriptions_by_admin ON push_subscriptions (admin_id);
