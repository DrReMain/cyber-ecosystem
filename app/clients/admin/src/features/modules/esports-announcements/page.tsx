import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { Announcement } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_admin_pb";
import {
  createAnnouncement,
  listAnnouncements,
  offlineAnnouncement,
  publishAnnouncement,
  updateAnnouncement,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_admin-MessageAdminService_connectquery";
import { Button, Modal, Tag, Typography } from "antd";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";
import { type AnnouncementDraftSeed, AnnouncementFormModal } from "./ui/announcement-form-modal";
import { AnnouncementsTable } from "./ui/announcements-table";
import { kindLabel, kindTags, scopeLabel, statusTags } from "./vocab";

export function AnnouncementsPage() {
  const [viewing, setViewing] = useState<Announcement | null>(null);
  const listQuery = useQuery(listAnnouncements, {});
  // The form's antd store outlives field unmount, so a reopened modal
  // would carry the last draft's values — every open mints a fresh
  // component (fresh form instance, fresh initialValues) instead.
  const [formSession, setFormSession] = useState<{
    key: number;
    seed: AnnouncementDraftSeed | null;
  }>({ key: 0, seed: null });
  const openCreate = () =>
    setFormSession((s) => ({ key: s.key + 1, seed: { content: "", title: "" } }));
  const openEdit = (row: Announcement) =>
    setFormSession((s) => ({
      key: s.key + 1,
      seed: {
        content: row.content ?? "",
        id: row.id ?? "",
        kind: row.kind,
        scope: row.scope,
        title: row.title ?? "",
      },
    }));
  const closeForm = () => setFormSession((s) => ({ ...s, seed: null }));

  // Create lands a draft, never a delivery — publishing is a separate,
  // deliberate move on the row, which is the contract's whole shape.
  const createMut = useMutation(createAnnouncement, {
    onSuccess: (resp) => {
      closeForm();
      void toast.success(
        m.esports_announcements_create_ok({ title: resp.announcement?.title ?? "" }),
      );
      void listQuery.refetch();
    },
  });
  const updateMut = useMutation(updateAnnouncement, {
    onSuccess: (resp) => {
      closeForm();
      void toast.success(
        m.esports_announcements_update_ok({ title: resp.announcement?.title ?? "" }),
      );
      void listQuery.refetch();
    },
  });
  const publishMut = useMutation(publishAnnouncement, {
    onSuccess: (resp) => {
      void toast.success(
        m.esports_announcements_publish_ok({
          scope: scopeLabel(String(resp.announcement?.scope ?? "")),
          title: resp.announcement?.title ?? "",
        }),
      );
      void listQuery.refetch();
    },
  });
  const offlineMut = useMutation(offlineAnnouncement, {
    onSuccess: (resp) => {
      void toast.success(
        m.esports_announcements_offline_ok({ title: resp.announcement?.title ?? "" }),
      );
      void listQuery.refetch();
    },
  });

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-start justify-between gap-2">
        <div>
          <Typography.Title className="!mb-0" level={4}>
            {m.esports_announcements_title()}
          </Typography.Title>
          <Typography.Text className="text-xs" type="secondary">
            {m.esports_announcements_desc()}
          </Typography.Text>
        </div>
        <Button icon={<Plus size={14} />} onClick={openCreate} type="primary">
          {m.esports_announcements_new()}
        </Button>
      </div>

      <AnnouncementsTable
        loading={listQuery.isLoading}
        onEdit={openEdit}
        onOffline={(row) => offlineMut.mutate({ announcementId: row.id ?? "" })}
        onPublish={(row) => publishMut.mutate({ announcementId: row.id ?? "" })}
        onRefresh={() => void listQuery.refetch()}
        onView={setViewing}
        rows={listQuery.data?.announcements ?? []}
      />

      <AnnouncementFormModal
        key={formSession.key}
        onClose={closeForm}
        onSubmit={(input) => {
          if (input.id !== undefined) {
            updateMut.mutate({
              announcementId: input.id,
              content: input.content,
              title: input.title,
            });
            return;
          }
          if (typeof input.kind !== "number" || typeof input.scope !== "number") {
            return;
          }
          createMut.mutate({
            content: input.content,
            kind: input.kind,
            scope: input.scope,
            title: input.title,
          });
        }}
        seed={formSession.seed}
        submitting={createMut.isPending || updateMut.isPending}
      />

      <Modal footer={null} onCancel={() => setViewing(null)} open={viewing !== null} width={640}>
        {viewing ? (
          <div className="flex flex-col gap-3">
            <Typography.Title className="!mb-0" level={4}>
              {m.esports_announcements_view_title({ title: viewing.title ?? "" })}
            </Typography.Title>
            <div className="flex flex-wrap gap-2">
              <Tag color={kindTags[String(viewing.kind ?? "")]?.color}>
                {kindLabel(String(viewing.kind ?? ""))}
              </Tag>
              <Tag>{scopeLabel(String(viewing.scope ?? ""))}</Tag>
              <Tag color={statusTags[String(viewing.status ?? "")]?.color}>
                {statusTags[String(viewing.status ?? "")]?.label() ?? "-"}
              </Tag>
            </div>
            <Typography.Paragraph
              className="whitespace-pre-wrap"
              style={{ background: "var(--ant-color-fill-tertiary)", padding: 12, borderRadius: 8 }}
            >
              {viewing.content ?? ""}
            </Typography.Paragraph>
            {viewing.publishedAt ? (
              <Typography.Text className="text-xs" type="secondary">
                {m.esports_announcements_view_published_at({ time: fullTime(viewing.publishedAt) })}
              </Typography.Text>
            ) : null}
          </div>
        ) : null}
      </Modal>
    </div>
  );
}
