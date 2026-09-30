import type { ErrorEnvelope } from "./envelope";
import type { ErrorKind } from "./kinds";

/** Recognizes a source-specific shape; return undefined when not yours. */
export interface SourceAdapter {
  readonly source: string;
  recognize(errorLike: unknown): ErrorEnvelope | undefined;
}

/** UI notification; receives the presentation answer, not the raw facts. */
export type FeedbackSink = (view: ErrorView) => void;

/** Telemetry; receives the raw facts. A sink must never throw back. */
export type ReporterSink = (envelope: ErrorEnvelope) => void;

/** The single copy seam: wording lives here, never in adapters. */
export interface CopyResolver {
  title(kind: ErrorKind): string;
  message(envelope: ErrorEnvelope): string;
}

/**
 * Per-callsite policy; integration defaults fill the gaps, silence wins.
 * A type alias, not an interface: consumers register it as query meta, and
 * meta registration requires an implicit index signature interfaces lack.
 */
export type ErrorPolicy = {
  feedback?: boolean;
  report?: boolean;
};

/** Presentation answer for renderers — everything a dumb view needs. */
export interface ErrorView {
  readonly kind: ErrorKind;
  readonly silent: boolean;
  readonly http: number;
  readonly title: string;
  readonly message: string;
  readonly reason?: string;
}
