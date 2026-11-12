import {describe, expect, it} from "vitest";
import {match, player} from "../test/fixture";
import {
    activeMatches,
    dayStates,
    displayName,
    initials,
    latestResults,
    localDate,
    upcomingMatches,
    upNextTag
} from "./home";

describe("activeMatches", () => {
    it("orders live before paused before provisional and drops the rest", () => {
        const ms = [
            match({id: 1, match_no: 1, status: "provisional"}),
            match({id: 2, match_no: 2, status: "confirmed"}),
            match({id: 3, match_no: 3, status: "paused"}),
            match({id: 4, match_no: 4, status: "live"}),
            match({id: 5, match_no: 5, status: "abandoned"}),
        ];
        expect(activeMatches(ms).map((m) => m.id)).toEqual([4, 3, 1]);
    });
});

describe("upcomingMatches", () => {
    it("puts leftover matches from earlier days first", () => {
        const ms = [
            match({id: 1, match_no: 30, match_day: 2}),
            match({id: 2, match_no: 5, match_day: 1}),
            match({id: 3, match_no: 1, match_day: 1, status: "confirmed"}),
            match({id: 4, match_no: 3, match_day: 1}),
        ];
        expect(upcomingMatches(ms).map((m) => m.match_no)).toEqual([3, 5, 30]);
    });
});

describe("latestResults", () => {
    it("returns the most recently confirmed results first", () => {
        const ms = [
            match({id: 1, match_no: 1, status: "confirmed", confirmed_at: "2026-11-11T09:00:00Z"}),
            match({id: 2, match_no: 2, status: "confirmed", confirmed_at: "2026-11-11T11:00:00Z"}),
            match({id: 3, match_no: 3, status: "provisional"}),
            match({id: 4, match_no: 4, status: "confirmed", confirmed_at: "2026-11-11T10:00:00Z"}),
        ];
        expect(latestResults(ms, 2).map((m) => m.id)).toEqual([2, 4]);
    });
});

describe("dayStates", () => {
    it("marks days before, on and after today", () => {
        const days = dayStates(["2026-11-11", "2026-11-13", "2026-11-18"], "2026-11-13");
        expect(days.map((d) => `${d.label}:${d.state}`)).toEqual(["DAY 1:past", "DAY 2:today", "DAY 3:future"]);
    });
});

describe("localDate", () => {
    it("formats the local calendar date", () => {
        expect(localDate(new Date(2026, 10, 5, 23, 30))).toBe("2026-11-05");
    });
});

describe("upNextTag", () => {
    it("labels the first match NEXT, same-day matches TODAY and later ones by date", () => {
        expect(upNextTag(match({id: 1, match_no: 1}), 0, "2026-11-13")).toBe("NEXT");
        expect(upNextTag(match({id: 2, match_no: 2, scheduled_date: "2026-11-13"}), 1, "2026-11-13")).toBe("TODAY");
        expect(upNextTag(match({id: 3, match_no: 3, scheduled_date: "2026-11-18"}), 2, "2026-11-13")).toBe("18 NOV");
    });
});

describe("player names", () => {
    it("prefers the nickname and falls back to the first name", () => {
        expect(displayName(player(1, "Job Gwese", "Honey Badger"))).toBe("Honey Badger");
        expect(displayName(player(2, "Job Gwese", null))).toBe("Job");
        expect(displayName(undefined)).toBe("TBC");
    });

    it("builds initials from the first and last name", () => {
        expect(initials(player(1, "Takudzwa Chikwanda", null))).toBe("TC");
        expect(initials(player(2, "Madonna", null))).toBe("M");
    });
});
