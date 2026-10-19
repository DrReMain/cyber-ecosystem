import type { Meta, StoryObj } from "@storybook/react-vite";
import { Button, Space } from "antd";
import { Download, Plus } from "lucide-react";
import { Label, Section } from "../helpers";

const meta: Meta<typeof Button> = {
  title: "Components/Buttons",
  component: Button,
};

export default meta;

type Story = StoryObj<typeof Button>;

const VARIANTS = ["solid", "outlined", "dashed", "filled", "text", "link"] as const;
// ButtonColorType has no success/warning — semantic colors beyond danger are
// preset-palette names only (v6 contract).
const COLORS = ["primary", "default", "danger", "blue", "purple"] as const;

function VariantMatrix() {
  return (
    <Space vertical size={16} style={{ display: "flex" }}>
      {VARIANTS.map((variant) => (
        <div key={variant}>
          <Label>{variant}</Label>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 8, marginTop: 4 }}>
            {COLORS.map((color) => (
              <Button color={color} key={color} variant={variant}>
                {color}
              </Button>
            ))}
          </div>
        </div>
      ))}
    </Space>
  );
}

function ButtonsPage() {
  return (
    <div style={{ maxWidth: 900 }}>
      <Section
        description="Every variant x color pair the skin must style coherently."
        title="Variant x color matrix"
      >
        <VariantMatrix />
      </Section>

      <Section title="Sizes">
        <Space wrap>
          <Button size="small">Small</Button>
          <Button>Middle</Button>
          <Button size="large">Large</Button>
          <Button size="small" type="primary">
            Small primary
          </Button>
          <Button type="primary">Middle primary</Button>
          <Button size="large" type="primary">
            Large primary
          </Button>
        </Space>
      </Section>

      <Section title="States">
        <Space wrap>
          <Button disabled>Disabled</Button>
          <Button disabled type="primary">
            Disabled primary
          </Button>
          <Button loading>Loading</Button>
          <Button loading type="primary">
            Loading primary
          </Button>
          <Button icon={<Plus aria-hidden size={14} />}>Icon start</Button>
          <Button icon={<Download aria-hidden size={14} />} iconPosition="end">
            Icon end
          </Button>
          <Button danger>Legacy danger</Button>
        </Space>
      </Section>

      <Section title="Block & shape">
        <Space vertical size={12} style={{ display: "flex", maxWidth: 360 }}>
          <Button block type="primary">
            Block primary
          </Button>
          <Button block>Block default</Button>
          <Space wrap>
            <Button shape="circle" icon={<Plus aria-hidden size={14} />} />
            <Button shape="round">Round</Button>
          </Space>
        </Space>
      </Section>
    </div>
  );
}

export const All: Story = {
  render: () => <ButtonsPage />,
};
