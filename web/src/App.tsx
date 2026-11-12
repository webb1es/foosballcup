import {lazy, Suspense} from "react";
import {BrowserRouter, Route, Routes} from "react-router";
import {TournamentProvider, useTournament} from "./api/TournamentProvider";
import {ArcMotif, Header} from "./components/Header";
import {BottomNav} from "./components/nav";

const Home = lazy(() => import("./pages/Home"));
const NotBuilt = lazy(() => import("./pages/NotBuilt"));

function Shell() {
    const {snapshot, error} = useTournament();
    const name = snapshot?.tournament.name ?? "Entelect Zim Foosball Cup";

    return (
        <div className="relative min-h-dvh overflow-x-clip">
            <ArcMotif/>
            <div className="relative mx-auto max-w-[1344px] xl:px-10">
                <Header name={name} matchDays={snapshot?.tournament.match_days ?? []}/>
                <main className="px-4 pb-32 pt-1 md:px-8 md:pb-8 xl:px-0 xl:pb-10">
                    {error && (
                        <p role="status" className="mb-3 rounded-xl bg-surface px-4 py-3 text-sm text-body">
                            Can't reach the server. Retrying every 15 seconds.
                        </p>
                    )}
                    {!snapshot && !error && <p className="text-sm text-muted">Loading…</p>}
                    <Suspense fallback={null}>
                        <Routes>
                            <Route path="/" element={<Home/>}/>
                            <Route path="*" element={<NotBuilt/>}/>
                        </Routes>
                    </Suspense>
                </main>
            </div>
            <BottomNav/>
        </div>
    );
}

export function App() {
    return (
        <BrowserRouter>
            <TournamentProvider>
                <Shell/>
            </TournamentProvider>
        </BrowserRouter>
    );
}
