export type MatchStatus = "not_started" | "live" | "paused" | "provisional" | "confirmed" | "abandoned";

export interface TournamentInfo {
    name: string;
    groups: string[];
    rounds: number;
    qualifiers: number;
    race_to: number;
    points_per_win: number;
    match_days: string[];
    group_stage_locked: boolean;
}

export interface Player {
    id: number;
    full_name: string;
    nickname: string | null;
    avatar_url: string | null;
    group: string;
}

export interface Match {
    id: number;
    stage: string;
    group: string | null;
    match_no: number;
    match_day: number | null;
    scheduled_date: string | null;
    slot_a: string | null;
    slot_b: string | null;
    player_a_id: number | null;
    player_b_id: number | null;
    score_a: number;
    score_b: number;
    status: MatchStatus;
    started_at: string | null;
    full_time_at: string | null;
    confirmed_at: string | null;
    version: number;
}

export interface StandingRow {
    player_id: number;
    rank: number;
    played: number;
    won: number;
    lost: number;
    goals_for: number;
    goals_against: number;
    goal_difference: number;
    points: number;
}

export interface GroupStanding {
    group: string;
    rows: StandingRow[];
}

export interface Snapshot {
    tournament: TournamentInfo;
    players: Player[];
    matches: Match[];
    standings: GroupStanding[];
    bracket: Match[];
}
