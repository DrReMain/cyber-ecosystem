import type { ScopeKind } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { Button, Checkbox, Typography } from "antd";
import { Asterisk, ChevronDown, X } from "lucide-react";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import {
  type GrantableView,
  type GrantMap,
  opOf,
  type PolicyOptionView,
  serviceDatascopeAny,
  wildcardOpOf,
} from "../grants";
import { MethodRow } from "./method-row";
import { PolicySubRow } from "./policy-sub-row";
import { PolicyToggle } from "./policy-toggle";
import { ScopeSelect } from "./scope-select";

interface ServiceBlockProps {
  svc: GrantableView;
  grants: GrantMap;
  policies: PolicyOptionView[];
  globalOn: boolean;
  open: boolean;
  onToggle: () => void;
  onSetScope: (op: string, scope: ScopeKind | undefined) => void;
  onSetPolicies: (op: string, policies: string[]) => void;
  onSetService: (svc: GrantableView, on: boolean) => void;
  onSetWildcard: (svc: GrantableView) => void;
}

export function ServiceBlock({
  svc,
  grants,
  policies,
  globalOn,
  open,
  onToggle,
  onSetScope,
  onSetPolicies,
  onSetService,
  onSetWildcard,
}: Readonly<ServiceBlockProps>) {
  const wildcard = wildcardOpOf(svc.fullName);
  const wildOn = grants[wildcard] !== undefined;
  const coveredBy = wildOn || globalOn;
  const ops = svc.methods.map((mt) => opOf(svc.fullName, mt.name));
  const checked = ops.filter((op) => grants[op] !== undefined).length;
  const covered = coveredBy ? svc.methods.length : checked;
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const toggleExpanded = (op: string) =>
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(op)) next.delete(op);
      else next.add(op);
      return next;
    });
  const subRowOpen = (op: string): boolean =>
    expanded.has(op) || (grants[op]?.policies.length ?? 0) > 0;

  return (
    <div className="rounded-lg border border-black/8 dark:border-white/8">
      <div className="flex items-center gap-2 border-black/5 border-b px-3 py-2 dark:border-white/5">
        <Checkbox
          checked={covered === svc.methods.length && covered > 0}
          disabled={globalOn}
          indeterminate={covered > 0 && covered < svc.methods.length}
          onChange={(e) => onSetService(svc, e.target.checked)}
        >
          <span className="font-medium">{svc.comment || svc.name}</span>
        </Checkbox>
        <Typography.Text className={`font-mono text-[14px] ${covered > 0 ? "text-primary" : ""}`}>
          {covered}/{svc.methods.length}
        </Typography.Text>
        <button
          aria-expanded={open}
          className="flex min-w-0 flex-1 cursor-pointer items-center gap-2 border-0 bg-transparent p-0 text-start"
          onClick={onToggle}
          type="button"
        >
          <span className="min-w-0 flex-1 truncate font-mono text-[13px] text-black/40 dark:text-white/40">
            {svc.fullName}
          </span>
          <ChevronDown
            className={`size-3.5 shrink-0 text-black/35 transition-transform dark:text-white/35 ${open ? "" : "-rotate-90"}`}
          />
        </button>
      </div>
      {open ? (
        <div className="flex flex-col">
          {wildOn && (
            <div className="m-2 flex min-h-9 flex-col gap-1 rounded border border-primary border-dashed bg-primary/5 px-3 py-2">
              <div className="flex flex-col gap-1">
                <div className="flex gap-2">
                  <span className="text-[14px]">{m.system_roles_perm_wildcard()}</span>
                  <span className="inline-flex items-center gap-2 rounded bg-black/9 px-1.5 py-0.5 font-mono text-[12px] dark:bg-black">
                    {svc.fullName}/*
                  </span>
                </div>

                <div className="flex min-h-9 items-center gap-2">
                  <div className="flex-1 ps-9">
                    {serviceDatascopeAny(svc) && (
                      <ScopeSelect
                        onChange={(value) => onSetScope(wildcard, value)}
                        value={grants[wildcard]?.scope}
                      />
                    )}
                  </div>
                  <PolicyToggle
                    count={grants[wildcard]?.policies.length ?? 0}
                    onClick={() => toggleExpanded(wildcard)}
                  />
                  <Button
                    aria-label="clear"
                    icon={<X size={14} />}
                    onClick={() => onSetScope(wildcard, undefined)}
                    size="small"
                    type="text"
                  />
                </div>
              </div>
              {subRowOpen(wildcard) ? (
                <PolicySubRow
                  grants={grants}
                  onSetPolicies={onSetPolicies}
                  op={wildcard}
                  policies={policies}
                />
              ) : null}
            </div>
          )}
          {svc.methods.map((mt) => (
            <MethodRow
              coveredBy={coveredBy}
              expanded={subRowOpen(opOf(svc.fullName, mt.name))}
              grants={grants}
              key={mt.name}
              method={mt}
              onSetPolicies={onSetPolicies}
              onSetScope={onSetScope}
              onToggleExpand={() => toggleExpanded(opOf(svc.fullName, mt.name))}
              policies={policies}
              service={svc.fullName}
            />
          ))}
          {!(wildOn || globalOn) && (
            <div className="px-3 py-1.5">
              <Button
                block
                color="default"
                icon={<Asterisk size={12} />}
                onClick={() => onSetWildcard(svc)}
                size="small"
                variant="dashed"
              >
                {m.system_roles_perm_wildcard_action()}
              </Button>
            </div>
          )}
        </div>
      ) : null}
    </div>
  );
}
