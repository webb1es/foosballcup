import {render, screen, within} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {describe, expect, it, vi} from "vitest";
import {App} from "./App";
import type {Snapshot} from "./api/types";
import {match, snapshot} from "./test/fixture";

function serve(snap: Snapshot) {
    vi.stubGlobal(
        "fetch",
        vi.fn(async () => new Response(JSON.stringify(snap), {
            status: 200,
            headers: {"Content-Type": "application/json"}
        })),
    );
}

describe("Home page", () => {
    it("shows the next match when nothing is live, and the rest under Up next", async () => {
        serve(snapshot());
        render(<App/>);

        const hero = await screen.findByRole("region", {name: "NEXT MATCH: match 2"});
        expect(within(hero).getByText("Tomishto")).toBeInTheDocument();
        expect(within(hero).getByText("Bento")).toBeInTheDocument();
        expect(screen.getAllByText("Malume v Bento").length).toBeGreaterThan(0);
    });

    it("shows a live match with its score in the live card", async () => {
        const snap = snapshot();
        snap.matches.push(match({
            id: 9,
            match_no: 9,
            player_a_id: 2,
            player_b_id: 3,
            status: "live",
            score_a: 7,
            score_b: 4
        }));
        serve(snap);
        render(<App/>);

        const live = await screen.findByRole("region", {name: "LIVE: match 9"});
        expect(within(live).getByRole("img", {name: "Progress to 10: Tomishto 7, Bento 4"})).toBeInTheDocument();
        expect(within(live).getByText("Race to 10 · 3 from the win")).toBeInTheDocument();
    });

    it("marks the qualifying line under the last qualifying place", async () => {
        serve(snapshot());
        render(<App/>);

        const table = await screen.findByRole("region", {name: "Group A"});
        const rows = within(table).getAllByRole("row").slice(1);
        expect(rows[1]).toHaveAttribute("data-qualifying-line", "true");
        expect(rows[0]).not.toHaveAttribute("data-qualifying-line");
    });

    it("switches groups on a phone", async () => {
        serve(snapshot());
        render(<App/>);

        const groupB = await screen.findByRole("button", {name: "Group B"});
        expect(groupB).toHaveAttribute("aria-pressed", "false");
        await userEvent.click(groupB);
        expect(groupB).toHaveAttribute("aria-pressed", "true");
    });

    it("says when the server can't be reached", async () => {
        vi.stubGlobal("fetch", vi.fn(async () => new Response("", {status: 500})));
        render(<App/>);

        expect(await screen.findByRole("status")).toHaveTextContent("Can't reach the server");
    });
});
