import { atom } from "jotai"
import { z } from "zod"
import { defineStore } from "#/stores/_core/define-store"

export interface TodoItem {
  id: number
  text: string
  done: boolean
}

const todoListSchema = z.array(
  z.object({
    id: z.number(),
    text: z.string(),
    done: z.boolean(),
  }),
)

export const todoListStore = defineStore<TodoItem[]>("store_todolist", [], {
  persist: true,
  debugLabel: "TodoList",
  schema: todoListSchema,
})

export const todoListAtom = todoListStore.atom

export const todoCountAtom = atom((get) => get(todoListStore.atom).length)
export const todoDoneCountAtom = atom(
  (get) => get(todoListStore.atom).filter((t) => t.done).length,
)
