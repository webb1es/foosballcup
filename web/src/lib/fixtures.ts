import type {Match, Player} from "../api/types";

export const KNOCKOUTS = "ko";

/** A section key is a group-stage match day number, or KNOCKOUTS. */
export type SectionKey = number | typeof KNOCKOUTS;

export interface FixtureSection {
    key: SectionKey;
    date: string | null;
    matches: Match[];
}

function sectionKey(m: Match): SectionKey {
    return m.stage === "group" && m.match_day != null ? m.match_day : KNOCKOUTS;
}

/** Group matches by match day in schedule order, with knockout matches in a final section. */
export function fixtureSections(matches: Match[], matchDays: string[]): FixtureSection[] {
    const sections = new Map<SectionKey, Match[]>();
    const rank = (k: SectionKey) => (k === KNOCKOUTS ? Number.MAX_SAFE_INTEGER : k);
    const sorted = [...matches].sort((a, b) => rank(sectionKey(a)) - rank(sectionKey(b)) || a.match_no - b.match_no);
    for (const m of sorted) {
        const key = sectionKey(m);
        sections.set(key, [...(sections.get(key) ?? []), m]);
    }
    return [...sections].map(([key, ms]) => ({
        key,
        date: key === KNOCKOUTS ? null : (matchDays[key - 1] ?? null),
        matches: ms,
    }));
}

/** Matches where either player's nickname or full name contains the query, ignoring case. */
export function filterByPlayer(matches: Match[], players: Map<number, Player>, query: string): Match[] {
    const needle = query.trim().toLowerCase();
    if (needle === "") return matches;
    const hit = (id: number | null) => {
        const p = id != null ? players.get(id) : undefined;
        return p != null && (p.full_name.toLowerCase().includes(needle) || (p.nickname ?? "").toLowerCase().includes(needle));
    };
    return matches.filter((m) => hit(m.player_a_id) || hit(m.player_b_id));
}

export function parseSectionKey(value: string | null): SectionKey | null {
    if (value === KNOCKOUTS) return KNOCKOUTS;
    const day = Number(value);
    return Number.isInteger(day) && day > 0 ? day : null;
}

const DAY_FORMAT = new Intl.DateTimeFormat("en-GB", {
    weekday: "short",
    day: "numeric",
    month: "short",
    timeZone: "UTC"
});

/** "2026-11-11" as "Wed 11 Nov". */
export function formatDay(date: string): string {
    return DAY_FORMAT.format(new Date(`${date}T00:00:00Z`));
}
