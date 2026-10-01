import { useMutation } from "@connectrpc/connect-query";
import type { File as FileView } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { uploadFile } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_proxy-FileProxyService_connectquery";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";

const MAX_PROXY_UPLOAD_BYTES = 4 * 1024 * 1024;

export function useProxyUpload() {
  const mutation = useMutation(uploadFile);
  const upload = async (file: File): Promise<FileView | undefined> => {
    if (file.size > MAX_PROXY_UPLOAD_BYTES) {
      toast.error(
        m.common_upload_too_large({ max: `${MAX_PROXY_UPLOAD_BYTES / (1024 * 1024)} MB` }),
      );
      return undefined;
    }
    try {
      const buf = await file.arrayBuffer();
      const resp = await mutation.mutateAsync({
        name: file.name,
        contentType: file.type === "" ? "application/octet-stream" : file.type,
        data: new Uint8Array(buf),
      });
      return resp.file;
    } catch {
      // The mutation cache's error policy already surfaced the copy.
      return undefined;
    }
  };
  return { upload, isPending: mutation.isPending };
}
