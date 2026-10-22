import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { DataTable } from "./data-table";

interface Row {
  id: string;
  name: string;
  code: string;
}

const rows: Row[] = [
  { id: "1", name: "Alpha", code: "a" },
  { id: "2", name: "Beta", code: "b" },
];

const baseColumns = [
  { title: "Name", dataIndex: "name", key: "name" },
  { title: "Code", dataIndex: "code", key: "code" },
];

// rc-table renders a hidden measuring row that duplicates header text;
// header assertions anchor on real thead cells instead of text queries.
function theadTexts(container: HTMLElement): string[] {
  return Array.from(container.querySelectorAll(".ant-table-thead th")).map(
    (th) => th.textContent ?? "",
  );
}

describe("DataTable", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("renders rows through the table", () => {
    const { container } = render(
      <DataTable<Row> columns={baseColumns} dataSource={rows} rowKey="id" />,
    );
    expect(screen.getByText("Alpha")).toBeTruthy();
    expect(screen.getByText("Beta")).toBeTruthy();
    expect(theadTexts(container)).toContain("Name");
  });

  it("renders no pagination by default", () => {
    const { container } = render(
      <DataTable<Row> columns={baseColumns} dataSource={rows} rowKey="id" />,
    );
    expect(container.querySelector(".ant-pagination")).toBeNull();
  });

  it("passes loading through to the table", () => {
    const { container } = render(
      <DataTable<Row> columns={baseColumns} dataSource={rows} loading rowKey="id" />,
    );
    expect(container.querySelectorAll(".ant-spin").length).toBeGreaterThan(0);
  });

  it("renders the rich empty state with its call-to-action", () => {
    render(
      <DataTable<Row>
        columns={baseColumns}
        dataSource={[]}
        emptyAction={<button type="button">Create</button>}
        emptyDescription="No records"
        rowKey="id"
      />,
    );
    expect(screen.getByText("No records")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Create" })).toBeTruthy();
  });

  it("hides the checkbox column until select mode is toggled on", () => {
    const { container } = render(
      <DataTable<Row>
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
        rowSelection={{ selectedRowKeys: [], onChange: vi.fn() }}
        selection={{ labels: { selected: (n) => `${n} selected`, clear: "Clear", mode: "Select" } }}
      />,
    );
    expect(container.querySelectorAll(".ant-table-selection-column").length).toBe(0);
    fireEvent.click(screen.getByRole("button", { name: "Select" }));
    expect(container.querySelectorAll(".ant-table-selection-column").length).toBeGreaterThan(0);
    expect(screen.getByText("0 selected")).toBeTruthy();
  });

  it("shows the alert bar in select mode and clears through onChange", () => {
    const onChange = vi.fn();
    render(
      <DataTable<Row>
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
        rowSelection={{ selectedRowKeys: ["1"], onChange }}
        selection={{ labels: { selected: (n) => `${n} selected`, clear: "Clear", mode: "Select" } }}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Select" }));
    expect(screen.getByText("1 selected")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Clear" }));
    expect(onChange).toHaveBeenCalledWith([], [], { type: "none" });
  });

  it("exits select mode through the alert close and resets keys", () => {
    const onChange = vi.fn();
    const { container } = render(
      <DataTable<Row>
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
        rowSelection={{ selectedRowKeys: ["1"], onChange }}
        selection={{ labels: { selected: (n) => `${n} selected`, clear: "Clear", mode: "Select" } }}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Select" }));
    const closeBtn = container.querySelector("button.ant-alert-close-icon");
    expect(closeBtn).toBeTruthy();
    fireEvent.click(closeBtn as HTMLButtonElement);
    expect(onChange).toHaveBeenCalledWith([], [], { type: "none" });
    expect(container.querySelectorAll(".ant-table-selection-column").length).toBe(0);
    expect(screen.queryByText(/selected/)).toBeNull();
  });
});

describe("DataTable column settings", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  const config = {
    storageKey: "test.dt.columns",
    locked: ["name"],
    labels: { title: "Columns", reset: "Reset" },
  };

  async function openPanel(anchor = "Code") {
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    await waitFor(() => {
      const label = screen
        .getAllByText(anchor)
        .map((n) => n.closest("label"))
        .find((l): l is HTMLLabelElement => l !== null);
      expect(label).toBeTruthy();
    });
  }

  function checkboxOf(label: string): HTMLInputElement {
    const wrapper = screen
      .getAllByText(label)
      .map((n) => n.closest("label"))
      .find((l): l is HTMLLabelElement => l !== null);
    const input = wrapper?.querySelector("input");
    if (!input) throw new Error(`checkbox for ${label} not found`);
    return input;
  }

  it("hides a column on uncheck, persists across remounts, and reset restores", async () => {
    const first = render(
      <DataTable<Row>
        columnSettings={config}
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
      />,
    );
    await openPanel();
    expect(checkboxOf("Name")).toBeDisabled(); // locked
    fireEvent.click(checkboxOf("Code"));
    await waitFor(() =>
      expect(theadTexts(first.container).some((t) => t.includes("Code"))).toBe(false),
    );
    expect(window.localStorage.getItem(config.storageKey)).toBe(JSON.stringify(["code"]));

    first.unmount();
    const second = render(
      <DataTable<Row>
        columnSettings={config}
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
      />,
    );
    await waitFor(() =>
      expect(theadTexts(second.container).some((t) => t.includes("Code"))).toBe(false),
    );

    await openPanel();
    fireEvent.click(await screen.findByRole("button", { name: "Reset" }));
    await waitFor(() =>
      expect(theadTexts(second.container).some((t) => t.includes("Code"))).toBe(true),
    );
    expect(window.localStorage.getItem(config.storageKey)).toBeNull();
  });

  it("tracks nested-path (array dataIndex) columns under a dotted id", async () => {
    interface NestedRow {
      id: string;
      name: string;
      meta: { updated: string };
    }
    const nestedRows: NestedRow[] = [
      { id: "1", name: "Alpha", meta: { updated: "x" } },
      { id: "2", name: "Beta", meta: { updated: "y" } },
    ];
    const nestedColumns = [
      { title: "Name", dataIndex: "name", key: "name" },
      { title: "Updated", dataIndex: ["meta", "updated"] },
    ];
    const { container } = render(
      <DataTable<NestedRow>
        columnSettings={config}
        columns={nestedColumns}
        dataSource={nestedRows}
        rowKey="id"
      />,
    );
    await openPanel("Updated");
    fireEvent.click(checkboxOf("Updated"));
    await waitFor(() =>
      expect(theadTexts(container).some((t) => t.includes("Updated"))).toBe(false),
    );
    expect(window.localStorage.getItem(config.storageKey)).toBe(JSON.stringify(["meta.updated"]));
  });

  it("renders no settings trigger without the config", () => {
    render(<DataTable<Row> columns={baseColumns} dataSource={rows} rowKey="id" />);
    expect(screen.queryByRole("button", { name: "Columns" })).toBeNull();
  });
});
