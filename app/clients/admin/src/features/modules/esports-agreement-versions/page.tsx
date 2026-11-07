import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { AgreementVersion } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin_pb";
import {
  createAgreementVersion,
  deleteAgreementVersion,
  listAgreementVersions,
  setAgreementCurrent,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin-AgreementAdminService_connectquery";
import { Button, Card, Input, Modal, Typography } from "antd";
import { Plus } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { NewVersionModal } from "./ui/new-version-modal";
import { VersionsTable } from "./ui/versions-table";
import { docOptions } from "./vocab";

export function AgreementVersionsPage() {
  const [createOpen, setCreateOpen] = useState(false);
  const [viewing, setViewing] = useState<AgreementVersion | null>(null);
  const listQuery = useQuery(listAgreementVersions, {});

  const docLabel = (row: AgreementVersion | null): string =>
    docOptions.find((o) => o.value === String(row?.doc))?.label() ?? "";

  // All three writes take the list with them on success — the current
  // flag and the agreed counts move the very row the operator is
  // looking at.
  const createMut = useMutation(createAgreementVersion, {
    onSuccess: (resp) => {
      toast.success(
        m.esports_agreement_versions_create_ok({
          no: resp.version?.no ?? "",
          type: docLabel(resp.version ?? null),
        }),
      );
      setCreateOpen(false);
      void listQuery.refetch();
    },
  });
  const currentMut = useMutation(setAgreementCurrent, {
    onSuccess: (resp) => {
      toast.success(
        m.esports_agreement_versions_current_ok({
          no: resp.version?.no ?? "",
          type: docLabel(resp.version ?? null),
        }),
      );
      void listQuery.refetch();
    },
  });
  const deleteMut = useMutation(deleteAgreementVersion, {
    onSuccess: () => {
      toast.success(m.esports_agreement_versions_delete_ok());
      void listQuery.refetch();
    },
  });

  return (
    <div className="flex flex-col gap-4 p-4">
      <div className="flex items-start justify-between gap-2">
        <div>
          <Typography.Title className="!mb-0" level={4}>
            {m.esports_agreement_versions_title()}
          </Typography.Title>
          <Typography.Text className="text-xs" type="secondary">
            {m.esports_agreement_versions_desc()}
          </Typography.Text>
        </div>
        <Button icon={<Plus size={14} />} onClick={() => setCreateOpen(true)} type="primary">
          {m.esports_agreement_versions_new()}
        </Button>
      </div>

      <Card>
        <VersionsTable
          loading={listQuery.isFetching}
          onDelete={(row) => deleteMut.mutate({ versionId: row.id ?? "" })}
          onRefresh={() => void listQuery.refetch()}
          onSetCurrent={(row) => currentMut.mutate({ versionId: row.id ?? "" })}
          onView={setViewing}
          rows={listQuery.data?.versions ?? []}
        />
      </Card>

      <NewVersionModal
        onClose={() => setCreateOpen(false)}
        onSubmit={(req) => createMut.mutate(req)}
        open={createOpen}
        submitting={createMut.isPending}
      />

      <Modal
        footer={null}
        onCancel={() => setViewing(null)}
        open={viewing !== null}
        title={m.esports_agreement_versions_view_title({
          no: viewing?.no ?? "",
          type: docLabel(viewing),
        })}
        width={640}
      >
        {/* The prototype's 查看全文: the legal text an operator quotes in a
            dispute, verbatim and read-only. */}
        <Input.TextArea
          autoSize={{ maxRows: 18, minRows: 10 }}
          readOnly
          value={viewing?.body ?? ""}
        />
      </Modal>
    </div>
  );
}
