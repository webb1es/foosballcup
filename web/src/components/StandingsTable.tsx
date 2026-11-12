import type {GroupStanding, Player} from "../api/types";

function rankColour(rank: number, qualifiers: number): string {
    if (rank === 1) return "text-live-text";
    return rank <= qualifiers ? "text-brand" : "text-faint";
}

function signed(n: number): string {
    return n > 0 ? `+${n}` : n < 0 ? `−${-n}` : "0";
}

export function StandingsTable({standing, players, qualifiers}: {
    standing: GroupStanding;
    players: Map<number, Player>;
    qualifiers: number
}) {
    return (
        <section aria-labelledby={`group-${standing.group}`}
                 className="card-cut border border-border bg-surface px-1.5 pb-1.5 pt-3.5 xl:px-2 xl:pt-4">
            <div className="flex items-baseline px-2.5 pb-1.5 xl:px-3 xl:pb-2">
                <h2 id={`group-${standing.group}`} className="grow font-display text-base font-bold xl:text-[17px]">
                    Group {standing.group}
                </h2>
                <span className="text-[11px] text-muted xl:text-xs">Top {qualifiers} qualify</span>
            </div>
            <table className="w-full table-fixed border-collapse text-[13px] tabular-nums xl:text-sm">
                <thead>
                <tr className="border-b border-surface-2 text-[11px] font-semibold text-muted">
                    <th scope="col" className="w-[30px] py-1 pl-2.5 text-left font-semibold">#</th>
                    <th scope="col" className="py-1 text-left font-semibold">Player</th>
                    <th scope="col" className="hidden w-7 py-1 text-right font-semibold md:table-cell">P</th>
                    <th scope="col" className="hidden w-7 py-1 text-right font-semibold md:table-cell">W</th>
                    <th scope="col" className="hidden w-7 py-1 text-right font-semibold md:table-cell">L</th>
                    <th scope="col" className="w-11 py-1 text-right font-semibold">GD</th>
                    <th scope="col" className="w-10 py-1 pr-2.5 text-right font-semibold">Pts</th>
                </tr>
                </thead>
                <tbody>
                {standing.rows.map((row, i) => {
                    const player = players.get(row.player_id);
                    const lastQualifier = i === qualifiers - 1 && i < standing.rows.length - 1;
                    return (
                        <tr
                            key={row.player_id}
                            data-qualifying-line={lastQualifier || undefined}
                            className={lastQualifier ? "border-b-[1.5px] border-dashed border-brand" : i < standing.rows.length - 1 ? "border-b border-bg" : ""}
                        >
                            <td className={`py-1.5 pl-2.5 font-score font-bold ${rankColour(row.rank, qualifiers)}`}>{row.rank}</td>
                            <td className="truncate py-1.5">
                                <span className="font-semibold">{player?.nickname ?? player?.full_name}</span>
                                {player?.nickname &&
                                    <span className="ml-1.5 text-[11px] text-muted">{player.full_name}</span>}
                            </td>
                            <td className="hidden py-1.5 text-right md:table-cell">{row.played}</td>
                            <td className="hidden py-1.5 text-right md:table-cell">{row.won}</td>
                            <td className="hidden py-1.5 text-right md:table-cell">{row.lost}</td>
                            <td className="py-1.5 text-right text-muted">{signed(row.goal_difference)}</td>
                            <td className="py-1.5 pr-2.5 text-right font-score text-[15px] font-bold xl:text-base">{row.points}</td>
                        </tr>
                    );
                })}
                </tbody>
            </table>
        </section>
    );
}
