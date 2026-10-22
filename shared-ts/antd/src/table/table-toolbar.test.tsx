import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TableToolbar } from "./table-toolbar";

const labels = {
  refresh: "Refresh",
  density: "Density",
  densityLarge: "Large",
  densityMiddle: "Middle",
  densitySmall: "Compact",
};

describe("TableToolbar", () => {
  it("renders the extra slot on the left", () => {
    render(<TableToolbar extra={<button type="button">Create</button>} />);
    expect(screen.getByRole("button", { name: "Create" })).toBeTruthy();
  });

  it("hides the refresh button unless onRefresh is provided", () => {
    const { rerender } = render(
      <TableToolbar labels={labels} size="middle" onSizeChange={() => {}} />,
    );
    expect(screen.queryByRole("button", { name: "Refresh" })).toBeNull();
    rerender(
      <TableToolbar labels={labels} onRefresh={() => {}} size="middle" onSizeChange={() => {}} />,
    );
    expect(screen.getByRole("button", { name: "Refresh" })).toBeTruthy();
  });

  it("disables refresh while loading and still fires it when enabled", () => {
    const onRefresh = vi.fn();
    const { rerender } = render(
      <TableToolbar
        labels={labels}
        loading
        onRefresh={onRefresh}
        size="middle"
        onSizeChange={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: "Refresh" })).toBeDisabled();
    rerender(
      <TableToolbar labels={labels} onRefresh={onRefresh} size="middle" onSizeChange={() => {}} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    expect(onRefresh).toHaveBeenCalledOnce();
  });

  it("offers the density menu and reports the picked size", () => {
    const onSizeChange = vi.fn();
    render(<TableToolbar labels={labels} size="middle" onSizeChange={onSizeChange} />);
    fireEvent.click(screen.getByRole("button", { name: "Density" }));
    fireEvent.click(screen.getByText("Compact"));
    expect(onSizeChange).toHaveBeenCalledWith("small");
  });

  it("renders the tools slot before density", () => {
    render(
      <TableToolbar
        labels={labels}
        size="middle"
        tools={<button type="button">Export</button>}
        onSizeChange={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: "Export" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "Density" })).toBeTruthy();
  });

  it("renders no icon buttons when neither density nor refresh is wired", () => {
    render(<TableToolbar />);
    expect(screen.queryByRole("button")).toBeNull();
  });
});
