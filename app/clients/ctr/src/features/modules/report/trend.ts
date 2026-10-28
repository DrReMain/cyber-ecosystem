import { createQueryOptions } from "@connectrpc/connect-query";
import { getTrendReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { connectTransport } from "#/services/connect";
import type { WireAccessVerdict } from "./cognitive";

// The trend aggregate's client seam. Each point carries its genre on the
// wire (the mixed FPS/MOBA contract — one history, two vocabularies), and
// the advanced verdict rides unconditionally so an empty page still
// renders its locked state from one call.

export function trendReportKey() {
  return createQueryOptions(getTrendReport, {}, { transport: connectTransport }).queryKey;
}

export interface WireTrendPoint {
  at: string;
  percentile: number;
  genre: string;
}

export interface WireTrendSeries {
  ability: string;
  points: WireTrendPoint[];
}

export interface WireTrend {
  sessionsTotal: number;
  abilities: WireTrendSeries[];
  advancedAccess: WireAccessVerdict;
}

export function wireTrendOf(data: unknown): WireTrend {
  if (data === null || data === undefined) {
    return {
      sessionsTotal: 0,
      abilities: [],
      advancedAccess: { allowed: false, reason: undefined },
    };
  }
  const d = data as Partial<WireTrend> & { advancedAccess?: Partial<WireAccessVerdict> };
  return {
    sessionsTotal: d.sessionsTotal ?? 0,
    // Row fields narrow per field: protojson drops a zero scalar — a
    // floor percentile of 0 still plots its point.
    abilities: (d.abilities ?? []).map((s) => ({
      ...s,
      points: (s.points ?? []).map((p) => ({ ...p, percentile: p.percentile ?? 0 })),
    })),
    advancedAccess: {
      allowed: d.advancedAccess?.allowed === true,
      reason: d.advancedAccess?.reason,
    },
  };
}
