import {dayStates, localDate} from "../lib/home";
import {HeaderNav} from "./nav";

const DAY_STYLE = {
    past: "bg-ink text-white",
    today: "bg-live text-on-live",
    future: "border border-border bg-surface text-muted",
};

export function Header({name, matchDays}: { name: string; matchDays: string[] }) {
    const days = dayStates(matchDays, localDate());

    return (
        <header
            className="relative flex flex-col gap-2.5 px-4 pb-3 pt-[18px] md:gap-3.5 md:px-8 md:pb-[18px] md:pt-6 xl:flex-row xl:items-center xl:gap-6 xl:px-0">
            <div className="flex items-center gap-2.5 md:gap-3">
                <img src="/logo.svg" alt="" className="size-[34px] md:size-10 xl:size-[42px]"/>
                <span className="font-display text-xl font-bold tracking-[-0.01em] md:text-[26px]">{name}</span>
            </div>
            <div className="flex items-center gap-4 xl:contents">
                <ul aria-label="Match days" className="flex gap-1.5 font-score text-xs font-bold">
                    {days.map((d) => (
                        <li key={d.date} title={d.date} className={`rounded-full px-3 py-[5px] ${DAY_STYLE[d.state]}`}>
                            {d.label}
                        </li>
                    ))}
                </ul>
                <HeaderNav/>
            </div>
        </header>
    );
}

/** Decorative arcs from the logo, partly off-screen top-right. */
export function ArcMotif() {
    return (
        <svg
            viewBox="0 0 300 300"
            fill="none"
            aria-hidden="true"
            className="pointer-events-none absolute -right-[150px] -top-[185px] size-[280px] md:-right-[150px] md:-top-[190px] md:size-[380px] xl:-top-[250px] xl:size-[440px]"
        >
            <circle cx="150" cy="150" r="120" stroke="var(--color-brand)" strokeWidth="22" strokeDasharray="420 400"
                    strokeLinecap="round"/>
            <circle cx="150" cy="150" r="84" stroke="var(--color-live)" strokeWidth="18" strokeDasharray="200 400"
                    strokeDashoffset="-150" strokeLinecap="round"/>
        </svg>
    );
}
