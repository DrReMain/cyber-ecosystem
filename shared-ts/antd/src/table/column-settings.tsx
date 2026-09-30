import { Button, Checkbox, Popover, Space, Tooltip } from "antd";
import { Columns3, RotateCcw } from "lucide-react";
import { useCallback, useMemo, useSyncExternalStore } from "react";
import { storageGet, storageRemove, storageSet } from "../storage";

export interface ColumnSettingsLabels {
  title?: string;
  reset?: string;
}

export interface ColumnSettingsConfig {
  storageKey: string;
  locked?: string[];
  labels?: ColumnSettingsLabels;
}

// Column visibility per table, persisted as the HIDDEN id list — new columns
// (ids absent from storage) default to visible, so adding a column never
// ships hidden for returning users. Storage is the single source of truth;
// the module map holds only cross-instance listeners.
const listenersByKey = new Map<string, Set<() => void>>();

function notify(key: string): void {
  const listeners = listenersByKey.get(key);
  if (!listeners) return;
  for (const notifyOne of listeners) notifyOne();
}

function readHidden(key: string): string[] {
  const raw = storageGet(key);
  if (raw === null) return [];
  try {
    const parsed: unknown = JSON.parse(raw);
    if (Array.isArray(parsed) && parsed.every((v) => typeof v === "string")) {
      return parsed;
    }
  } catch {
    // corrupt payload: treat as nothing hidden
  }
  return [];
}

export function setColumnHidden(key: string, hidden: string[]): void {
  storageSet(key, JSON.stringify(hidden));
  notify(key);
}

export function resetColumnVisibility(key: string): void {
  storageRemove(key);
  notify(key);
}

const EMPTY: string[] = [];
const hiddenByKey = new Map<string, string[]>();

function hiddenSnapshot(key: string): string[] {
  const parsed = readHidden(key);
  const cached = hiddenByKey.get(key);
  if (
    cached !== undefined &&
    cached.length === parsed.length &&
    cached.every((id, i) => id === parsed[i])
  ) {
    return cached;
  }
  hiddenByKey.set(key, parsed);
  return parsed;
}

export function useColumnVisibility(storageKey: string | undefined, allIds: string[]): string[] {
  const subscribe = useCallback(
    (onStoreChange: () => void) => {
      if (storageKey === undefined) return () => {};
      let listeners = listenersByKey.get(storageKey);
      if (!listeners) {
        listeners = new Set();
        listenersByKey.set(storageKey, listeners);
      }
      listeners.add(onStoreChange);
      return () => {
        listeners.delete(onStoreChange);
        if (listeners.size === 0) listenersByKey.delete(storageKey);
      };
    },
    [storageKey],
  );
  const getSnapshot = useCallback(
    () => (storageKey === undefined ? EMPTY : hiddenSnapshot(storageKey)),
    [storageKey],
  );
  const hidden = useSyncExternalStore(subscribe, getSnapshot, () => EMPTY);
  return useMemo(
    () => (hidden.length === 0 ? allIds : allIds.filter((id) => !hidden.includes(id))),
    [allIds, hidden],
  );
}

export interface ColumnSettingsEntry {
  id: string;
  title: React.ReactNode;
  locked: boolean;
}

export function ColumnSettingsPopover({
  entries,
  visibleIds,
  config,
}: {
  entries: ColumnSettingsEntry[];
  visibleIds: string[];
  config: ColumnSettingsConfig;
}) {
  const { title = "Columns", reset = "Reset" } = config.labels ?? {};
  const options = entries.map((e) => ({
    label: e.title,
    value: e.id,
    disabled: e.locked,
  }));
  return (
    <Popover
      placement="left"
      content={
        <Space orientation="vertical" size={4}>
          <Checkbox.Group
            className="flex flex-col gap-1"
            onChange={(checked) => {
              const visible = checked as string[];
              const hidden = entries.map((e) => e.id).filter((id) => !visible.includes(id));
              setColumnHidden(config.storageKey, hidden);
            }}
            options={options}
            value={visibleIds}
          />
          <Button
            block
            icon={<RotateCcw size={14} />}
            onClick={() => resetColumnVisibility(config.storageKey)}
            size="small"
            variant="filled"
            color="default"
          >
            {reset}
          </Button>
        </Space>
      }
      trigger="click"
    >
      <Tooltip title={title}>
        <Button aria-label={title} color="default" icon={<Columns3 size={14} />} variant="filled" />
      </Tooltip>
    </Popover>
  );
}
