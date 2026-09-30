import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { TableToolbar } from "./table-toolbar";

const labels = {
  refresh: "刷新",
  density: "密度",
  densityLarge: "宽松",
  densityMiddle: "默认",
  densitySmall: "紧凑",
};

describe("TableToolbar", () => {
  it("renders the extra slot on the left", () => {
    render(<TableToolbar extra={<button type="button">新建</button>} />);
    expect(screen.getByRole("button", { name: "新建" })).toBeTruthy();
  });

  it("hides the refresh button unless onRefresh is provided", () => {
    const { rerender } = render(
      <TableToolbar labels={labels} size="middle" onSizeChange={() => {}} />,
    );
    expect(screen.queryByRole("button", { name: "刷新" })).toBeNull();
    rerender(
      <TableToolbar labels={labels} onRefresh={() => {}} size="middle" onSizeChange={() => {}} />,
    );
    expect(screen.getByRole("button", { name: "刷新" })).toBeTruthy();
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
    expect(screen.getByRole("button", { name: "刷新" })).toBeDisabled();
    rerender(
      <TableToolbar labels={labels} onRefresh={onRefresh} size="middle" onSizeChange={() => {}} />,
    );
    fireEvent.click(screen.getByRole("button", { name: "刷新" }));
    expect(onRefresh).toHaveBeenCalledOnce();
  });

  it("offers the density menu and reports the picked size", () => {
    const onSizeChange = vi.fn();
    render(<TableToolbar labels={labels} size="middle" onSizeChange={onSizeChange} />);
    fireEvent.click(screen.getByRole("button", { name: "密度" }));
    fireEvent.click(screen.getByText("紧凑"));
    expect(onSizeChange).toHaveBeenCalledWith("small");
  });

  it("renders the tools slot before density", () => {
    render(
      <TableToolbar
        labels={labels}
        size="middle"
        tools={<button type="button">导出</button>}
        onSizeChange={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: "导出" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "密度" })).toBeTruthy();
  });

  it("renders no icon buttons when neither density nor refresh is wired", () => {
    render(<TableToolbar />);
    expect(screen.queryByRole("button")).toBeNull();
  });
});
