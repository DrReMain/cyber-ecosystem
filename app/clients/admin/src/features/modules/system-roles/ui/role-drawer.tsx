import type { Role, RoleGrant } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { Button, Drawer, Form, Input, Skeleton, Space, Tabs } from "antd";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import {
  type GrantableView,
  type GrantMap,
  grantMapOf,
  type PolicyOptionView,
  type RoleDraft,
  roleGrantsOf,
} from "../grants";
import { GrantTree } from "./grant-tree";
import { type MemberView, RoleMembers } from "./role-members";

const CODE_PATTERN = /^[a-z][a-z0-9_]*$/;

interface RoleDrawerProps {
  open: boolean;
  mode: "create" | "edit";
  role: Role | null;
  initialGrants: RoleGrant[];
  catalog: GrantableView[];
  policies: PolicyOptionView[];
  members: MemberView[];
  loading: boolean;
  pending: boolean;
  pendingMemberId: string | null;
  afterOpenChange: (open: boolean) => void;
  onCancel: () => void;
  onSave: (draft: RoleDraft) => void;
  onUnbind: (member: MemberView) => void;
}

export function RoleDrawer({
  open,
  mode,
  role,
  initialGrants,
  catalog,
  policies,
  members,
  loading,
  pending,
  pendingMemberId,
  afterOpenChange,
  onCancel,
  onSave,
  onUnbind,
}: Readonly<RoleDrawerProps>) {
  const [form] = Form.useForm<{ name: string; code: string; remark: string }>();
  const [draft, setDraft] = useState<GrantMap | null>(null);
  const grants = draft ?? grantMapOf(initialGrants);
  const initialValues =
    mode === "edit"
      ? { name: role?.name ?? "", code: role?.code ?? "", remark: role?.remark ?? "" }
      : { name: "", code: "", remark: "" };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      onSave({ ...values, remark: values.remark ?? "", grants: roleGrantsOf(grants) });
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
      size={720}
      title={
        mode === "edit" ? m.system_roles_drawer_edit_title() : m.system_roles_drawer_create_title()
      }
    >
      <Form form={form} initialValues={initialValues} layout="vertical">
        <Form.Item
          label={m.system_roles_form_name()}
          name="name"
          rules={[{ required: true, message: m.system_roles_form_name_required() }]}
        >
          <Input maxLength={128} placeholder={m.system_roles_ph_name()} showCount />
        </Form.Item>
        <Form.Item
          label={
            mode === "edit" ? (
              <>
                {m.system_roles_form_code()}
                <span className="ms-1.5 font-normal text-[11px] text-ink-tertiary">
                  {m.system_roles_form_code_hint()}
                </span>
              </>
            ) : (
              m.system_roles_form_code()
            )
          }
          name="code"
          rules={[
            { required: true, message: m.system_roles_form_code_required() },
            { pattern: CODE_PATTERN, message: m.system_roles_form_code_pattern() },
          ]}
        >
          <Input
            className="font-mono"
            disabled={mode === "edit"}
            maxLength={64}
            placeholder={m.system_roles_ph_code()}
            showCount
          />
        </Form.Item>
        <Form.Item label={m.system_roles_form_remark()} name="remark">
          <Input.TextArea
            maxLength={255}
            placeholder={m.system_roles_ph_remark()}
            rows={2}
            showCount
          />
        </Form.Item>
      </Form>
      {loading ? (
        <Skeleton active paragraph={{ rows: 8 }} />
      ) : (
        <Tabs
          items={[
            {
              key: "grants",
              label: m.system_roles_tab_grants(),
              children: (
                <GrantTree
                  catalog={catalog}
                  grants={grants}
                  onChange={setDraft}
                  policies={policies}
                />
              ),
            },
            ...(mode === "edit"
              ? [
                  {
                    key: "members",
                    label: m.system_roles_tab_members(),
                    children: (
                      <RoleMembers
                        members={members}
                        onUnbind={onUnbind}
                        pendingId={pendingMemberId}
                      />
                    ),
                  },
                ]
              : []),
          ]}
        />
      )}
    </Drawer>
  );
}
