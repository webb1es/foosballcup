import {render, screen, within} from "@testing-library/react";
import {beforeEach, describe, expect, it} from "vitest";
import {App} from "../App";
import {serve, snapshot} from "../test/fixture";

describe("Standings page", () => {
    beforeEach(() => window.history.pushState({}, "", "/standings"));

    it("shows every group with goals for, against and recent form", async () => {
        serve(snapshot());
        render(<App/>);

        const groupA = await screen.findByRole("region", {name: "Group A"});
        expect(screen.getByRole("region", {name: "Group B"})).toBeInTheDocument();
        const [header, first] = within(groupA).getAllByRole("row");
        expect(within(header).getByTitle("Goals for")).toBeInTheDocument();
        expect(within(first).getByText("Trust Mandinyenya")).toBeInTheDocument();
        expect(within(first).getByLabelText("Last 1: W")).toBeInTheDocument();
    });

    it("marks the qualifying line under the last qualifying place", async () => {
        serve(snapshot());
        render(<App/>);

        const rows = within(await screen.findByRole("region", {name: "Group A"})).getAllByRole("row").slice(1);
        expect(rows[1]).toHaveAttribute("data-qualifying-line", "true");
    });
});
