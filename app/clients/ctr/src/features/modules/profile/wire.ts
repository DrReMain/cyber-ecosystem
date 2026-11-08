import { ProfileSex } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";

// The read-side wire names (proto-JSON) for the sex enum. Transport
// responses re-serialize through toJson, so a sex on the wire is the
// enum's value-name string, never the number the generated types claim.
// The write side takes the enum itself, so this mapping lives only where
// reads land.
export const SEX_FROM_WIRE: Record<string, ProfileSex> = {
  PROFILE_SEX_MALE: ProfileSex.MALE,
  PROFILE_SEX_FEMALE: ProfileSex.FEMALE,
  PROFILE_SEX_UNDETERMINED: ProfileSex.UNDETERMINED,
};

export function sexFromWire(raw: ProfileSex | string | undefined): ProfileSex | undefined {
  return typeof raw === "string" ? SEX_FROM_WIRE[raw] : undefined;
}
