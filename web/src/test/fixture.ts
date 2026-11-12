import type {Match, Player, Snapshot} from "../api/types";

export function player(id: number, fullName: string, nickname: string | null, group = "A"): Player {
    return {id, full_name: fullName, nickname, avatar_url: null, group};
}

export function match(overrides: Partial<Match> & Pick<Match, "id" | "match_no">): Match {
    return {
        stage: "group",
        group: "A",
        match_day: 1,
        scheduled_date: "2026-11-11",
        slot_a: null,
        slot_b: null,
        player_a_id: 1,
        player_b_id: 2,
        score_a: 0,
        score_b: 0,
        status: "not_started",
        started_at: null,
        full_time_at: null,
        confirmed_at: null,
        version: 0,
        ...overrides,
    };
}

export function snapshot(overrides: Partial<Snapshot> = {}): Snapshot {
    return {
        tournament: {
            name: "Entelect Zim Foosball Cup",
            groups: ["A", "B"],
            rounds: 2,
            qualifiers: 2,
            race_to: 10,
            points_per_win: 1,
            match_days: ["2026-11-11", "2026-11-13"],
            group_stage_locked: false,
        },
        players: [
            player(1, "Trust Mandinyenya", "Malume"),
            player(2, "Thomas Nyagwaya", "Tomishto"),
            player(3, "Benard Mwanjeya", "Bento"),
            player(4, "Tapfuma Chigwada", "The Tapf", "B"),
            player(5, "Job Gwese", "Honey Badger", "B"),
        ],
        matches: [
            match({
                id: 1,
                match_no: 1,
                status: "confirmed",
                score_a: 10,
                score_b: 3,
                confirmed_at: "2026-11-11T10:00:00Z"
            }),
            match({id: 2, match_no: 2, player_a_id: 2, player_b_id: 3}),
            match({id: 3, match_no: 3, player_a_id: 1, player_b_id: 3}),
            match({
                id: 4,
                match_no: 4,
                group: "B",
                player_a_id: 4,
                player_b_id: 5,
                status: "confirmed",
                score_a: 10,
                score_b: 4,
                confirmed_at: "2026-11-11T11:00:00Z"
            }),
        ],
        standings: [
            {
                group: "A",
                rows: [
                    {
                        player_id: 1,
                        rank: 1,
                        played: 1,
                        won: 1,
                        lost: 0,
                        goals_for: 10,
                        goals_against: 3,
                        goal_difference: 7,
                        points: 1
                    },
                    {
                        player_id: 3,
                        rank: 2,
                        played: 0,
                        won: 0,
                        lost: 0,
                        goals_for: 0,
                        goals_against: 0,
                        goal_difference: 0,
                        points: 0
                    },
                    {
                        player_id: 2,
                        rank: 3,
                        played: 1,
                        won: 0,
                        lost: 1,
                        goals_for: 3,
                        goals_against: 10,
                        goal_difference: -7,
                        points: 0
                    },
                ],
            },
            {
                group: "B",
                rows: [
                    {
                        player_id: 4,
                        rank: 1,
                        played: 1,
                        won: 1,
                        lost: 0,
                        goals_for: 10,
                        goals_against: 4,
                        goal_difference: 6,
                        points: 1
                    },
                    {
                        player_id: 5,
                        rank: 2,
                        played: 1,
                        won: 0,
                        lost: 1,
                        goals_for: 4,
                        goals_against: 10,
                        goal_difference: -6,
                        points: 0
                    },
                ],
            },
        ],
        bracket: [],
        ...overrides,
    };
}
