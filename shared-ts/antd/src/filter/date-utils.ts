import type { Dayjs } from "dayjs"
import dayjs from "dayjs"

export const dayjsToProps = (v?: number) => ({
  value: v != null ? dayjs(v) : undefined,
})
export const dayjsFromEvent = (v: Dayjs | null) => v?.valueOf()
