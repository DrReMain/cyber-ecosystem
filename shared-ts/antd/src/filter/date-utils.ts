import type { Dayjs } from "dayjs";
import dayjs from "dayjs";

export const dayjsToProps = (v?: number) => ({
  value: v != null ? dayjs(v) : undefined,
});
export const dayjsFromEvent = (v: Dayjs | null) => v?.valueOf();

export type RangeTuple = [number | null, number | null] | null | undefined;

export const rangeTupleToProps = (v?: RangeTuple) => ({
  value: v
    ? ([v[0] != null ? dayjs(v[0]) : null, v[1] != null ? dayjs(v[1]) : null] as [
        Dayjs | null,
        Dayjs | null,
      ])
    : null,
});
export const rangeTupleFromEvent = (dates: [Dayjs | null, Dayjs | null] | null): RangeTuple =>
  dates
    ? [dates[0] != null ? dates[0].valueOf() : null, dates[1] != null ? dates[1].valueOf() : null]
    : [null, null];
