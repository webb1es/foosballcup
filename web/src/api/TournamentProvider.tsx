import {createContext, useContext, useEffect, useState, type ReactNode} from "react";
import type {Snapshot} from "./types";

const POLL_INTERVAL_MS = 15_000;

interface TournamentState {
    snapshot: Snapshot | null;
    error: boolean;
}

const TournamentContext = createContext<TournamentState>({snapshot: null, error: false});

export function TournamentProvider({children}: { children: ReactNode }) {
    const [state, setState] = useState<TournamentState>({snapshot: null, error: false});

    useEffect(() => {
        let controller: AbortController | null = null;

        const refresh = async () => {
            controller?.abort();
            controller = new AbortController();
            try {
                const res = await fetch("/api/tournament", {signal: controller.signal});
                if (!res.ok) throw new Error(`status ${res.status}`);
                const snapshot = (await res.json()) as Snapshot;
                setState({snapshot, error: false});
            } catch (err) {
                if (err instanceof DOMException && err.name === "AbortError") return;
                setState((prev) => ({...prev, error: true}));
            }
        };

        const onVisible = () => {
            if (document.visibilityState === "visible") void refresh();
        };

        void refresh();
        const timer = window.setInterval(refresh, POLL_INTERVAL_MS);
        document.addEventListener("visibilitychange", onVisible);
        return () => {
            window.clearInterval(timer);
            document.removeEventListener("visibilitychange", onVisible);
            controller?.abort();
        };
    }, []);

    return <TournamentContext.Provider value={state}>{children}</TournamentContext.Provider>;
}

export function useTournament(): TournamentState {
    return useContext(TournamentContext);
}
