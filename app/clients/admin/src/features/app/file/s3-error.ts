export type S3ErrorKind = "expired" | "tooLarge" | "network" | "unknown";

// S3 answers presigned-URL misuse with XML error documents, not our error
// domain — this is the single translation point for the byte leg. The kind
// drives recovery: "expired" re-mints, "network" retries the part, the rest
// surface as upload failure.
export class S3Error extends Error {
  readonly kind: S3ErrorKind;
  readonly s3Code: string;

  constructor(
    body: string | undefined,
    readonly status: number,
  ) {
    const code = parseCode(body);
    super(code ?? (status === 0 ? "network" : `status ${status}`));
    this.s3Code = code ?? "";
    this.kind = classify(code, status);
  }
}

function parseCode(body: string | undefined): string | undefined {
  if (body === undefined || body === "") {
    return undefined;
  }
  const doc = new DOMParser().parseFromString(body, "application/xml");
  return doc.querySelector("Error > Code")?.textContent ?? undefined;
}

function classify(code: string | undefined, status: number): S3ErrorKind {
  if (status === 0) {
    return "network";
  }
  // Expired presigns surface as either code depending on the backend.
  if (code === "ExpiredToken" || code === "AccessDenied") {
    return "expired";
  }
  if (code === "EntityTooLarge") {
    return "tooLarge";
  }
  return "unknown";
}
