import { S3Error } from "./s3-error";

export interface PutUrlOptions {
  contentType?: string;
  onProgress?: (loaded: number) => void;
  signal?: AbortSignal;
}

export interface PutUrlResult {
  etag: string;
}

// XHR rather than fetch: upload progress rides xhr.upload, and the ETag
// response header (multipart confirm needs it) is only readable when the
// bucket's CORS policy exposes it.
export function putUrl(url: string, body: Blob, opts: PutUrlOptions = {}): Promise<PutUrlResult> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    xhr.open("PUT", url);
    if (opts.contentType) {
      // The presign does not sign Content-Type; the backend stores the
      // header verbatim, so the object keeps whatever we declare here.
      xhr.setRequestHeader("Content-Type", opts.contentType);
    }
    xhr.upload.onprogress = (e) => opts.onProgress?.(e.loaded);
    // Aborting the signal stops the in-flight request — the cancel path for
    // an upload being discarded mid-transfer. An already-aborted signal
    // never fires the listener, so it is checked directly instead.
    if (opts.signal?.aborted) {
      reject(new S3Error(undefined, 0));
      return;
    }
    opts.signal?.addEventListener("abort", () => xhr.abort(), { once: true });
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve({ etag: xhr.getResponseHeader("etag") ?? "" });
      } else {
        reject(new S3Error(xhr.responseText, xhr.status));
      }
    };
    // Every terminal XHR path must settle the promise — an aborted request
    // would otherwise hang its caller forever.
    xhr.onabort = () => reject(new S3Error(undefined, 0));
    xhr.onerror = () => reject(new S3Error(undefined, 0));
    xhr.send(body);
  });
}
