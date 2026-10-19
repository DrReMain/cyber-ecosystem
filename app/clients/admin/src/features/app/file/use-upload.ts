import { useMutation } from "@connectrpc/connect-query";
import type { File as FileView } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import { FileStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_pb";
import type {
  CreateUploadResponse,
  ListUploadedPartsResponse,
  UploadedPart,
  UploadPartUrl,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_presign_pb";
import {
  abortUpload,
  confirmUpload,
  createUpload,
  listUploadedParts,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/file_presign-FilePresignService_connectquery";
import { useRef, useState } from "react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { putUrl } from "./put-url";
import { S3Error } from "./s3-error";

const PART_CONCURRENCY = 6;
const PART_ATTEMPTS = 2;

export interface UploadProgress {
  loaded: number;
  total: number;
}

interface PartETag {
  partNumber: number;
  etag: string;
}

// The RPC legs and the per-drive outlets the upload orchestration needs;
// the hook supplies them from its mutations and its controller registry,
// the module-level drivers consume them. All per-drive state is keyed by
// the upload row id — the table renders progress by that key, and aborting
// a row cancels that drive's in-flight PUTs through it.
interface ChannelDeps {
  create: (req: {
    name: string;
    contentType: string;
    size: bigint;
  }) => Promise<CreateUploadResponse>;
  confirm: (req: { id: string; parts?: PartETag[] }) => Promise<{ file?: FileView }>;
  abort: (req: { id: string }) => Promise<unknown>;
  listParts: (req: { id: string }) => Promise<ListUploadedPartsResponse>;
  onCreated?: () => void;
  beginDrive: (id: string) => AbortSignal;
  endDrive: (id: string) => void;
  setProgress: (id: string, p: UploadProgress) => void;
}

interface UseUploadOptions {
  // Fired once the metadata row exists, so the caller can refetch the list
  // and surface the in-flight row (with its progress) immediately.
  onCreated?: () => void;
}

export function useUpload(opts: UseUploadOptions = {}) {
  const createMutation = useMutation(createUpload);
  const confirmMutation = useMutation(confirmUpload);
  const listMutation = useMutation(listUploadedParts);
  const abortMutation = useMutation(abortUpload);
  const [progressById, setProgressById] = useState<Record<string, UploadProgress>>({});
  const [uploading, setUploading] = useState(false);
  const drivesRef = useRef(new Map<string, AbortController>());

  const deps: ChannelDeps = {
    create: (req) => createMutation.mutateAsync(req),
    confirm: (req) => confirmMutation.mutateAsync(req),
    abort: (req) => abortMutation.mutateAsync(req),
    listParts: (req) => listMutation.mutateAsync(req),
    onCreated: opts.onCreated,
    // Re-entering an id (a second resume click) cancels the earlier drive
    // instead of letting two part pools interleave on one row.
    beginDrive: (id) => {
      drivesRef.current.get(id)?.abort();
      const controller = new AbortController();
      drivesRef.current.set(id, controller);
      return controller.signal;
    },
    endDrive: (id) => {
      drivesRef.current.delete(id);
      setProgressById((prev) => {
        if (!(id in prev)) {
          return prev;
        }
        const next = { ...prev };
        delete next[id];
        return next;
      });
    },
    setProgress: (id, p) => setProgressById((prev) => ({ ...prev, [id]: p })),
  };

  // The direct channel takes any size — routing between proxy and direct is
  // the caller's (the page's explicit buttons), not the library's.
  const uploadDirect = async (file: File): Promise<FileView | undefined> => {
    setUploading(true);
    try {
      return await directUpload(file, true, deps);
    } finally {
      setUploading(false);
    }
  };

  // Resume re-entry: every piece of client state is rehydrated from the
  // server (row via listUploadedParts, part truth via ListParts, fresh URLs
  // via the mint) — the caller only supplies a fresh file handle.
  const resume = async (file: File, id: string): Promise<FileView | undefined> => {
    setUploading(true);
    try {
      return await resumeUpload(file, id, deps);
    } finally {
      setUploading(false);
    }
  };

  // Aborting a row whose bytes are still moving: stop them before the RPC
  // tears down the session, or a PUT racing the delete lands an orphaned
  // object behind it.
  const cancelUpload = (id: string) => drivesRef.current.get(id)?.abort();

  return { uploadDirect, resume, cancelUpload, progressById, uploading };
}

async function directUpload(
  file: File,
  allowNewSession: boolean,
  deps: ChannelDeps,
): Promise<FileView | undefined> {
  let id = "";
  try {
    const created = await deps.create({
      name: file.name,
      contentType: contentTypeOf(file),
      size: BigInt(file.size),
    });
    id = created.file?.id ?? "";
    deps.onCreated?.();
    // Responses arrive as toJson JSON (the transport's anti-corruption
    // layer): a oneof member renders inline — created.single /
    // created.multipart, with no channel wrapper. The generated type
    // claims otherwise; it lies here the same way it does for proto maps.
    const flat = created as unknown as {
      single?: { url?: string };
      multipart?: { partUrls?: UploadPartUrl[]; partSize?: bigint | number | string };
    };
    if (flat.single) {
      return await singleLeg(
        file,
        id,
        flat.single.url ?? "",
        {
          abortSession: () => deps.abort({ id }),
          confirm: () => deps.confirm({ id }),
          setProgress: (p) => deps.setProgress(id, p),
        },
        deps.beginDrive(id),
      );
    }
    if (flat.multipart) {
      const parts = await driveParts(
        file,
        id,
        Number(flat.multipart.partSize ?? 0),
        flat.multipart.partUrls ?? [],
        [],
        true,
        deps,
        deps.beginDrive(id),
      );
      return (await deps.confirm({ id, parts })).file ?? undefined;
    }
    return undefined;
  } catch (err) {
    // The aborted single-PUT session is already gone; retry once on a fresh
    // session instead of surfacing the expiry. A second expiry surfaces as
    // the raw S3 failure — a TTL shorter than the transfer is pathological,
    // not worth another silent loop.
    if (err instanceof SessionExpiredError && allowNewSession) {
      return directUpload(file, false, deps);
    }
    if (err instanceof UploadCancelledError) {
      return undefined;
    }
    toastS3Failure(err);
    return undefined;
  } finally {
    // Covers the expired-session retry too: the recursive call owns its own
    // (new-id) drive; this releases the abandoned one.
    if (id !== "") {
      deps.endDrive(id);
    }
  }
}

async function resumeUpload(
  file: File,
  id: string,
  deps: ChannelDeps,
): Promise<FileView | undefined> {
  const signal = deps.beginDrive(id);
  try {
    const info = await deps.listParts({ id });
    if (info.file?.status === FileStatus.CONFIRMED) {
      return info.file;
    }
    const partSize = Number(info.partSize);
    // partSize === size marks the degenerate single-PUT session (its only
    // part IS the object; a multipart plan is always partSize < size).
    // Resume = direct confirm if the PUT landed, else one whole-object PUT
    // on the re-minted URL — the same leg as a fresh upload, content-type
    // included.
    if (partSize === file.size) {
      const part = (info.missing ?? []).find((p) => p.partNumber === 1);
      if (part?.url) {
        deps.setProgress(id, { loaded: 0, total: file.size });
        await putUrl(part.url, file, {
          contentType: contentTypeOf(file),
          signal,
          onProgress: (loaded) => deps.setProgress(id, { loaded, total: file.size }),
        });
      }
      return (await deps.confirm({ id })).file ?? undefined;
    }
    const uploaded = (info.uploaded ?? []).map(toPartETag);
    const parts = await driveParts(
      file,
      id,
      partSize,
      info.missing ?? [],
      uploaded,
      true,
      deps,
      signal,
    );
    return (await deps.confirm({ id, parts })).file ?? undefined;
  } catch (err) {
    if (err instanceof UploadCancelledError) {
      return undefined;
    }
    toastS3Failure(err);
    return undefined;
  } finally {
    deps.endDrive(id);
  }
}

interface SingleDeps {
  abortSession: () => Promise<unknown>;
  confirm: () => Promise<{ file?: FileView }>;
  setProgress: (p: UploadProgress) => void;
}

// A single-PUT session cannot re-mint its URL — on expiry the only recovery
// is a fresh session, so the caller aborts and re-enters once.
async function singleLeg(
  file: File,
  id: string,
  url: string,
  deps: SingleDeps,
  signal: AbortSignal,
  retry = true,
): Promise<FileView | undefined> {
  deps.setProgress({ loaded: 0, total: file.size });
  try {
    await putUrl(url, file, {
      contentType: contentTypeOf(file),
      signal,
      onProgress: (loaded) => deps.setProgress({ loaded, total: file.size }),
    });
  } catch (err) {
    if (signal.aborted) {
      throw new UploadCancelledError();
    }
    if (err instanceof S3Error && err.kind === "expired" && retry) {
      await deps.abortSession();
      throw new SessionExpiredError(id);
    }
    throw err;
  }
  return (await deps.confirm()).file ?? undefined;
}

class SessionExpiredError extends Error {
  constructor(readonly id: string) {
    super("presign session expired before the object landed");
  }
}

// A deliberate cancel (abort clicked mid-flight): unwinds the drive
// silently — the abort RPC owns the user feedback and the table refetch.
class UploadCancelledError extends Error {
  constructor() {
    super("upload drive cancelled");
  }
}

// The only place multipart part bytes move. A worker pool drains the missing-parts
// queue; per-part network retries are local (part re-PUT overwrites
// idempotently). An expired URL stops the whole pool and resyncs once at
// this level — a resync started inside a worker would race the others.
async function driveParts(
  file: File,
  id: string,
  partSize: number,
  partUrls: UploadPartUrl[],
  uploadedETags: PartETag[],
  allowResync: boolean,
  deps: ChannelDeps,
  signal: AbortSignal,
): Promise<PartETag[]> {
  const total = file.size;
  const partCount = Math.ceil(total / partSize);
  const partBytes = (n: number) => (n === partCount ? total - (n - 1) * partSize : partSize);
  const etags = new Map(uploadedETags.map((p) => [p.partNumber, p.etag]));
  let completed = [...etags.keys()].reduce((sum, n) => sum + partBytes(n), 0);
  const live = new Map<number, number>();
  const bump = () => {
    let loaded = completed;
    for (const v of live.values()) {
      loaded += v;
    }
    deps.setProgress(id, { loaded, total });
  };
  deps.setProgress(id, { loaded: completed, total });

  let expired = false;
  let firstErr: unknown;
  let firstErrSet = false;
  const queue = partUrls.filter((p) => !etags.has(p.partNumber));
  let cursor = 0;
  // One part's outcome: undefined = ok, "expired" = stop the pool and
  // resync, anything else = the hard error value.
  const runOne = async (item: UploadPartUrl): Promise<undefined | "expired" | unknown> => {
    const n = item.partNumber;
    const blob = file.slice((n - 1) * partSize, n * partSize);
    live.set(n, 0);
    try {
      const { etag } = await putPart(blob, item, signal);
      etags.set(n, etag);
      live.delete(n);
      completed += partBytes(n);
      bump();
      return undefined;
    } catch (err) {
      // Parts finishing after the flag are re-read from server truth at
      // resync, so losing their local etags here costs nothing.
      live.delete(n);
      return partFailure(err, allowResync) === "expired" ? "expired" : err;
    }
  };
  const worker = async (): Promise<void> => {
    while (cursor < queue.length && !expired) {
      const item = queue[cursor];
      cursor += 1;
      if (!item) {
        continue;
      }
      const outcome = await runOne(item);
      if (outcome === "expired") {
        expired = true;
        return;
      }
      if (outcome === undefined) {
        continue;
      }
      if (!firstErrSet) {
        firstErr = outcome;
        firstErrSet = true;
      }
      return;
    }
  };
  await Promise.all(Array.from({ length: Math.min(PART_CONCURRENCY, queue.length) }, worker));
  if (firstErrSet) {
    throw firstErr;
  }
  if (expired) {
    const info = await deps.listParts({ id });
    return driveParts(
      file,
      id,
      partSize,
      info.missing ?? [],
      mergeETags(uploadedETags, info.uploaded),
      false,
      deps,
      signal,
    );
  }
  return [...etags.entries()]
    .map(([partNumber, etag]) => ({ partNumber, etag }))
    .sort((a, b) => a.partNumber - b.partNumber);
}

async function putPart(blob: Blob, item: UploadPartUrl, signal: AbortSignal): Promise<PartETag> {
  let lastErr: unknown;
  for (let attempt = 0; attempt < PART_ATTEMPTS; attempt++) {
    try {
      const { etag } = await putUrl(item.url ?? "", blob, { signal });
      return { partNumber: item.partNumber, etag };
    } catch (err) {
      // A cancelled drive must not spend its network retry on a session
      // the abort RPC is already tearing down.
      if (signal.aborted) {
        throw new UploadCancelledError();
      }
      if (err instanceof S3Error && err.kind === "network") {
        lastErr = err;
        continue;
      }
      throw err;
    }
  }
  throw lastErr;
}

function partFailure(err: unknown, allowResync: boolean): "expired" | "hard" {
  if (err instanceof S3Error && err.kind === "expired" && allowResync) {
    return "expired";
  }
  return "hard";
}

function contentTypeOf(file: File): string {
  return file.type === "" ? "application/octet-stream" : file.type;
}

function toPartETag(p: UploadedPart): PartETag {
  return { partNumber: p.partNumber, etag: p.etag ?? "" };
}

function mergeETags(local: PartETag[], server: UploadedPart[]): PartETag[] {
  const merged = new Map(local.map((p) => [p.partNumber, p.etag]));
  for (const p of server) {
    merged.set(p.partNumber, p.etag ?? "");
  }
  return [...merged].map(([partNumber, etag]) => ({ partNumber, etag }));
}

// RPC-leg failures are already surfaced by the mutation cache's error
// policy; byte-leg (S3) failures have no such outlet and get the generic
// upload-failure copy here. Anything else is an orchestration bug — logged
// so it never dies silently.
function toastS3Failure(err: unknown) {
  if (err instanceof S3Error) {
    toast.error(m.files_upload_failed());
    return;
  }
  console.error("[upload] orchestration failure", err);
}
