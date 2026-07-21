import { createServerFn } from "@tanstack/react-start";
import { getCookie } from "@tanstack/react-start/server";
import { readCookie } from "#/lib/cookie";
import { getRegistry } from "./define-store";

type Serializable =
  | null
  | boolean
  | number
  | string
  | Serializable[]
  | { [key: string]: Serializable };

export type StoreCookieData = Record<string, unknown>;

export const getStoreCookies = createServerFn({ method: "GET" }).handler(async () => {
  const data: StoreCookieData = {};
  for (const store of getRegistry()) {
    if (!store.persist) continue;
    data[store.key] = readCookie(getCookie(store.key), store.schema, store.initial);
  }
  return data as Record<string, Serializable>;
});
