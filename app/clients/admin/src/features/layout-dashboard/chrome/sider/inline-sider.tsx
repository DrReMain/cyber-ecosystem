import { theme as antdTheme } from "antd";
import { SiderChrome } from "./sider-chrome";

interface InlineSiderProps {
  collapsed: boolean;
  onToggleCollapse: () => void;
}

export function InlineSider({ collapsed, onToggleCollapse }: Readonly<InlineSiderProps>) {
  const { token } = antdTheme.useToken();

  return (
    <div className="flex h-full flex-col" style={{ backgroundColor: token.colorBgContainer }}>
      <SiderChrome collapsed={collapsed} onToggleCollapse={onToggleCollapse} />
    </div>
  );
}
