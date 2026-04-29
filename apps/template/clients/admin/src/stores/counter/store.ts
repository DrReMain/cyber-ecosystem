import { defineStore } from "#/stores/_core/define-store"

export const counterStore = defineStore("store_counter", 0, {
  debugLabel: "Counter",
})

export const counterAtom = counterStore.atom
