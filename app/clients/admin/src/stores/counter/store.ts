import { defineStore } from "@cyber-ecosystem/shared-store";

export const counterStore = defineStore("store_counter", 0, {
  debugLabel: "Counter",
});

export const counterAtom = counterStore.atom;
