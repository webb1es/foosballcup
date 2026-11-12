import {NavLink} from "react-router";

const ITEMS = [
    {to: "/", label: "Home", icon: "M3 11l9-7 9 7v9a1 1 0 0 1-1 1h-5v-6h-6v6H4a1 1 0 0 1-1-1z"},
    {to: "/fixtures", label: "Fixtures", icon: "M4 5h16v15H4zM4 10h16M9 3v4M15 3v4"},
    {to: "/standings", label: "Tables", icon: "M4 6h16M4 12h16M4 18h16"},
    {to: "/bracket", label: "Bracket", icon: "M4 5h5v6H4M4 13h5v6H4M9 8h4v8h-4M13 12h7"},
];

const pill = ({isActive}: { isActive: boolean }) =>
    `flex min-h-10 items-center rounded-full px-[18px] text-[13px] font-semibold xl:px-5 ${
        isActive ? "bg-live text-on-live" : "text-[#c3cbdd] hover:text-white"
    }`;

/** Tablet and desktop: pill tabs in the header. */
export function HeaderNav() {
    return (
        <nav aria-label="Main" className="ml-auto hidden gap-1 rounded-full bg-night p-1 md:flex">
            {ITEMS.map((item) => (
                <NavLink key={item.to} to={item.to} end={item.to === "/"} className={pill}>
                    {item.label}
                </NavLink>
            ))}
        </nav>
    );
}

/** Phone: floating tab bar at the bottom of the screen. */
export function BottomNav() {
    return (
        <nav
            aria-label="Main"
            className="card-cut fixed inset-x-4 bottom-[max(20px,env(safe-area-inset-bottom))] z-20 grid grid-cols-4 gap-1 bg-night p-1.5 md:hidden"
        >
            {ITEMS.map((item) => (
                <NavLink
                    key={item.to}
                    to={item.to}
                    end={item.to === "/"}
                    className={({isActive}) =>
                        `flex min-h-[50px] flex-col items-center justify-center gap-0.5 rounded-[18px] text-[11px] font-semibold ${
                            isActive ? "bg-live text-on-live" : "text-[#c3cbdd]"
                        }`
                    }
                >
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9"
                         strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">
                        <path d={item.icon}/>
                    </svg>
                    {item.label}
                </NavLink>
            ))}
        </nav>
    );
}
