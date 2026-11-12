import {describe, expect, it} from "vitest";
import {match} from "../test/fixture";
import {recentForm} from "./standings";

describe("recentForm", () => {
    it("keeps the latest confirmed results per player, oldest first", () => {
        const ms = [
            match({
                id: 1,
                match_no: 1,
                status: "confirmed",
                score_a: 10,
                score_b: 3,
                confirmed_at: "2026-11-11T10:00:00Z"
            }),
            match({
                id: 2,
                match_no: 2,
                status: "confirmed",
                score_a: 4,
                score_b: 10,
                confirmed_at: "2026-11-11T11:00:00Z"
            }),
            match({
                id: 3,
                match_no: 3,
                status: "confirmed",
                score_a: 10,
                score_b: 9,
                confirmed_at: "2026-11-11T12:00:00Z"
            }),
            match({id: 4, match_no: 4, status: "provisional", score_a: 10, score_b: 0}),
        ];
        const form = recentForm(ms, 2);
        expect(form.get(1)).toEqual(["L", "W"]);
        expect(form.get(2)).toEqual(["W", "L"]);
    });
});
