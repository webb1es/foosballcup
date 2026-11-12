import type {Match, Player} from "../api/types";
import {displayName, initials, upNextTag} from "../lib/home";

function AvatarPair({a, b}: { a: Player | undefined; b: Player | undefined }) {
    const base = "flex size-8 items-center justify-center rounded-full border-2 border-white text-[11px] font-bold text-white xl:size-[34px]";
    return (
        <span className="flex shrink-0" aria-hidden="true">
      <span className={`${base} bg-brand`}>{initials(a)}</span>
      <span className={`${base} -ml-2.5 bg-ink`}>{initials(b)}</span>
    </span>
    );
}

function meta(m: Match): string {
    return `${m.group ? `Group ${m.group} · ` : ""}Match ${m.match_no}`;
}

export function UpNext({matches, players, today}: { matches: Match[]; players: Map<number, Player>; today: string }) {
    if (matches.length === 0) return null;
    const items = matches.map((m, i) => {
        const a = m.player_a_id != null ? players.get(m.player_a_id) : undefined;
        const b = m.player_b_id != null ? players.get(m.player_b_id) : undefined;
        return {m, a, b, title: `${displayName(a)} v ${displayName(b)}`, tag: upNextTag(m, i, today)};
    });

    return (
        <section aria-labelledby="up-next" className="flex flex-col gap-2.5">
            <h2 id="up-next" className="sr-only font-display text-base font-bold md:not-sr-only xl:text-[17px]">
                Up next
            </h2>
            <ul className="card-cut flex flex-col border border-border bg-surface px-3.5 py-1 md:px-4 xl:hidden">
                {items.map(({m, a, b, title, tag}, i) => (
                    <li key={m.id}
                        className={`flex min-h-11 items-center gap-3 py-2.5 ${i < items.length - 1 ? "border-b border-surface-2" : ""}`}>
                        <AvatarPair a={a} b={b}/>
                        <span className="flex min-w-0 grow flex-col">
              <span className="truncate font-display text-[15px] font-bold">{title}</span>
              <span className="text-[11px] text-muted">{meta(m)}</span>
            </span>
                        <span className="whitespace-nowrap font-score text-xs font-bold text-brand">{tag}</span>
                    </li>
                ))}
            </ul>
            <ul className="hidden grid-cols-3 gap-3.5 xl:grid">
                {items.map(({m, a, b, title, tag}) => (
                    <li key={m.id}
                        className="flex flex-col gap-2.5 rounded-[20px_20px_20px_6px] border border-border bg-surface p-3.5">
            <span className="flex items-center justify-between">
              <AvatarPair a={a} b={b}/>
              <span className="whitespace-nowrap font-score text-xs font-bold text-brand">{tag}</span>
            </span>
                        <span className="flex flex-col">
              <span className="truncate font-display text-base font-bold">{title}</span>
              <span className="text-xs text-muted">{meta(m)}</span>
            </span>
                    </li>
                ))}
            </ul>
        </section>
    );
}

export function LatestResults({matches, players}: { matches: Match[]; players: Map<number, Player> }) {
    if (matches.length === 0) return null;
    return (
        <section aria-labelledby="latest-results"
                 className="card-cut hidden border border-border bg-surface px-4 pb-1.5 pt-3.5 md:block xl:px-5">
            <h2 id="latest-results" className="mb-1 font-display text-base font-bold xl:text-[17px]">
                Latest results
            </h2>
            <ul>
                {matches.map((m, i) => {
                    const a = m.player_a_id != null ? players.get(m.player_a_id) : undefined;
                    const b = m.player_b_id != null ? players.get(m.player_b_id) : undefined;
                    return (
                        <li
                            key={m.id}
                            className={`grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2.5 py-2.5 text-sm xl:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)_150px] xl:gap-3.5 ${
                                i < matches.length - 1 ? "border-b border-surface-2" : ""
                            }`}
                        >
                            <span
                                className={`truncate xl:text-right ${m.score_a > m.score_b ? "font-bold" : ""}`}>{displayName(a)}</span>
                            <span
                                className="rounded-full bg-bg px-3 py-0.5 font-score text-base font-bold tabular-nums">
                {m.score_a} : {m.score_b}
              </span>
                            <span
                                className={`truncate text-right xl:text-left ${m.score_b > m.score_a ? "font-bold" : ""}`}>{displayName(b)}</span>
                            <span className="hidden text-right text-xs text-muted xl:block">{meta(m)}</span>
                        </li>
                    );
                })}
            </ul>
        </section>
    );
}
