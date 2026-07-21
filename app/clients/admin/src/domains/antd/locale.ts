import dayjs from "dayjs";
import "dayjs/locale/ar-sa";
import "dayjs/locale/zh-cn";
import arEG from "antd/locale/ar_EG";
import enUS from "antd/locale/en_US";
import zhCN from "antd/locale/zh_CN";
import { useEffect } from "react";
import { getLocale, getTextDirection } from "#/paraglide/runtime";

interface LocaleEntry {
  antd: typeof enUS;
  dayjs: string;
}

const defaultEntry: LocaleEntry = { antd: enUS, dayjs: "en" };
const registry: Record<string, LocaleEntry> = {
  "en-US": { antd: enUS, dayjs: "en" },
  "zh-CN": { antd: zhCN, dayjs: "zh-cn" },
  "ar-SA": { antd: arEG, dayjs: "ar-sa" },
};

export function useAntdLocale() {
  const locale = getLocale();
  const entry = registry[locale] ?? defaultEntry;
  const direction = getTextDirection(locale);

  useEffect(() => {
    dayjs.locale(entry.dayjs);
  }, [entry.dayjs]);

  return { locale: entry.antd, direction };
}
