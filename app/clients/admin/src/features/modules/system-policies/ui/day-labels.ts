import { m } from "#/paraglide/messages";

// Index = Go time.Weekday (0 = Sunday), matching the wire `days` field.
export const dayLabels = (): string[] => [
  m.system_policies_day_0(),
  m.system_policies_day_1(),
  m.system_policies_day_2(),
  m.system_policies_day_3(),
  m.system_policies_day_4(),
  m.system_policies_day_5(),
  m.system_policies_day_6(),
];
