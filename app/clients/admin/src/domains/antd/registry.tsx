import { createCache, StyleProvider } from "@ant-design/cssinjs";
import { type PropsWithChildren, useState } from "react";
import { StyleCollector } from "./style-collector";

export function AntdRegistry({ children }: Readonly<PropsWithChildren>) {
  const [cache] = useState(() => createCache());
  return (
    <StyleProvider cache={cache} layer>
      {children}
      <StyleCollector cache={cache} />
    </StyleProvider>
  );
}
