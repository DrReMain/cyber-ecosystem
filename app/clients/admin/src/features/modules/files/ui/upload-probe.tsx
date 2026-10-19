import { useQuery } from "@connectrpc/connect-query";
import type { File as FileView } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { listUploadedParts } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_presign-FilePresignService_connectquery";
import { Progress } from "antd";

export function UploadProbe({ file }: Readonly<{ file: FileView }>) {
  const id = file.id ?? "";
  const size = Number(file.size ?? 0);
  const { data } = useQuery(listUploadedParts, { id }, { retry: false, staleTime: 15_000 });
  if (!data || size <= 0) {
    return null;
  }
  const partSize = Number(data.partSize ?? 0);
  if (partSize <= 0) {
    return null;
  }
  const partCount = Math.ceil(size / partSize);
  const uploadedBytes = (data.uploaded ?? []).reduce((sum, p) => {
    const n = p.partNumber ?? 0;
    if (n < 1 || n > partCount) {
      return sum;
    }
    return sum + (n === partCount ? size - (n - 1) * partSize : partSize);
  }, 0);
  return (
    <Progress percent={Math.min(100, Math.floor((uploadedBytes / size) * 100))} size="small" />
  );
}
