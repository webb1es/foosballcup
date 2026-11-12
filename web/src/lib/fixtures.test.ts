import {describe, expect, it} from "vitest";
import {match, player} from "../test/fixture";
import {filterByPlayer, fixtureSections, formatDay, KNOCKOUTS, parseSectionKey} from "./fixtures";

describe("fixtureSections", () => {
    it("groups by match day in match order and puts knockouts last", () => {
        const ms = [
            match({id: 1, match_no: 30, stage: "final", group: null, match_day: null}),
            match({id: 2, match_no: 12, match_day: 2}),
            match({id: 3, match_no: 2, match_day: 1}),
            match({id: 4, match_no: 1, match_day: 1}),
        ];
        const sections = fixtureSections(ms, ["2026-11-11", "2026-11-13"]);
        expect(sections.map((s) => [s.key, s.date, s.matches.map((m) => m.id)])).toEqual([
            [1, "2026-11-11", [4, 3]],
            [2, "2026-11-13", [2]],
            [KNOCKOUTS, null, [1]],
        ]);
    });
});

describe("filterByPlayer", () => {
    const players = new Map([
        [1, player(1, "Trust Mandinyenya", "Malume")],
        [2, player(2, "Thomas Nyagwaya", null)],
        [3, player(3, "Benard Mwanjeya", "Bento")],
    ]);
    const ms = [
        match({id: 1, match_no: 1, player_a_id: 1, player_b_id: 2}),
        match({id: 2, match_no: 2, player_a_id: 2, player_b_id: 3}),
        match({id: 3, match_no: 3, player_a_id: null, player_b_id: null}),
    ];
    const ids = (q: string) => filterByPlayer(ms, players, q).map((m) => m.id);

    it.each([
        ["", [1, 2, 3]],
        ["  ", [1, 2, 3]],
        ["malume", [1]],
        ["NYAGWAYA", [1, 2]],
        [" bento ", [2]],
        ["zzz", []],
    ])("query %j matches %j", (q, want) => {
        expect(ids(q)).toEqual(want);
    });
});

describe("parseSectionKey", () => {
    it.each([
        [null, null],
        ["2", 2],
        ["ko", KNOCKOUTS],
        ["0", null],
        ["x", null],
        ["1.5", null],
    ])("%j is %j", (raw, want) => {
        expect(parseSectionKey(raw)).toBe(want);
    });
});

describe("formatDay", () => {
    it("formats a date without shifting by time zone", () => {
        expect(formatDay("2026-11-11")).toBe("Wed 11 Nov");
    });
});
