import {useState} from "react";
import {useTournament} from "../api/TournamentProvider";
import {GroupSwitch} from "../components/GroupSwitch";
import {LiveCard} from "../components/LiveCard";
import {LatestResults, UpNext} from "../components/MatchLists";
import {StandingsTable} from "../components/StandingsTable";
import {activeMatches, latestResults, localDate, playerIndex, upcomingMatches} from "../lib/home";

const UP_NEXT_COUNT = 3;
const RESULTS_COUNT = 3;

export default function Home() {
    const {snapshot} = useTournament();
    const [group, setGroup] = useState<string | null>(null);
    if (!snapshot) return null;

    const {tournament, matches, standings} = snapshot;
    const players = playerIndex(snapshot.players);
    const today = localDate();
    const active = activeMatches(matches);
    const upcoming = upcomingMatches(matches);
    const hero = active.length > 0 ? active : upcoming.slice(0, 1);
    const upNext = upcoming.slice(active.length > 0 ? 0 : 1, (active.length > 0 ? 0 : 1) + UP_NEXT_COUNT);
    const selected = group ?? tournament.groups[0];

    return (
        <div className="home-grid">
            <div className="flex flex-col gap-3.5 [grid-area:live]">
                {hero.map((m) => (
                    <LiveCard key={m.id} match={m} players={players} raceTo={tournament.race_to}
                              mode={m.status === "not_started" ? "next" : (m.status as "live" | "paused" | "provisional")}/>
                ))}
            </div>

            <div className="flex flex-col gap-4 [grid-area:next] md:gap-5 xl:gap-6">
                <UpNext matches={upNext} players={players} today={today}/>
                <LatestResults matches={latestResults(matches, RESULTS_COUNT)} players={players}/>
            </div>

            <div className="flex flex-col gap-3 [grid-area:tables] md:gap-5 xl:gap-6">
                <GroupSwitch groups={tournament.groups} selected={selected} onSelect={setGroup}/>
                {standings.map((s) => (
                    <div key={s.group} className={s.group === selected ? "" : "hidden md:block"}>
                        <StandingsTable standing={s} players={players} qualifiers={tournament.qualifiers}/>
                    </div>
                ))}
            </div>
        </div>
    );
}
