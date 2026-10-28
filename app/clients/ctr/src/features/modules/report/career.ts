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
  return {
    access: {
      allowed: d.access?.allowed === true,
      reason: d.access?.reason,
    },
    recommendations: d.recommendations ?? [],
    improvements: d.improvements ?? [],
  };
}
