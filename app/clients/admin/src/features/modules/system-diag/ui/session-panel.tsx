import { Collapse, Descriptions } from "antd";
import { m } from "#/paraglide/messages";

interface MembershipView {
  tenant: string;
  population: string;
}

interface SessionPanelProps {
  email: string;
  permissions: string[];
  memberships: readonly MembershipView[];
}

export function SessionPanel({ email, permissions, memberships }: Readonly<SessionPanelProps>) {
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
                  key: "memberships",
                  label: m.system_diag_session_seats(),
                  children: (
                    <div className="flex flex-col gap-1">
                      {memberships.length === 0 ? (
                        <span className="text-[12px]">—</span>
                      ) : (
                        memberships.map((membership) => (
                          <span
                            className="font-mono text-[12px]"
                            key={`${membership.tenant}/${membership.population}`}
                          >
                            {membership.tenant} / {membership.population}
                          </span>
                        ))
                      )}
                    </div>
                  ),
                },
                {
                  key: "patterns",
                  label: "patterns",
                  children: (
                    <div className="flex flex-col gap-1">
                      {permissions.map((p) => (
                        <span className="font-mono text-[12px]" key={p}>
                          {p}
                        </span>
                      ))}
                    </div>
                  ),
                },
              ]}
              layout="vertical"
              size="small"
            />
          ),
        },
      ]}
    />
  );
}
