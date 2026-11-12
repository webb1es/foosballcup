import {useState} from "react";
import {useTournament} from "../api/TournamentProvider";
import {GroupSwitch} from "../components/GroupSwitch";
import {StandingsTable} from "../components/StandingsTable";
import {playerIndex} from "../lib/home";
import {recentForm} from "../lib/standings";

const FORM_COUNT = 5;

export default function Standings() {
    const {snapshot} = useTournament();
    const [group, setGroup] = useState<string | null>(null);
    if (!snapshot) return null;

    const {tournament, standings} = snapshot;
    const players = playerIndex(snapshot.players);
    const form = recentForm(snapshot.matches, FORM_COUNT);
    const selected = group ?? tournament.groups[0];

    return (
        <div className="flex flex-col gap-3 md:gap-5 xl:gap-6">
            <div className="flex flex-col gap-1">
                <h1 className="font-display text-2xl font-bold md:text-[28px]">Standings</h1>
                <p className="text-[13px] text-muted md:text-sm">
                    Top {tournament.qualifiers} in each group go through. Ranked by points, goal difference, goals
                    scored, then head-to-head. Only confirmed results count.
                </p>
            </div>
            <GroupSwitch groups={tournament.groups} selected={selected} onSelect={setGroup}/>
            <div className="grid gap-5 xl:grid-cols-2 xl:gap-6">
                {standings.map((s) => (
                    <div key={s.group} className={s.group === selected ? "" : "hidden md:block"}>
                        <StandingsTable standing={s} players={players} qualifiers={tournament.qualifiers} form={form}/>
                    </div>
                ))}
            </div>
        </div>
    );
}
