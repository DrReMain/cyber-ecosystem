import { createQueryOptions } from "@connectrpc/connect-query";
import { getCognitiveReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { connectTransport } from "#/services/connect";

// The cognitive aggregate's client seam: one query, one wire narrowing.
// The verdict rides inside the response (the embedded-verdict contract),
// so the page renders the locked face from the same call — no second
// probe, and the numbers behind the gate never arrive at all.

export function cognitiveReportKey() {
  return createQueryOptions(getCognitiveReport, {}, { transport: connectTransport }).queryKey;
}

/** The wire flattens the WKT DoubleValue composite to a scalar and the
 * enums to value-name strings; an empty repeated field's key drops
 * entirely under protojson, so absent rows stay a promised empty list. */
export interface WireAccessVerdict {
  allowed: boolean;
  reason: string | undefined;
}

export interface WireCognitiveDimension {
  dimension: string;
  value: number;
  tier: string | undefined;
}

export interface WireCognitive {
  access: WireAccessVerdict;
  coveredAbilities: number;
  testedDimensions: number;
  totalDimensions: number;
  dimensions: WireCognitiveDimension[];
  composite: number | undefined;
  tier: string | undefined;
}

export function wireCognitiveOf(data: unknown): WireCognitive {
  if (data === null || data === undefined) {
    return {
      access: { allowed: false, reason: undefined },
      coveredAbilities: 0,
      testedDimensions: 0,
      totalDimensions: 8,
      dimensions: [],
      composite: undefined,
      tier: undefined,
    };
  }
  const d = data as Partial<WireCognitive> & { access?: Partial<WireAccessVerdict> };
  return {
    access: {
      allowed: d.access?.allowed === true,
      reason: d.access?.reason,
    },
    coveredAbilities: d.coveredAbilities ?? 0,
    testedDimensions: d.testedDimensions ?? 0,
    totalDimensions: d.totalDimensions ?? 8,
    dimensions: d.dimensions ?? [],
    composite: d.composite,
    tier: d.tier,
  };
}
