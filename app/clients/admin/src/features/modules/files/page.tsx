import { useMutation, useQuery } from "@connectrpc/connect-query";
import { FileSource, FileStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import {
  deleteFile,
  listFiles,
  renameFile,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file-FileService_connectquery";
import { UserService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter, useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useRouteContext, useSearch } from "@tanstack/react-router";
import { Button, Card, Input, Modal, Upload } from "antd";
import { FileUp } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { useProxyUpload } from "#/features/app/use-proxy-upload";
import { useUserDirectory } from "#/features/app/use-user-directory";
import { isOperationAllowed, op } from "#/features/layout-dashboard/auth/permissions";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { filesSearchSchema, parseFilesSearch } from "./search";
import { FilesTable } from "./ui/files-table";

const clean = (v: string | undefined): string | undefined =>
  v !== undefined && v !== "" ? v : undefined;

const AREA = "/dashboard/files/list";

export function FilesPage() {
  const search = useSearch({ from: AREA });
  const { permissions } = useRouteContext({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseFilesSearch);
  const { values, onFilter, onReset } = useFilter(store, filesSearchSchema);
  const [renaming, setRenaming] = useState<{ id: string; name: string } | null>(null);

  const { byId: usersById, known: ownerKnown } = useUserDirectory(
    isOperationAllowed(permissions, op(UserService, "listUsers")),
  );

  const input = useMemo(
    () => ({
      page: {
        pageNo: search.pageNo,
        pageSize: search.pageSize,
      },
      orderBy: [search.sort ?? "createdAt:desc"],
      name: clean(values.name),
      contentType: clean(values.contentType),
      status: values.status === undefined ? undefined : FileStatus[values.status],
      source: values.source === undefined ? undefined : FileSource[values.source],
    }),
    [values, search.pageNo, search.pageSize, search.sort],
  );

  const filesQuery = useQuery(listFiles, input);
  const pagination = useServerPagination(store, filesSearchSchema, filesQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.files_total({ n: t }),
  });
  const refetch = () => void filesQuery.refetch();
  const deleteMutation = useMutation(deleteFile, { onSuccess: refetch });
  const renameMutation = useMutation(renameFile, {
    onSuccess: () => {
      setRenaming(null);
      refetch();
    },
  });
  const { upload: proxyUpload, isPending: uploadPending } = useProxyUpload();
  const onUploadFile = (file: File) => {
    void proxyUpload(file).then((f) => {
      if (!f) return;
      refetch();
      toast.success(m.files_upload_done({ name: f.name ?? "" }));
    });
  };
  const uploadEntry = (
    <Upload
      beforeUpload={(file) => {
        onUploadFile(file);
        return false;
      }}
      showUploadList={false}
    >
      <Button color="primary" icon={<FileUp size={14} />} loading={uploadPending} variant="filled">
        {m.files_upload()}
      </Button>
    </Upload>
  );

  const filtersActive =
    clean(values.name) !== undefined ||
    clean(values.contentType) !== undefined ||
    values.status !== undefined ||
    values.source !== undefined;

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <Filter
          columns={3}
          initialValues={values}
          labels={{
            search: m.common_filter_search(),
            reset: m.common_filter_reset(),
            fold: m.common_filter_fold(),
            expand: m.common_filter_expand(),
          }}
          onFilter={onFilter}
          onReset={onReset}
          options={[
            {
              label: m.files_filter_name(),
              name: "name",
              placeholder: m.files_filter_name_ph(),
            },
            {
              label: m.files_filter_type(),
              name: "contentType",
              placeholder: m.files_filter_type_ph(),
            },
            {
              label: m.files_filter_status(),
              name: "status",
              placeholder: m.files_filter_status_ph(),
              type: "select",
              options: [
                { label: m.files_status_confirmed(), value: "CONFIRMED" },
                { label: m.files_status_uploading(), value: "UPLOADING" },
              ],
            },
            {
              label: m.files_filter_source(),
              name: "source",
              placeholder: m.files_filter_source_ph(),
              type: "select",
              options: [
                { label: m.files_source_client(), value: "CLIENT_UPLOAD" },
                { label: m.files_source_server(), value: "SERVER_GENERATED" },
              ],
            },
          ]}
        />
      </Card>
      <FilesTable
        emptyDescription={filtersActive ? m.files_empty_filtered() : m.files_empty()}
        files={filesQuery.data?.list ?? []}
        loading={filesQuery.isFetching}
        onDelete={(file) => deleteMutation.mutate({ id: file.id ?? "" })}
        onRefresh={refetch}
        onRename={(file) => setRenaming({ id: file.id ?? "", name: file.name ?? "" })}
        ownerKnown={ownerKnown}
        ownerName={(id) => usersById.get(id)}
        pagination={pagination}
        pendingDeleteId={deleteMutation.isPending ? (deleteMutation.variables?.id ?? null) : null}
        toolbarExtra={uploadEntry}
      />
      <Modal
        cancelButtonProps={{ variant: "filled", color: "default" }}
        cancelText={m.common_cancel()}
        okButtonProps={{ loading: renameMutation.isPending, variant: "filled", color: "primary" }}
        okText={m.common_save()}
        onCancel={() => setRenaming(null)}
        onOk={() => renaming && renameMutation.mutate({ id: renaming.id, name: renaming.name })}
        open={renaming !== null}
        title={m.files_rename_title()}
      >
        <Input
          onChange={(e) => setRenaming((prev) => (prev ? { ...prev, name: e.target.value } : prev))}
          placeholder={m.files_rename_label()}
          value={renaming?.name}
        />
      </Modal>
    </div>
  );
}
