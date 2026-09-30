import type { Policy } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { Button, Drawer, Form, Input, Skeleton, Space, Switch, Tag, Typography } from "antd";
import { CalendarDays, Clock } from "lucide-react";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { draftOf, type PolicyFormDraft, policyIssues } from "../params";
import { KindCard } from "./kind-card";
import { PolicyCalendarForm } from "./policy-calendar-form";
import { PolicyTimeWindowForm } from "./policy-time-window-form";

const seed = (policy: Policy | null): PolicyFormDraft => {
  const params = policy ? draftOf(policy) : null;
  return {
    name: policy?.name ?? "",
    kind: params?.kind ?? "time_window",
    timeWindow:
      params?.kind === "time_window"
        ? params.timeWindow
        : { windows: [{ start: "09:00", end: "18:00" }], days: [] }, // [] = every day, canonical
    calendar: params?.kind === "calendar" ? params.calendar : { base: "WEEKDAYS", overrides: [] },
    enabled: policy?.enabled ?? true,
  };
};

interface PolicyDrawerProps {
  open: boolean;
  mode: "create" | "edit";
  policy: Policy | null;
  loading: boolean;
  pending: boolean;
  afterOpenChange: (open: boolean) => void;
  onCancel: () => void;
  onSave: (draft: PolicyFormDraft) => void;
}

export function PolicyDrawer({
  open,
  mode,
  policy,
  loading,
  pending,
  afterOpenChange,
  onCancel,
  onSave,
}: Readonly<PolicyDrawerProps>) {
  const [form] = Form.useForm<{ name: string }>();
  // Lazy seed: recompute from the (possibly still-in-flight) policy until the
  // user touches anything, so a late GetPolicy payload lands without a remount.
  const [touched, setTouched] = useState<PolicyFormDraft | null>(null);
  const draft = touched ?? seed(policy);
  const issues = policyIssues(draft);

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      onSave({ ...draft, name: values.name });
    } catch {
      // invalid fields: the per-field feedback is the terminal state here
    }
  };

  return (
    <Drawer
      afterOpenChange={afterOpenChange}
      footer={
        <Space className="w-full justify-end">
          <Button color="default" onClick={onCancel} variant="filled">
            {m.common_cancel()}
          </Button>
          <Button
            color="primary"
            disabled={issues.length > 0}
            loading={pending}
            onClick={() => void handleSave()}
            variant="filled"
          >
            {m.common_save()}
          </Button>
        </Space>
      }
      onClose={onCancel}
      open={open}
      placement={getTextDirection() === "rtl" ? "left" : "right"}
      size={520}
      title={
        mode === "edit"
          ? m.system_policies_drawer_edit_title()
          : m.system_policies_drawer_create_title()
      }
    >
      {loading ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : (
        <div className="flex flex-col gap-4">
          <Form form={form} initialValues={{ name: draft.name }} layout="vertical">
            <Form.Item
              label={m.system_policies_form_name()}
              name="name"
              rules={[{ required: true, message: m.system_policies_form_name_required() }]}
            >
              <Input maxLength={128} placeholder={m.system_policies_ph_name()} showCount />
            </Form.Item>
          </Form>
          <div className="flex flex-col gap-2">
            <span className="font-medium text-[13px]">{m.system_policies_form_kind()}</span>
            {mode === "edit" ? (
              <div className="flex items-center gap-2">
                <Tag>
                  {draft.kind === "time_window"
                    ? m.system_policies_kind_time_window()
                    : m.system_policies_kind_calendar()}
                </Tag>
                <Typography.Text className="text-[12px]" type="secondary">
                  {m.system_policies_form_kind_locked()}
                </Typography.Text>
              </div>
            ) : (
              <div className="flex gap-3">
                <KindCard
                  desc={m.system_policies_form_kind_time_window_desc()}
                  icon={<Clock size={15} />}
                  onSelect={() => setTouched({ ...draft, kind: "time_window" })}
                  selected={draft.kind === "time_window"}
                  title={m.system_policies_kind_time_window()}
                />
                <KindCard
                  desc={m.system_policies_form_kind_calendar_desc()}
                  icon={<CalendarDays size={15} />}
                  onSelect={() => setTouched({ ...draft, kind: "calendar" })}
                  selected={draft.kind === "calendar"}
                  title={m.system_policies_kind_calendar()}
                />
              </div>
            )}
          </div>
          {draft.kind === "time_window" ? (
            <PolicyTimeWindowForm
              draft={draft.timeWindow}
              onChange={(timeWindow) => setTouched({ ...draft, timeWindow })}
            />
          ) : (
            <PolicyCalendarForm
              draft={draft.calendar}
              onChange={(calendar) => setTouched({ ...draft, calendar })}
            />
          )}
          <div className="flex flex-col gap-1 border-black/8 border-t pt-3 dark:border-white/8">
            <div className="flex items-center gap-2">
              <Switch
                checked={draft.enabled}
                onChange={(enabled) => setTouched({ ...draft, enabled })}
              />
              <span>{m.system_policies_form_enabled()}</span>
            </div>
            <Typography.Text className="text-[12px]" type="warning">
              {m.system_policies_relax_hint()}
            </Typography.Text>
          </div>
        </div>
      )}
    </Drawer>
  );
}
