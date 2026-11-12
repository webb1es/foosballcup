import type {Match} from "../api/types";

export type Outcome = "W" | "L";

/** Last `count` confirmed results per player, oldest first. */
export function recentForm(matches: Match[], count: number): Map<number, Outcome[]> {
    const form = new Map<number, Outcome[]>();
    const confirmed = matches
        .filter((m) => m.status === "confirmed" && m.player_a_id != null && m.player_b_id != null)
        .sort((a, b) => (a.confirmed_at ?? "").localeCompare(b.confirmed_at ?? "") || a.match_no - b.match_no);
    for (const m of confirmed) {
        const aWon = m.score_a > m.score_b;
        form.set(m.player_a_id!, [...(form.get(m.player_a_id!) ?? []), aWon ? "W" : "L"]);
        form.set(m.player_b_id!, [...(form.get(m.player_b_id!) ?? []), aWon ? "L" : "W"]);
    }
    for (const [id, outcomes] of form) form.set(id, outcomes.slice(-count));
    return form;
}
