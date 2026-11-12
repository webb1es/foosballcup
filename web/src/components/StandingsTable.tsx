import type {GroupStanding, Player} from "../api/types";
import type {Outcome} from "../lib/standings";

function rankColour(rank: number, qualifiers: number): string {
    if (rank === 1) return "text-live-text";
    return rank <= qualifiers ? "text-brand" : "text-faint";
}

function signed(n: number): string {
    return n > 0 ? `+${n}` : n < 0 ? `−${-n}` : "0";
}

function FormGuide({outcomes}: { outcomes: Outcome[] }) {
    if (outcomes.length === 0) return <span className="text-faint">No results</span>;
    return (
        <span className="inline-flex gap-0.5" aria-label={`Last ${outcomes.length}: ${outcomes.join(" ")}`}>
            {outcomes.map((o, i) => (
                <span key={i} aria-hidden="true"
                      className={`flex size-[18px] items-center justify-center rounded-full font-score text-[10px] font-bold ${o === "W" ? "bg-brand text-white" : "bg-surface-2 text-muted"}`}>
                    {o}
                </span>
            ))}
        </span>
    );
}

/**
 * Compact on Home (P/W/L from tablet up). `form` switches to the detailed Standings page layout:
 * P/W/L always, GF, GA and form from tablet up, full name under the nickname.
 */
export function StandingsTable({standing, players, qualifiers, form}: {
    standing: GroupStanding;
    players: Map<number, Player>;
    qualifiers: number;
    form?: Map<number, Outcome[]>;
}) {
    const detailed = form != null;
    const pwl = detailed ? "table-cell" : "hidden md:table-cell";
    const num = "py-1 text-right font-semibold";

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
                    <th scope="col" className={`w-7 ${num} ${pwl}`}><abbr title="Played"
                                                                          className="no-underline">P</abbr></th>
                    <th scope="col" className={`w-7 ${num} ${pwl}`}><abbr title="Won" className="no-underline">W</abbr>
                    </th>
                    <th scope="col" className={`w-7 ${num} ${pwl}`}><abbr title="Lost" className="no-underline">L</abbr>
                    </th>
                    {detailed && <>
                        <th scope="col" className={`hidden w-9 md:table-cell ${num}`}><abbr title="Goals for"
                                                                                            className="no-underline">GF</abbr>
                        </th>
                        <th scope="col" className={`hidden w-9 md:table-cell ${num}`}><abbr title="Goals against"
                                                                                            className="no-underline">GA</abbr>
                        </th>
                    </>}
                    <th scope="col" className={`w-11 ${num}`}><abbr title="Goal difference"
                                                                    className="no-underline">GD</abbr></th>
                    <th scope="col" className={`w-10 pr-2.5 ${num}`}><abbr title="Points"
                                                                           className="no-underline">Pts</abbr></th>
                    {detailed && <th scope="col"
                                     className="hidden w-[118px] py-1 pr-2.5 text-left font-semibold md:table-cell">Form</th>}
                </tr>
                </thead>
                <tbody>
                {standing.rows.map((row, i) => {
                    const player = players.get(row.player_id);
                    const lastQualifier = i === qualifiers - 1 && i < standing.rows.length - 1;
                    const cell = detailed ? "py-2" : "py-1.5";
                    return (
                        <tr
                            key={row.player_id}
                            data-qualifying-line={lastQualifier || undefined}
                            className={lastQualifier ? "border-b-[1.5px] border-dashed border-brand" : i < standing.rows.length - 1 ? "border-b border-bg" : ""}
                        >
                            <td className={`${cell} pl-2.5 font-score font-bold ${rankColour(row.rank, qualifiers)}`}>{row.rank}</td>
                            <td className={`truncate ${cell}`}>
                                <span className="font-semibold">{player?.nickname ?? player?.full_name}</span>
                                {player?.nickname && (
                                    <span
                                        className={detailed ? "block truncate text-[11px] text-muted xl:text-xs" : "ml-1.5 text-[11px] text-muted"}>
                                        {player.full_name}
                                    </span>
                                )}
                            </td>
                            <td className={`${cell} text-right ${pwl}`}>{row.played}</td>
                            <td className={`${cell} text-right ${pwl}`}>{row.won}</td>
                            <td className={`${cell} text-right ${pwl}`}>{row.lost}</td>
                            {detailed && <>
                                <td className={`hidden ${cell} text-right md:table-cell`}>{row.goals_for}</td>
                                <td className={`hidden ${cell} text-right md:table-cell`}>{row.goals_against}</td>
                            </>}
                            <td className={`${cell} text-right text-muted`}>{signed(row.goal_difference)}</td>
                            <td className={`${cell} pr-2.5 text-right font-score text-[15px] font-bold xl:text-base`}>{row.points}</td>
                            {detailed && (
                                <td className={`hidden ${cell} pr-2.5 md:table-cell`}>
                                    <FormGuide outcomes={form.get(row.player_id) ?? []}/>
                                </td>
                            )}
                        </tr>
                    );
                })}
                </tbody>
            </table>
        </section>
    );
}
