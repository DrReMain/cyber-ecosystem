import type { Announcement } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Tag } from "antd";
import { Ban, Eye, Pencil, Rocket } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";
import { kindTags, scopeLabel, statusTags } from "../vocab";

interface AnnouncementsTableProps {
  rows: Announcement[];
  loading?: boolean;
  onEdit: (row: Announcement) => void;
  onOffline: (row: Announcement) => void;
  onPublish: (row: Announcement) => void;
  onRefresh: () => void;
  onView: (row: Announcement) => void;
}

export function AnnouncementsTable({
  rows,
  loading,
  onEdit,
  onOffline,
  onPublish,
  onRefresh,
  onView,
}: Readonly<AnnouncementsTableProps>) {
  const { fieldAction } = useColumns();

  const columns: TableProps<Announcement>["columns"] = [
    {
      title: () => m.esports_announcements_col_title(),
      dataIndex: "title",
      render: (_, row) => <span className="font-semibold">{row.title ?? "-"}</span>,
    },
    {
      title: () => m.esports_announcements_col_kind(),
      dataIndex: "kind",
      width: 100,
      render: (_, row) => {
        const tag = kindTags[String(row.kind ?? "")];
        return tag ? (
          <Tag className="m-0" color={tag.color}>
            {tag.label()}
          </Tag>
        ) : (
          String(row.kind ?? "")
        );
      },
    },
    {
      title: () => m.esports_announcements_col_scope(),
      dataIndex: "scope",
      width: 110,
      render: (_, row) => scopeLabel(String(row.scope ?? "")),
    },
    {
      title: () => m.esports_announcements_col_status(),
      dataIndex: "status",
      width: 90,
      render: (_, row) => {
        const tag = statusTags[String(row.status ?? "")];
        return tag ? (
          <Tag className="m-0" color={tag.color}>
            {tag.label()}
          </Tag>
        ) : (
          String(row.status ?? "")
        );
      },
    },
    {
      title: () => m.esports_announcements_col_published(),
      dataIndex: "publishedAt",
      width: 170,
      // Drafts have no delivery moment — the dash is the honest marker,
      // not a missing publish.
      render: (_, row) => (row.publishedAt ? fullTime(row.publishedAt) : "-"),
    },
    {
      ...fieldAction<Announcement>(
        (_, row) => {
          const draft = String(row.status) === "ANNOUNCEMENT_STATUS_DRAFT";
          const published = String(row.status) === "ANNOUNCEMENT_STATUS_PUBLISHED";
          return (
            <div className="flex gap-1">
              <ActionIconButton
                icon={Eye}
                label={m.esports_announcements_act_view()}
                onClick={() => onView(row)}
              />
              {draft ? (
                <ActionIconButton
                  icon={Pencil}
                  label={m.esports_announcements_act_edit()}
                  onClick={() => onEdit(row)}
                />
              ) : null}
              {draft ? (
                <ActionIconButton
                  color="primary"
                  confirm={{
                    // A banner goes live without delivering anything — the
                    // scope-bearing sentence belongs to the letter kind
                    // alone.
                    description:
                      String(row.kind) === "ANNOUNCEMENT_KIND_BANNER"
                        ? m.esports_announcements_publish_confirm_desc_banner()
                        : m.esports_announcements_publish_confirm_desc({
                            scope: scopeLabel(String(row.scope ?? "")),
                          }),
                    onConfirm: () => onPublish(row),
                    title: m.esports_announcements_publish_confirm_title({
                      title: row.title ?? "",
                    }),
                  }}
                  icon={Rocket}
                  label={m.esports_announcements_act_publish()}
                />
              ) : null}
              {published ? (
                <ActionIconButton
                  color="danger"
                  confirm={{
                    // Offlining a banner removes a live surface; offlining
                    // a letter row only archives it — the letters already
                    // in the boxes stay put.
                    description:
                      String(row.kind) === "ANNOUNCEMENT_KIND_BANNER"
                        ? m.esports_announcements_offline_confirm_desc()
                        : m.esports_announcements_offline_confirm_desc_letter(),
                    onConfirm: () => onOffline(row),
                    title: m.esports_announcements_offline_confirm_title({
                      title: row.title ?? "",
                    }),
                  }}
                  icon={Ban}
                  label={m.esports_announcements_act_offline()}
                />
              ) : null}
            </div>
          );
        },
        { title: m.esports_announcements_col_actions() },
      ),
      width: 130,
    },
  ];

  return (
    <DataTable<Announcement>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_announcements_empty()}
      loading={loading}
      pagination={false}
      rowKey={(r) => r.id ?? ""}
      toolbar={{
        labels: {
          density: m.common_toolbar_density(),
          densityLarge: m.common_toolbar_density_large(),
          densityMiddle: m.common_toolbar_density_middle(),
          densitySmall: m.common_toolbar_density_small(),
          refresh: m.common_toolbar_refresh(),
        },
        loading,
        onRefresh,
      }}
    />
  );
}
