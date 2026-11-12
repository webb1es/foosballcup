import {useSearchParams} from "react-router";
import type {Player} from "../api/types";
import {useTournament} from "../api/TournamentProvider";
import {FixtureCard} from "../components/FixtureCard";
import {
    filterByPlayer,
    type FixtureSection,
    fixtureSections,
    formatDay,
    KNOCKOUTS,
    parseSectionKey,
    type SectionKey
} from "../lib/fixtures";
import {localDate, playerIndex} from "../lib/home";

function sectionTitle(key: SectionKey): string {
    return key === KNOCKOUTS ? "Knockouts" : `Day ${key}`;
}

function SectionPicker({keys, selected, onSelect}: {
    keys: SectionKey[];
    selected: SectionKey | null;
    onSelect: (key: SectionKey | null) => void
}) {
    const options: { key: SectionKey | null; label: string }[] = [
        {key: null, label: "All"},
        ...keys.map((k) => ({key: k, label: sectionTitle(k)})),
    ];
    return (
        <div role="group" aria-label="Match day" className="flex gap-1 overflow-x-auto rounded-full bg-surface-2 p-1">
            {options.map((o) => (
                <button
                    key={String(o.key)}
                    type="button"
                    aria-pressed={o.key === selected}
                    onClick={() => onSelect(o.key)}
                    className={`min-h-10 grow whitespace-nowrap rounded-full px-4 text-[13px] font-semibold ${o.key === selected ? "bg-ink text-white" : "text-body"}`}
                >
                    {o.label}
                </button>
            ))}
        </div>
    );
}

function Section({section, players, today}: {
    section: FixtureSection;
    players: Map<number, Player>;
    today: string
}) {
    const id = `fixtures-${section.key}`;
    const done = section.matches.filter((m) => m.status === "confirmed").length;
    return (
        <section aria-labelledby={id} className="flex flex-col gap-2.5">
            <div className="flex items-baseline gap-2.5 px-1">
                <h2 id={id} className="font-display text-lg font-bold xl:text-xl">{sectionTitle(section.key)}</h2>
                {section.date && <span className="text-[13px] text-muted">{formatDay(section.date)}</span>}
                {section.date === today && (
                    <span
                        className="self-center rounded-full bg-live px-2.5 py-0.5 font-score text-[11px] font-bold text-on-live">TODAY</span>
                )}
                <span
                    className="ml-auto font-score text-xs font-bold text-muted">{done}/{section.matches.length} PLAYED</span>
            </div>
            <ul className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                {section.matches.map((m) => <FixtureCard key={m.id} match={m} players={players}/>)}
            </ul>
        </section>
    );
}

export default function Fixtures() {
    const {snapshot} = useTournament();
    const [params, setParams] = useSearchParams();
    if (!snapshot) return null;

    const query = params.get("q") ?? "";
    const players = playerIndex(snapshot.players);
    const all = fixtureSections([...snapshot.matches, ...snapshot.bracket], snapshot.tournament.match_days);
    const selected = parseSectionKey(params.get("day"));
    const sections = all
        .filter((s) => selected == null || s.key === selected)
        .map((s) => ({...s, matches: filterByPlayer(s.matches, players, query)}))
        .filter((s) => s.matches.length > 0);

    const update = (key: "q" | "day", value: string | null) => {
        setParams((prev) => {
            const next = new URLSearchParams(prev);
            if (value == null || value === "") next.delete(key);
            else next.set(key, value);
            return next;
        }, {replace: true});
    };

    return (
        <div className="flex flex-col gap-4 md:gap-5 xl:gap-6">
            <h1 className="px-1 font-display text-2xl font-bold md:text-[28px]">Fixtures</h1>
            <div className="flex flex-col gap-2.5 md:flex-row md:items-center md:gap-4">
                <label
                    className="flex min-h-11 items-center gap-2 rounded-full border border-border bg-surface px-4 md:w-80">
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"
                         strokeLinecap="round" aria-hidden="true" className="shrink-0 text-muted">
                        <circle cx="11" cy="11" r="7"/>
                        <path d="m20 20-3.5-3.5"/>
                    </svg>
                    <input
                        type="search"
                        value={query}
                        onChange={(e) => update("q", e.target.value)}
                        list="fixture-players"
                        placeholder="Player or nickname"
                        aria-label="Filter by player or nickname"
                        className="min-w-0 grow bg-transparent text-[15px] outline-none placeholder:text-faint"
                    />
                </label>
                <datalist id="fixture-players">
                    {snapshot.players.map((p) => <option key={p.id}
                                                         value={p.nickname ?? p.full_name}>{p.full_name}</option>)}
                </datalist>
                <div className="md:ml-auto">
                    <SectionPicker keys={all.map((s) => s.key)} selected={selected}
                                   onSelect={(k) => update("day", k == null ? null : String(k))}/>
                </div>
            </div>

            {sections.length === 0 ? (
                <p role="status" className="card-cut border border-border bg-surface px-4 py-5 text-sm text-body">
                    {query ? `No matches for "${query.trim()}".` : "No matches scheduled."}
                </p>
            ) : (
                sections.map((s) => <Section key={String(s.key)} section={s} players={players} today={localDate()}/>)
            )}
        </div>
    );
}
