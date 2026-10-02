import { Collapse, Descriptions } from "antd";
import { m } from "#/paraglide/messages";

interface SessionPanelProps {
  email: string;
  permissions: string[];
}

export function SessionPanel({ email, permissions }: Readonly<SessionPanelProps>) {
  return (
    <Collapse
      items={[
        {
          key: "session",
          label: m.system_diag_session(),
          children: (
            <Descriptions
              colon={false}
              column={1}
              items={[
                { key: "email", label: "GetCurrentUser", children: email },
                {
                  key: "patterns",
                  label: "patterns",
                  children: (
                    <span className="flex flex-wrap gap-1">
                      {permissions.map((p) => (
                        <span
                          className="rounded bg-fill-tertiary px-1.5 py-0.5 font-mono text-[12px]"
                          key={p}
                        >
                          {p}
                        </span>
                      ))}
                    </span>
                  ),
                },
              ]}
              size="small"
            />
          ),
        },
      ]}
    />
  );
}
