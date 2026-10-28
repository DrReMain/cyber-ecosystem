import { createQueryOptions } from "@connectrpc/connect-query";
import { getCareerReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { connectTransport } from "#/services/connect";
import type { WireAccessVerdict } from "./cognitive";

// The career aggregate's client seam. The seeded library's own facts
// (elimination rate, training load wording, advice copy) travel on each
// recommendation row — library copy is data, never page wording — so the
// page renders whatever the rows carry and invents nothing.

export function careerReportKey() {
  return createQueryOptions(getCareerReport, {}, { transport: connectTransport }).queryKey;
}

export interface WireCareerRecommendation {
  careerArea: string;
  name: string;
  matchedScore: number;
  basisDimensions: string[];
  eliminationRate: number;
  trainingIntensity: string;
  suggestion: string;
}

export interface WireCareerImprovement {
  dimension: string;
  current: number;
  target: number;
  gapPercent: number;
  suggestions: string[];
}

export interface WireCareer {
  access: WireAccessVerdict;
  recommendations: WireCareerRecommendation[];
  improvements: WireCareerImprovement[];
}

export function wireCareerOf(data: unknown): WireCareer {
  if (data === null || data === undefined) {
    return {
      access: { allowed: false, reason: undefined },
      recommendations: [],
      improvements: [],
    };
  }
  const d = data as Partial<WireCareer> & { access?: Partial<WireAccessVerdict> };
  // Row fields narrow per field: protojson drops a zero scalar entirely
  // (an at-the-line gap of 0 arrives fieldless), so an absent number is
  // the wire's own zero, never an undefined the template would print.
  return {
    access: {
      allowed: d.access?.allowed === true,
      reason: d.access?.reason,
    },
    recommendations: (d.recommendations ?? []).map((rec) => ({
      ...rec,
      matchedScore: rec.matchedScore ?? 0,
      eliminationRate: rec.eliminationRate ?? 0,
      basisDimensions: rec.basisDimensions ?? [],
    })),
    improvements: (d.improvements ?? []).map((imp) => ({
      ...imp,
      current: imp.current ?? 0,
      target: imp.target ?? 0,
      gapPercent: imp.gapPercent ?? 0,
      suggestions: imp.suggestions ?? [],
    })),
  };
}
