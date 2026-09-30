import type { Service } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource_pb";
import { Select } from "antd";
import { useMemo } from "react";
import { m } from "#/paraglide/messages";
import { type GrantableView, grantableServices, opOf } from "../../system-roles/grants";

function operationOptions(services: GrantableView[]) {
  return services.map((svc) => ({
    label: `${svc.comment || svc.fullName} · ${svc.fullName}`,
    options: svc.methods
      .filter((mt) => mt.name)
      .map((mt) => ({
        value: opOf(svc.fullName, mt.name),
        label: (
          <span className="flex items-center gap-2">
            <span>{mt.comment || mt.name}</span>
            <span className="font-mono text-[12px] text-black/40 dark:text-white/40">
              {mt.name}
            </span>
          </span>
        ),
        search: `${mt.name} ${mt.comment ?? ""} ${svc.fullName}`,
      })),
  }));
}

interface OperationSelectProps {
  catalog: Service[];
  value?: string;
  onChange: (next: string) => void;
}

export function OperationSelect({ catalog, value, onChange }: Readonly<OperationSelectProps>) {
  // Same policy-gated surface as the roles grant tree (grantableServices):
  // builtin and public operations sit outside the authorization surface
  // this page diagnoses - explaining them would report phantom DENYs.
  const services = useMemo(() => grantableServices(catalog), [catalog]);
  const options = useMemo(() => operationOptions(services), [services]);
  return (
    <Select
      aria-label={m.system_diag_operation()}
      className="w-full"
      onChange={onChange}
      options={options}
      placeholder={m.system_diag_ph_operation()}
      showSearch={{
        filterOption: (input, option) =>
          ((option as { search?: string } | undefined)?.search ?? "")
            .toLowerCase()
            .includes(input.toLowerCase()),
      }}
      value={value}
    />
  );
}
