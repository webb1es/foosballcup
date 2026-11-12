import {render, screen, within} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import {beforeEach, describe, expect, it} from "vitest";
import {App} from "../App";
import {match, serve, snapshot} from "../test/fixture";

/** Fixture cards on the page, excluding the header's day pills. */
function cards() {
    return within(screen.getByRole("main")).queryAllByRole("listitem");
}

function withDayTwo() {
    const snap = snapshot();
    snap.matches.push(match({
        id: 5,
        match_no: 5,
        match_day: 2,
        scheduled_date: "2026-11-13",
        player_a_id: 3,
        player_b_id: 2
    }));
    return snap;
}

describe("Fixtures page", () => {
    beforeEach(() => window.history.pushState({}, "", "/fixtures"));

    it("lists matches under their match day with played counts", async () => {
        serve(withDayTwo());
        render(<App/>);

        const day1 = await screen.findByRole("region", {name: "Day 1"});
        expect(within(day1).getAllByRole("listitem")).toHaveLength(4);
        expect(within(day1).getByText("2/4 PLAYED")).toBeInTheDocument();
        expect(within(day1).getByRole("listitem", {name: "Match 1: Malume v Tomishto, 10 to 3, final"})).toBeInTheDocument();
        expect(within(screen.getByRole("region", {name: "Day 2"})).getAllByRole("listitem")).toHaveLength(1);
    });

    it("filters by nickname or full name, ignoring case", async () => {
        serve(withDayTwo());
        render(<App/>);
        const input = await screen.findByRole("combobox", {name: "Filter by player or nickname"});

        await userEvent.type(input, "bento");
        expect(cards().map((li) => li.getAttribute("aria-label"))).toEqual([
            "Match 2: Tomishto v Bento",
            "Match 3: Malume v Bento",
            "Match 5: Bento v Tomishto",
        ]);

        await userEvent.clear(input);
        await userEvent.type(input, "GWESE");
        expect(cards()).toHaveLength(1);
        expect(window.location.search).toBe("?q=GWESE");
    });

    it("shows one day when picked", async () => {
        serve(withDayTwo());
        render(<App/>);

        await userEvent.click(await screen.findByRole("button", {name: "Day 2"}));
        expect(screen.queryByRole("region", {name: "Day 1"})).not.toBeInTheDocument();
        expect(screen.getByRole("region", {name: "Day 2"})).toBeInTheDocument();
        expect(screen.getByRole("button", {name: "Day 2"})).toHaveAttribute("aria-pressed", "true");
    });

    it("restores the filter from the URL", async () => {
        window.history.pushState({}, "", "/fixtures?q=tapf&day=1");
        serve(withDayTwo());
        render(<App/>);

        expect(await screen.findByRole("combobox")).toHaveValue("tapf");
        expect(cards()).toHaveLength(1);
    });

    it("says when nothing matches", async () => {
        serve(snapshot());
        render(<App/>);

        await userEvent.type(await screen.findByRole("combobox"), "nobody");
        expect(screen.getByRole("status")).toHaveTextContent('No matches for "nobody".');
    });
});
