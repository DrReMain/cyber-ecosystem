import type { File as UploadedFile } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import type { Role } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import type { User } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { Link } from "@tanstack/react-router";
import {
  Button,
  Drawer,
  Form,
  Input,
  Popconfirm,
  Space,
  TreeSelect,
  Typography,
  Upload,
} from "antd";
import { FileUp, KeyRound, RefreshCw } from "lucide-react";
import { useState } from "react";
import { useFileUrl } from "#/features/app/file/use-file-url";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import type { DeptNode, OperationCatalog, UserDraft } from "../types";
import { GrantPreview } from "./grant-preview";
import { RoleCards } from "./role-cards";
import { UserAvatar } from "./user-avatar";

const PASSWORD_ALPHABET = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789";

function generatePassword(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  let out = "";
  for (const b of bytes) out += PASSWORD_ALPHABET[b % PASSWORD_ALPHABET.length];
  return out;
}

interface UserDrawerProps {
  open: boolean;
  mode: "create" | "edit";
  user: User | null;
  roles: Role[];
  deptTree: DeptNode[];
  catalog: OperationCatalog;
  selected: string[];
  previewOperations: string[];
  previewLoading: boolean;
  pending: boolean;
  resetPending: boolean;
  resetResult: string | null;
  afterOpenChange: (open: boolean) => void;
  onCancel: () => void;
  onUploadAvatar: (file: File) => Promise<UploadedFile | undefined>;
  avatarUploading: boolean;
  onSelectionChange: (next: string[]) => void;
  onSave: (draft: UserDraft) => void;
  onResetPassword: (user: User, password: string) => void;
}

export function UserDrawer({
  open,
  mode,
  user,
  roles,
  deptTree,
  catalog,
  selected,
  previewOperations,
  previewLoading,
  pending,
  resetPending,
  resetResult,
  afterOpenChange,
  onCancel,
  onUploadAvatar,
  avatarUploading,
  onSelectionChange,
  onSave,
  onResetPassword,
}: Readonly<UserDrawerProps>) {
  const [form] = Form.useForm<{ email: string; deptId?: string }>();
  const [password, setPassword] = useState(generatePassword);
  const [avatarId, setAvatarId] = useState<string | undefined>(user?.avatar);
  const previewUrl = useFileUrl(avatarId);
  const initialValues =
    mode === "edit"
      ? { email: user?.email ?? "", deptId: user?.deptId ?? undefined }
      : { email: "", deptId: undefined };

  const handleSave = async () => {
    try {
      const values = await form.validateFields();
      onSave({
        email: values.email,
        password,
        deptId: values.deptId,
        roles: selected,
        avatar: avatarId,
        avatarChanged: avatarId !== (user?.avatar ?? undefined),
      });
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
        mode === "edit" ? m.system_users_drawer_edit_title() : m.system_users_drawer_create_title()
      }
    >
      <Form form={form} initialValues={initialValues} layout="vertical">
        <Form.Item
          label={
            mode === "edit" ? (
              <>
                {m.system_users_form_email()}
                <span className="ms-1.5 font-normal text-[11px] text-black/40 dark:text-white/40">
                  {m.system_users_form_email_hint()}
                </span>
              </>
            ) : (
              m.system_users_form_email()
            )
          }
          name="email"
          rules={[
            { required: true, message: m.system_users_form_email_required() },
            { type: "email", message: m.system_users_form_email_invalid() },
          ]}
        >
          <Input
            disabled={mode === "edit"}
            maxLength={128}
            placeholder={m.system_users_ph_email()}
            showCount
          />
        </Form.Item>
        {mode === "create" && (
          <Form.Item label={m.system_users_form_password()} required>
            <div className="flex items-center gap-2">
              <Typography.Text
                className="flex min-h-8 flex-1 items-center rounded-md border border-black/8 bg-black/2 px-3 font-mono text-[13px] dark:border-white/8 dark:bg-white/5"
                copyable={{ text: password, tooltips: m.system_users_password_copy() }}
              >
                {password}
              </Typography.Text>

              <Button
                color="default"
                icon={<RefreshCw size={13} />}
                onClick={() => setPassword(generatePassword())}
                variant="filled"
              >
                {m.system_users_password_regen()}
              </Button>
            </div>
          </Form.Item>
        )}
        <Form.Item label={m.system_users_form_avatar()}>
          <div className="flex items-center gap-3">
            <UserAvatar email={user?.email ?? ""} size={48} url={previewUrl} />
            <Upload
              beforeUpload={(file) => {
                void onUploadAvatar(file).then((f) => {
                  if (f?.id !== undefined && f.id !== "") setAvatarId(f.id);
                });
                return false;
              }}
              showUploadList={false}
            >
              <Button
                color="default"
                icon={<FileUp size={14} />}
                loading={avatarUploading}
                variant="filled"
              >
                {m.system_users_form_avatar_upload()}
              </Button>
            </Upload>
            {avatarId !== undefined && (
              <Button color="default" onClick={() => setAvatarId(undefined)} variant="text">
                {m.system_users_form_avatar_clear()}
              </Button>
            )}
          </div>
        </Form.Item>
        <Form.Item label={m.system_users_form_dept()} name="deptId">
          <TreeSelect
            allowClear
            fieldNames={{ label: "name", value: "id", children: "children" }}
            placeholder={m.system_users_ph_dept()}
            treeData={deptTree}
            treeDefaultExpandAll
          />
        </Form.Item>
      </Form>
      {mode === "edit" && (
        <div className="mb-4 flex flex-col gap-2">
          <div className="flex items-baseline gap-1">
            <span className="font-medium text-[13px]">{m.system_users_credentials_title()}</span>
            {resetResult !== null && (
              <Typography.Text className="text-[12px]" type="danger">
                {m.system_users_reset_done_hint()}
              </Typography.Text>
            )}
          </div>
          {resetResult !== null ? (
            <div className="flex items-center gap-2">
              <Typography.Text
                className="flex min-h-8 flex-1 items-center rounded-md border border-black/8 bg-black/2 px-3 font-mono text-[13px] dark:border-white/8 dark:bg-white/5"
                copyable={{ text: resetResult, tooltips: m.system_users_password_copy() }}
              >
                {resetResult}
              </Typography.Text>
            </div>
          ) : (
            <Popconfirm
              cancelButtonProps={{ variant: "filled", color: "default" }}
              description={m.system_users_reset_confirm_desc()}
              okButtonProps={{ variant: "filled", color: "danger" }}
              onConfirm={() => user && onResetPassword(user, generatePassword())}
              title={m.system_users_reset_confirm()}
            >
              <Button
                color="default"
                icon={<KeyRound size={13} />}
                loading={resetPending}
                variant="filled"
              >
                {m.system_users_reset_password()}
              </Button>
            </Popconfirm>
          )}
        </div>
      )}
      <div className="mb-2 font-medium text-[13px]">{m.system_users_form_roles()}</div>
      <RoleCards onChange={onSelectionChange} roles={roles} selected={selected} />
      <div className="mt-4 flex flex-col gap-2">
        <div className="flex items-center justify-between">
          <span className="font-medium text-[13px]">{m.system_users_preview_title()}</span>
          {mode === "edit" && user?.id && (
            <Link
              className="text-[13px] text-primary"
              search={{ userId: user.id }}
              to="/dashboard/system/diag"
            >
              {m.system_users_preview_link()} →
            </Link>
          )}
        </div>
        <GrantPreview catalog={catalog} loading={previewLoading} operations={previewOperations} />
      </div>
    </Drawer>
  );
}
