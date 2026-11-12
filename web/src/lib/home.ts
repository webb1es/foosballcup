import type {Match, Player} from "../api/types";

const ACTIVE_ORDER: Partial<Record<Match["status"], number>> = {live: 0, paused: 1, provisional: 2};

/** Matches the public should see first: live, then paused, then awaiting confirmation. */
export function activeMatches(matches: Match[]): Match[] {
    return matches
        .filter((m) => m.status in ACTIVE_ORDER)
        .sort((a, b) => ACTIVE_ORDER[a.status]! - ACTIVE_ORDER[b.status]! || a.match_no - b.match_no);
}

/** Unplayed matches in schedule order. Matches left over from earlier days come first. */
export function upcomingMatches(matches: Match[]): Match[] {
    return matches
        .filter((m) => m.status === "not_started")
        .sort((a, b) => (a.match_day ?? 0) - (b.match_day ?? 0) || a.match_no - b.match_no);
}

export function latestResults(matches: Match[], count: number): Match[] {
    return matches
        .filter((m) => m.status === "confirmed")
        .sort((a, b) => (b.confirmed_at ?? "").localeCompare(a.confirmed_at ?? "") || b.match_no - a.match_no)
        .slice(0, count);
}

export type DayState = "past" | "today" | "future";

export function dayStates(days: string[], today: string): { label: string; date: string; state: DayState }[] {
    return days.map((date, i) => ({
        label: `DAY ${i + 1}`,
        date,
        state: date < today ? "past" : date === today ? "today" : "future",
    }));
}

/** Local calendar date as YYYY-MM-DD, matching the API's scheduled_date format. */
export function localDate(now: Date = new Date()): string {
    const pad = (n: number) => String(n).padStart(2, "0");
    return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`;
}

const MONTHS = ["JAN", "FEB", "MAR", "APR", "MAY", "JUN", "JUL", "AUG", "SEP", "OCT", "NOV", "DEC"];

export function upNextTag(match: Match, index: number, today: string): string {
    if (index === 0) return "NEXT";
    if (!match.scheduled_date || match.scheduled_date === today) return "TODAY";
    const [, month, day] = match.scheduled_date.split("-").map(Number);
    return `${day} ${MONTHS[month - 1]}`;
}

export function displayName(player: Player | undefined): string {
    if (!player) return "TBC";
    return player.nickname ?? player.full_name.split(" ")[0];
}

export function initials(player: Player | undefined): string {
    if (!player) return "?";
    const parts = player.full_name.trim().split(/\s+/);
    return (parts[0][0] + (parts.length > 1 ? parts[parts.length - 1][0] : "")).toUpperCase();
}

export function playerIndex(players: Player[]): Map<number, Player> {
    return new Map(players.map((p) => [p.id, p]));
}
