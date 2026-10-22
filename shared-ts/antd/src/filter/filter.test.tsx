import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { Filter } from "./filter";

const labels = { search: "Search", reset: "Reset", fold: "Collapse", expand: "Expand" };

const options = [
  { label: "Keyword", name: "keyword", placeholder: "Name or code" },
  {
    label: "Status",
    name: "status",
    type: "select" as const,
    options: [
      { label: "Enabled", value: "on" },
      { label: "Disabled", value: "off" },
    ],
  },
];

describe("Filter", () => {
  it("renders labels, an input with placeholder and a select", () => {
    const { container } = render(<Filter labels={labels} options={options} />);
    expect(screen.getByText("Keyword")).toBeTruthy();
    expect(screen.getByText("Status")).toBeTruthy();
    expect(screen.getByPlaceholderText("Name or code")).toBeTruthy();
    expect(container.querySelectorAll(".ant-select").length).toBe(1);
  });

  it("submits entered values through onFilter", async () => {
    const onFilter = vi.fn();
    render(<Filter labels={labels} onFilter={onFilter} options={options} />);
    fireEvent.change(screen.getByPlaceholderText("Name or code"), { target: { value: "audit" } });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    // rc-field-form validates asynchronously before onFinish fires.
    await waitFor(() => expect(onFilter).toHaveBeenCalledOnce());
    expect(onFilter).toHaveBeenCalledWith({ keyword: "audit", status: undefined });
  });

  it("reset routes through onReset when provided", () => {
    const onReset = vi.fn();
    render(<Filter labels={labels} onFilter={vi.fn()} onReset={onReset} options={options} />);
    fireEvent.click(screen.getByRole("button", { name: "Reset" }));
    expect(onReset).toHaveBeenCalledOnce();
  });

  it("prefills from initialValues", () => {
    render(
      <Filter
        initialValues={{ keyword: "prefilled" } as Record<string, unknown>}
        labels={labels}
        options={options}
      />,
    );
    expect(screen.getByPlaceholderText("Name or code")).toHaveValue("prefilled");
  });

  it("renders custom element options (escape hatch)", () => {
    render(
      <Filter
        labels={labels}
        options={[
          ...options,
          { label: "Custom", name: "custom", element: <span data-testid="custom-field">X</span> },
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
          { label: "Text", name: "kw", placeholder: "t" },
          { label: "Select", name: "st", type: "select", options: [{ label: "A", value: "a" }] },
          { label: "Date", name: "d", type: "date" },
          { label: "Number", name: "n", type: "number" },
          { label: "Number range", name: ["nA", "nZ"], type: "range-number" },
          { label: "Date range", name: ["dA", "dZ"], type: "range-date" },
        ]}
      />,
    );
    for (const label of ["Text", "Select", "Date", "Number", "Number range", "Date range"]) {
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
          { label: "Time", name: ["createdAtA", "createdAtZ"], type: "range-datetime" as const },
        ]}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
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
          { label: "Time", name: ["createdAtA", "createdAtZ"], type: "range-datetime" as const },
        ]}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    await waitFor(() => expect(onFilter).toHaveBeenCalledOnce());
    expect(onFilter).toHaveBeenCalledWith({ createdAtA: 1000, createdAtZ: undefined });
  });
});
