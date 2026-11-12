import type {Match, Player} from "../api/types";
import {displayName, initials} from "../lib/home";

type Mode = "live" | "paused" | "provisional" | "next";

const RING_OUTER_R = 56;
const RING_INNER_R = 43;

function ringDash(radius: number, score: number, raceTo: number): string {
    const circumference = 2 * Math.PI * radius;
    const filled = (Math.min(score, raceTo) / raceTo) * circumference;
    return `${filled} ${circumference}`;
}

function Rings({scoreA, scoreB, raceTo, nameA, nameB}: {
    scoreA: number;
    scoreB: number;
    raceTo: number;
    nameA: string;
    nameB: string
}) {
    return (
        <div className="relative size-32 shrink-0 md:size-[168px] xl:size-[220px]">
            <svg viewBox="0 0 132 132" fill="none" className="size-full -rotate-90" role="img"
                 aria-label={`Progress to ${raceTo}: ${nameA} ${scoreA}, ${nameB} ${scoreB}`}>
                <circle cx="66" cy="66" r="62" fill="var(--color-night)"/>
                <circle cx="66" cy="66" r={RING_OUTER_R} stroke="var(--color-night-line)" strokeWidth="8"/>
                {scoreA > 0 && <circle cx="66" cy="66" r={RING_OUTER_R} stroke="var(--color-live)" strokeWidth="8"
                                       strokeLinecap="round" strokeDasharray={ringDash(RING_OUTER_R, scoreA, raceTo)}/>}
                <circle cx="66" cy="66" r={RING_INNER_R} stroke="var(--color-night-line)" strokeWidth="8"/>
                {scoreB > 0 && <circle cx="66" cy="66" r={RING_INNER_R} stroke="var(--color-ring-b)" strokeWidth="8"
                                       strokeLinecap="round" strokeDasharray={ringDash(RING_INNER_R, scoreB, raceTo)}/>}
            </svg>
            <span aria-hidden="true"
                  className="absolute inset-0 flex items-center justify-center gap-1 font-score text-[34px] font-bold tabular-nums md:text-[46px] xl:text-[60px]">
        <span className={scoreA >= scoreB ? "text-white" : "text-on-night-muted"}>{scoreA}</span>
        <span className="text-[#2f4475]">:</span>
        <span className={scoreB > scoreA ? "text-white" : "text-on-night-muted"}>{scoreB}</span>
      </span>
        </div>
    );
}

function Side({player, align, ring}: { player: Player | undefined; align: "start" | "end"; ring: string }) {
    const end = align === "end";
    return (
        <div
            className={`flex min-w-0 flex-col gap-1.5 md:flex-row md:items-center md:gap-3.5 xl:flex-col xl:items-start xl:gap-2.5 ${end ? "items-end text-right md:flex-row-reverse xl:items-end" : ""}`}>
      <span
          className={`flex size-10 shrink-0 items-center justify-center rounded-full border-[3px] bg-white text-[13px] font-bold text-night md:size-14 md:text-[17px] xl:size-16 xl:border-4 xl:text-[19px] ${ring}`}>
        {initials(player)}
      </span>
            <span className="flex min-w-0 flex-col">
        <span
            className="truncate font-display text-base font-bold md:text-[22px] xl:text-[28px] xl:leading-tight">{displayName(player)}</span>
        <span
            className={`truncate text-[11px] md:text-xs xl:text-[13px] ${end ? "text-on-night-muted" : "text-[#c9d4ee]"}`}>{player?.full_name ?? ""}</span>
      </span>
        </div>
    );
}

const LABEL: Record<Mode, string> = {live: "LIVE", paused: "PAUSED", provisional: "PROVISIONAL", next: "NEXT MATCH"};

export function LiveCard({match, players, raceTo, mode}: {
    match: Match;
    players: Map<number, Player>;
    raceTo: number;
    mode: Mode
}) {
    const a = match.player_a_id != null ? players.get(match.player_a_id) : undefined;
    const b = match.player_b_id != null ? players.get(match.player_b_id) : undefined;
    const border = mode === "provisional" ? "border-[1.5px] border-dashed border-brand" : mode === "next" ? "border border-night-line" : "border-[1.5px] border-live";
    const toWin = raceTo - Math.max(match.score_a, match.score_b);

    return (
        <section aria-label={`${LABEL[mode]}: match ${match.match_no}`}
                 className={`card-cut-lg relative overflow-hidden bg-night text-white xl:rounded-[32px_32px_32px_8px] ${border}`}>
            <div aria-hidden="true" className="absolute inset-y-0 left-0 w-1/2 bg-night-2"/>
            <div
                className="relative grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-1 px-3 pb-3 pt-4 md:gap-4 md:px-6 md:pb-4 md:pt-5 xl:gap-5 xl:px-7 xl:pt-6">
                <Side player={a} align="start" ring="border-live"/>
                <Rings scoreA={match.score_a} scoreB={match.score_b} raceTo={raceTo} nameA={displayName(a)}
                       nameB={displayName(b)}/>
                <Side player={b} align="end" ring="border-ring-b"/>
            </div>
            <div
                className="relative flex items-center gap-2 bg-night px-4 py-2.5 text-xs text-on-night-muted md:px-6 md:py-3 md:text-[13px] xl:px-7"
                aria-live="polite">
                {mode !== "next" &&
                    <span className={`size-2 rounded-full ${mode === "live" ? "bg-live" : "bg-on-night-muted"}`}/>}
                <span
                    className={`font-extrabold tracking-[0.1em] ${mode === "live" ? "text-live" : "text-white"}`}>{LABEL[mode]}</span>
                <span>
          {match.group ? `Group ${match.group} · ` : ""}Match {match.match_no}
        </span>
                <span className="ml-auto">
          {mode === "live" && toWin > 0 ? `Race to ${raceTo} · ${toWin} from the win` : `Race to ${raceTo}`}
        </span>
            </div>
        </section>
    );
}
