/** Phone-only pill switch between groups. Tablet and desktop show every group at once. */
export function GroupSwitch({groups, selected, onSelect}: {
    groups: string[];
    selected: string;
    onSelect: (group: string) => void
}) {
    if (groups.length < 2) return null;
    return (
        <div role="group" aria-label="Group"
             className="grid gap-1 rounded-full bg-surface-2 p-1 md:hidden"
             style={{gridTemplateColumns: `repeat(${groups.length}, minmax(0, 1fr))`}}>
            {groups.map((g) => (
                <button
                    key={g}
                    type="button"
                    aria-pressed={g === selected}
                    onClick={() => onSelect(g)}
                    className={`min-h-10 rounded-full text-[13px] font-semibold ${g === selected ? "bg-ink text-white" : "text-body"}`}
                >
                    Group {g}
                </button>
            ))}
        </div>
    );
}
