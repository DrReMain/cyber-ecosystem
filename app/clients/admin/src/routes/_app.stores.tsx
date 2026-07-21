import { createFileRoute } from "@tanstack/react-router";
import { useAtom, useAtomValue } from "jotai";
import { useState } from "react";
import { counterAtom } from "#/stores/counter/store";
import {
  type TodoItem,
  todoCountAtom,
  todoDoneCountAtom,
  todoListAtom,
} from "#/stores/todolist/store";

export const Route = createFileRoute("/_app/stores")({
  component: StoresDemo,
});

function StoresDemo() {
  const [count, setCount] = useAtom(counterAtom);
  const [todos, setTodos] = useAtom(todoListAtom);
  const total = useAtomValue(todoCountAtom);
  const done = useAtomValue(todoDoneCountAtom);
  const [text, setText] = useState("");

  function addTodo() {
    const t = text.trim();
    if (!t) return;
    setTodos((draft) => {
      draft.push({ id: Date.now(), text: t, done: false });
    });
    setText("");
  }

  function toggleTodo(id: number) {
    setTodos((draft) => {
      const item = draft.find((d) => d.id === id);
      if (item) item.done = !item.done;
    });
  }

  return (
    <div className="space-y-8">
      <section className="space-y-3">
        <div className="flex items-center gap-3">
          <button
            className="rounded border px-3 py-1"
            onClick={() => setCount(count - 1)}
            type="button"
          >
            -
          </button>
          <span className="w-8 text-center tabular-nums">{count}</span>
          <button
            className="rounded border px-3 py-1"
            onClick={() => setCount(count + 1)}
            type="button"
          >
            +
          </button>
        </div>
      </section>

      <section className="space-y-3">
        <div className="flex gap-2">
          <input
            className="rounded border px-2 py-1"
            onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") addTodo();
            }}
            value={text}
          />
          <button className="rounded border px-3 py-1" onClick={addTodo} type="button">
            添加
          </button>
        </div>
        <p className="text-sm opacity-70">
          总计 {total} / 完成 {done}
        </p>
        <ul className="space-y-1">
          {todos.map((t: TodoItem) => (
            <li className="flex items-center gap-2" key={t.id}>
              <input checked={t.done} onChange={() => toggleTodo(t.id)} type="checkbox" />
              <span className={t.done ? "line-through opacity-50" : ""}>{t.text}</span>
            </li>
          ))}
        </ul>
      </section>
    </div>
  );
}
