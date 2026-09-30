import type { Dept } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";
import { Button, Drawer, Form, Input, Space, TreeSelect } from "antd";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import type { DeptDraft, DeptNode } from "../types";
import { DeptMembers, type DeptMemberView } from "./dept-members";

interface DeptDrawerProps {
  open: boolean;
  mode: "create" | "create-child" | "edit";
  source: Dept | null; // edit target, or preset parent for create-child
  parentTree: DeptNode[];
  pending: boolean;
  members: DeptMemberView[];
  pendingMemberId: string | null;
  afterOpenChange: (open: boolean) => void;
  onCancel: () => void;
  onSave: (draft: DeptDraft) => void;
  onRemoveMember: (member: DeptMemberView) => void;
}

export function DeptDrawer({
  open,
  mode,
  source,
  parentTree,
  pending,
  members,
  pendingMemberId,
  afterOpenChange,
  onCancel,
  onSave,
  onRemoveMember,
}: Readonly<DeptDrawerProps>) {
  const [form] = Form.useForm<{ name: string; parentId?: string; remark?: string }>();
  const initialValues =
    mode === "edit"
      ? {
          name: source?.name ?? "",
          parentId: source?.parentId ?? undefined,
          remark: source?.remark ?? "",
        }
      : {
          name: "",
          parentId: mode === "create-child" ? (source?.id ?? undefined) : undefined,
          remark: "",
        };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      onSave({ name: values.name, parentId: values.parentId, remark: values.remark ?? "" });
    } catch {
      // invalid fields: the per-field feedback is the terminal state here
    }
  };

  const title =
    mode === "edit"
      ? m.system_depts_drawer_edit_title()
      : mode === "create-child"
        ? m.system_depts_drawer_create_child_title()
        : m.system_depts_drawer_create_title();

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
      size={480}
      title={title}
    >
      <Form form={form} initialValues={initialValues} layout="vertical">
        <Form.Item
          label={m.system_depts_form_name()}
          name="name"
          rules={[{ required: true, message: m.system_depts_form_name_required() }]}
        >
          <Input maxLength={64} placeholder={m.system_depts_ph_name()} showCount />
        </Form.Item>
        <Form.Item label={m.system_depts_form_parent()} name="parentId">
          <TreeSelect
            allowClear
            fieldNames={{ label: "name", value: "id", children: "children" }}
            placeholder={m.system_depts_ph_parent()}
            treeData={parentTree}
            treeDefaultExpandAll
          />
        </Form.Item>
        <Form.Item label={m.system_depts_form_remark()} name="remark">
          <Input.TextArea
            maxLength={255}
            placeholder={m.system_depts_ph_remark()}
            rows={2}
            showCount
          />
        </Form.Item>
      </Form>
      {mode === "edit" && (
        <div className="flex flex-col gap-2 border-black/8 border-t pt-3 dark:border-white/8">
          <span className="font-medium text-[13px]">{m.system_depts_members_label()}</span>
          <DeptMembers
            deptId={source?.id ?? ""}
            members={members}
            onRemove={onRemoveMember}
            pendingId={pendingMemberId}
          />
        </div>
      )}
    </Drawer>
  );
}
