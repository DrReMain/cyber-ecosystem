import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { DataTable } from "./data-table";

interface Row {
  id: string;
  name: string;
  code: string;
}

const rows: Row[] = [
  { id: "1", name: "甲", code: "a" },
  { id: "2", name: "乙", code: "b" },
];

const baseColumns = [
  { title: "名称", dataIndex: "name", key: "name" },
  { title: "编码", dataIndex: "code", key: "code" },
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
    expect(screen.getByText("甲")).toBeTruthy();
    expect(screen.getByText("乙")).toBeTruthy();
    expect(theadTexts(container)).toContain("名称");
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
        emptyAction={<button type="button">新建</button>}
        emptyDescription="暂无数据"
        rowKey="id"
      />,
    );
    expect(screen.getByText("暂无数据")).toBeTruthy();
    expect(screen.getByRole("button", { name: "新建" })).toBeTruthy();
  });

  it("shows the selection bar while rows are selected and clears through onChange", () => {
    const onChange = vi.fn();
    const { rerender } = render(
      <DataTable<Row>
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
        rowSelection={{ selectedRowKeys: ["1"], onChange }}
        selection={{ labels: { selected: (n) => `已选择 ${n} 项`, clear: "清空" } }}
      />,
    );
    expect(screen.getByText("已选择 1 项")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "清空" }));
    expect(onChange).toHaveBeenCalledWith([], [], { type: "none" });

    rerender(
      <DataTable<Row>
        columns={baseColumns}
        dataSource={rows}
        rowKey="id"
        rowSelection={{ selectedRowKeys: [], onChange }}
        selection={{ labels: { selected: (n) => `已选择 ${n} 项`, clear: "清空" } }}
      />,
    );
    expect(screen.queryByText(/已选择/)).toBeNull();
  });
});

describe("DataTable column settings", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  const config = {
    storageKey: "test.dt.columns",
    locked: ["name"],
    labels: { title: "列设置", reset: "恢复" },
  };

  async function openPanel(anchor = "编码") {
    fireEvent.click(screen.getByRole("button", { name: "列设置" }));
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
    expect(checkboxOf("名称")).toBeDisabled(); // locked
    fireEvent.click(checkboxOf("编码"));
    await waitFor(() =>
      expect(theadTexts(first.container).some((t) => t.includes("编码"))).toBe(false),
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
      expect(theadTexts(second.container).some((t) => t.includes("编码"))).toBe(false),
    );

    await openPanel();
    fireEvent.click(await screen.findByRole("button", { name: "恢复" }));
    await waitFor(() =>
      expect(theadTexts(second.container).some((t) => t.includes("编码"))).toBe(true),
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
      { id: "1", name: "甲", meta: { updated: "x" } },
      { id: "2", name: "乙", meta: { updated: "y" } },
    ];
    const nestedColumns = [
      { title: "名称", dataIndex: "name", key: "name" },
      { title: "更新", dataIndex: ["meta", "updated"] },
    ];
    const { container } = render(
      <DataTable<NestedRow>
        columnSettings={config}
        columns={nestedColumns}
        dataSource={nestedRows}
        rowKey="id"
      />,
    );
    await openPanel("更新");
    fireEvent.click(checkboxOf("更新"));
    await waitFor(() => expect(theadTexts(container).some((t) => t.includes("更新"))).toBe(false));
    expect(window.localStorage.getItem(config.storageKey)).toBe(JSON.stringify(["meta.updated"]));
  });

  it("renders no settings trigger without the config", () => {
    render(<DataTable<Row> columns={baseColumns} dataSource={rows} rowKey="id" />);
    expect(screen.queryByRole("button", { name: "列设置" })).toBeNull();
  });
});
