import { createFileRoute } from "@tanstack/react-router"
import {
  Alert,
  Button,
  Card,
  Checkbox,
  Flex,
  Input,
  Segmented,
  Select,
  Switch,
  Tag,
} from "antd"
import { useAtom, useAtomValue } from "jotai"
import { useState } from "react"
import { getSkinIds } from "#/domains/antd/skins"
import { ErrorBoundary } from "#/domains/errors"
import { useTheme } from "#/domains/theme"
import { getLocale } from "#/paraglide/runtime"
import { counterAtom } from "#/stores/counter/store"
import {
  todoCountAtom,
  todoDoneCountAtom,
  todoListAtom,
} from "#/stores/todolist/store"

export const Route = createFileRoute("/_app/")({
  component: Home,
})

function Home() {
  const { skinId, mode, preference, setMode, setSkinId, compact, setCompact } =
    useTheme()

  return (
    <section className="relative flex min-h-screen flex-col items-center gap-8 overflow-hidden p-8">
      <div className="pointer-events-none absolute inset-0 -z-10 animate-hero-gradient opacity-60 dark:opacity-40" />

      <div className="flex flex-wrap items-center justify-center gap-2 rounded-xl bg-antd-fill/80 p-4 backdrop-blur-md">
        <Select
          onChange={setSkinId}
          options={getSkinIds().map((s) => ({ value: s, label: s }))}
          style={{ width: 200 }}
          value={skinId}
        />
        <Segmented
          onChange={(v) => setMode(v as "light" | "dark" | "system")}
          options={[
            { label: "Light", value: "light" },
            { label: "Dark", value: "dark" },
            { label: "System", value: "system" },
          ]}
          value={mode}
        />
        <Switch
          checked={compact}
          checkedChildren="Compact"
          onChange={(c) => setCompact(c)}
          unCheckedChildren="Normal"
        />
        <span className="text-antd-text-tertiary text-sm">
          {getLocale()} · {preference}({mode}){compact ? " · compact" : ""}
        </span>
      </div>

      <div className="w-full max-w-2xl space-y-4">
        <CounterDemo />
        <TodoListDemo />
        <ErrorBoundary
          fallback={<Alert title="Component error caught" type="error" />}
        >
          <ErrorDemo />
        </ErrorBoundary>
      </div>

      <style>{`
        @keyframes hero-gradient {
          0%, 100% {
            background:
              radial-gradient(ellipse at 20% 50%, rgba(59,130,246,0.3) 0%, transparent 50%),
              radial-gradient(ellipse at 80% 20%, rgba(168,85,247,0.25) 0%, transparent 50%),
              radial-gradient(ellipse at 60% 80%, rgba(34,197,94,0.2) 0%, transparent 50%);
          }
          33% {
            background:
              radial-gradient(ellipse at 60% 30%, rgba(59,130,246,0.3) 0%, transparent 50%),
              radial-gradient(ellipse at 30% 70%, rgba(168,85,247,0.25)  0%, transparent 50%),
              radial-gradient(ellipse at 80% 60%, rgba(34,197,94,0.2) 0%, transparent 50%);
          }
          66% {
            background:
              radial-gradient(ellipse at 40% 70%, rgba(59,130,246,0.3) 0%, transparent 50%),
              radial-gradient(ellipse at 70% 40%, rgba(168,85,247,0.25) 0%, transparent 50%),
              radial-gradient(ellipse at 20% 30%, rgba(34,197,94,0.2) 0%, transparent 50%);
          }
        }
        .animate-hero-gradient {
          animation: hero-gradient 12s ease-in-out infinite;
        }
      `}</style>
    </section>
  )
}

function CounterDemo() {
  const [count, setCount] = useAtom(counterAtom)

  return (
    <Card size="small" title="Counter (client-only)">
      <Flex align="center" gap="small">
        <Button
          color="default"
          onClick={() => setCount((c) => c - 1)}
          variant="filled"
        >
          -
        </Button>
        <Tag color="blue">{count}</Tag>
        <Button
          color="default"
          onClick={() => setCount((c) => c + 1)}
          variant="filled"
        >
          +
        </Button>
        <Button color="default" onClick={() => setCount(0)} variant="solid">
          Reset
        </Button>
        <span className="ml-2 text-antd-text-tertiary text-xs">
          not persisted
        </span>
      </Flex>
    </Card>
  )
}

function TodoListDemo() {
  const [todos, setTodos] = useAtom(todoListAtom)
  const total = useAtomValue(todoCountAtom)
  const done = useAtomValue(todoDoneCountAtom)

  function addTodo() {
    setTodos((draft) => {
      draft.push({
        id: Date.now(),
        text: `Task ${draft.length + 1}`,
        done: false,
      })
    })
  }

  function toggleTodo(id: number) {
    setTodos((draft) => {
      const item = draft.find((t) => t.id === id)
      if (item) item.done = !item.done
    })
  }

  function removeTodo(id: number) {
    setTodos((draft) => {
      const idx = draft.findIndex((t) => t.id === id)
      if (idx !== -1) draft.splice(idx, 1)
    })
  }

  return (
    <Card
      size="small"
      title={`TodoList (cookie-persisted) — ${done}/${total} done`}
    >
      <Flex align="center" className="mb-3" gap="small">
        <Button color="primary" onClick={addTodo} variant="filled">
          Add Todo
        </Button>
        <span className="text-antd-text-tertiary text-xs">
          refresh page → data persists · uses immer draft
        </span>
      </Flex>
      <Flex gap="small" vertical>
        {todos.map((todo) => (
          <Flex align="center" gap="small" key={todo.id}>
            <Checkbox
              checked={todo.done}
              onChange={() => toggleTodo(todo.id)}
            />
            <span
              className={`flex-1 text-sm ${todo.done ? "text-antd-text-tertiary line-through" : "text-antd-text"}`}
            >
              {todo.text}
            </span>
            <Button
              color="danger"
              danger
              onClick={() => removeTodo(todo.id)}
              size="small"
              variant="text"
            >
              del
            </Button>
          </Flex>
        ))}
        {todos.length === 0 && (
          <p className="text-antd-text-tertiary text-sm">
            No items. Click "Add Todo".
          </p>
        )}
      </Flex>
    </Card>
  )
}

function ErrorDemo() {
  const [value, setValue] = useState("{}")
  const [json, setJson] = useState(value)
  return (
    <Flex gap={8} vertical>
      <Button
        block
        color="danger"
        onClick={() => setJson(value)}
        type="primary"
        variant="solid"
      >
        stringify
      </Button>
      <Input onChange={(e) => setValue(e.target.value)} value={value} />
      <pre>{JSON.stringify(JSON.parse(json), null, 2)}</pre>
    </Flex>
  )
}
