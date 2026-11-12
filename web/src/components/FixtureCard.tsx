import type {Match, Player} from "../api/types";
import {displayName} from "../lib/home";

const FRAME: Record<Match["status"], string> = {
    not_started: "border border-border bg-surface",
    live: "border-[1.5px] border-live bg-live-tint",
    paused: "border border-border bg-surface",
    provisional: "border-[1.5px] border-dashed border-brand bg-surface",
    confirmed: "border border-border bg-surface",
    abandoned: "border border-border bg-surface text-faint",
};

const TAG: Record<Match["status"], { text: string; style: string } | null> = {
    not_started: null,
    live: {text: "LIVE", style: "text-live-text"},
    paused: {text: "PAUSED", style: "text-muted"},
    provisional: {text: "PROVISIONAL", style: "text-brand"},
    confirmed: {text: "FINAL", style: "text-muted"},
    abandoned: {text: "ABANDONED", style: "text-faint"},
};

function Side({player, slot, align, won}: {
    player: Player | undefined;
    slot: string | null;
    align: "start" | "end";
    won: boolean
}) {
    return (
        <span className={`flex min-w-0 flex-col ${align === "end" ? "items-end text-right" : ""}`}>
            <span className={`max-w-full truncate font-display text-[15px] ${won ? "font-bold" : "font-medium"}`}>
                {player ? displayName(player) : (slot ?? "TBC")}
            </span>
            {player?.nickname && <span className="max-w-full truncate text-[11px] text-muted">{player.full_name}</span>}
        </span>
    );
}

export function FixtureCard({match, players}: { match: Match; players: Map<number, Player> }) {
    const a = match.player_a_id != null ? players.get(match.player_a_id) : undefined;
    const b = match.player_b_id != null ? players.get(match.player_b_id) : undefined;
    const played = match.status !== "not_started";
    const decided = match.status === "confirmed";
    const tag = TAG[match.status];
    const nameA = a ? displayName(a) : (match.slot_a ?? "TBC");
    const nameB = b ? displayName(b) : (match.slot_b ?? "TBC");
    const label = `Match ${match.match_no}: ${nameA} v ${nameB}${played ? `, ${match.score_a} to ${match.score_b}` : ""}${tag ? `, ${tag.text.toLowerCase()}` : ""}`;

    return (
        <li aria-label={label}
            className={`flex flex-col gap-1.5 rounded-[14px] px-3.5 pb-3 pt-2.5 ${FRAME[match.status]}`}>
            <span aria-hidden="true" className="flex items-center gap-2 font-score text-[11px] font-bold">
                <span className="text-muted">#{match.match_no}{match.group ? ` · GROUP ${match.group}` : ""}</span>
                {tag && (
                    <span className={`ml-auto flex items-center gap-1.5 tracking-[0.06em] ${tag.style}`}>
                        {match.status === "live" && <span className="size-1.5 rounded-full bg-live"/>}
                        {tag.text}
                    </span>
                )}
            </span>
            <span aria-hidden="true" className="grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2.5">
                <Side player={a} slot={match.slot_a} align="start" won={decided && match.score_a > match.score_b}/>
                {played ? (
                    <span className="rounded-full bg-bg px-3 py-0.5 font-score text-base font-bold tabular-nums">
                        {match.score_a} : {match.score_b}
                    </span>
                ) : (
                    <span className="px-3 font-score text-sm font-bold text-faint">v</span>
                )}
                <Side player={b} slot={match.slot_b} align="end" won={decided && match.score_b > match.score_a}/>
            </span>
        </li>
    );
}
