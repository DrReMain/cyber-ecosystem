import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Filter } from "./filter";

const labels = { search: "查询", reset: "重置", fold: "收起", expand: "展开" };

const options = [
  { label: "关键词", name: "keyword", placeholder: "名称或编码" },
  {
    label: "状态",
    name: "status",
    type: "select" as const,
    options: [
      { label: "启用", value: "on" },
      { label: "停用", value: "off" },
    ],
  },
];

describe("Filter", () => {
  it("renders labels, an input with placeholder and a select", () => {
    const { container } = render(<Filter labels={labels} options={options} />);
    expect(screen.getByText("关键词")).toBeTruthy();
    expect(screen.getByText("状态")).toBeTruthy();
    expect(screen.getByPlaceholderText("名称或编码")).toBeTruthy();
    expect(container.querySelectorAll(".ant-select").length).toBe(1);
  });

  it("submits entered values through onFilter", async () => {
    const onFilter = vi.fn();
    render(<Filter labels={labels} onFilter={onFilter} options={options} />);
    fireEvent.change(screen.getByPlaceholderText("名称或编码"), { target: { value: "审计" } });
    fireEvent.click(screen.getByRole("button", { name: "查询" }));
    // rc-field-form validates asynchronously before onFinish fires.
    await waitFor(() => expect(onFilter).toHaveBeenCalledOnce());
    expect(onFilter).toHaveBeenCalledWith({ keyword: "审计", status: undefined });
  });

  it("reset routes through onReset when provided", () => {
    const onReset = vi.fn();
    render(<Filter labels={labels} onFilter={vi.fn()} onReset={onReset} options={options} />);
    fireEvent.click(screen.getByRole("button", { name: "重置" }));
    expect(onReset).toHaveBeenCalledOnce();
  });

  it("prefills from initialValues", () => {
    render(
      <Filter
        initialValues={{ keyword: "预填" } as Record<string, unknown>}
        labels={labels}
        options={options}
      />,
    );
    expect(screen.getByPlaceholderText("名称或编码")).toHaveValue("预填");
  });

  it("renders custom element options (escape hatch)", () => {
    render(
      <Filter
        labels={labels}
        options={[
          ...options,
          { label: "自定义", name: "custom", element: <span data-testid="custom-field">X</span> },
        ]}
      />,
    );
    expect(screen.getByTestId("custom-field")).toBeTruthy();
  });

  it("renders every builtin field kind", () => {
    const { container } = render(
      <Filter
        labels={labels}
        options={[
          { label: "文本", name: "kw", placeholder: "t" },
          { label: "选择", name: "st", type: "select", options: [{ label: "A", value: "a" }] },
          { label: "日期", name: "d", type: "date" },
          { label: "数字", name: "n", type: "number" },
          { label: "数值区间", name: ["nA", "nZ"], type: "range-number" },
          { label: "日期区间", name: ["dA", "dZ"], type: "range-date" },
        ]}
      />,
    );
    for (const label of ["文本", "选择", "日期", "数字", "数值区间", "日期区间"]) {
      expect(screen.getByText(label)).toBeTruthy();
    }
    expect(container.querySelectorAll(".ant-select").length).toBe(1);
    expect(container.querySelectorAll(".ant-picker").length).toBe(2); // date + range picker
    expect(container.querySelectorAll(".ant-picker-range").length).toBe(1);
    expect(container.querySelectorAll(".ant-input-number").length).toBe(3); // number + range-number pair
  });

  it("flattens range-date tuples into their key pair on submit", async () => {
    const onFilter = vi.fn();
    render(
      <Filter
        initialValues={{ createdAtA: 1000, createdAtZ: 2000 } as Record<string, unknown>}
        labels={labels}
        onFilter={onFilter}
        options={[
          { label: "时间", name: ["createdAtA", "createdAtZ"], type: "range-datetime" as const },
        ]}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "查询" }));
    await waitFor(() => expect(onFilter).toHaveBeenCalledOnce());
    expect(onFilter).toHaveBeenCalledWith({ createdAtA: 1000, createdAtZ: 2000 });
  });

  it("keeps a one-sided range one-sided through the tuple round-trip", async () => {
    const onFilter = vi.fn();
    render(
      <Filter
        initialValues={{ createdAtA: 1000 } as Record<string, unknown>}
        labels={labels}
        onFilter={onFilter}
        options={[
          { label: "时间", name: ["createdAtA", "createdAtZ"], type: "range-datetime" as const },
        ]}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "查询" }));
    await waitFor(() => expect(onFilter).toHaveBeenCalledOnce());
    expect(onFilter).toHaveBeenCalledWith({ createdAtA: 1000, createdAtZ: undefined });
  });
});
