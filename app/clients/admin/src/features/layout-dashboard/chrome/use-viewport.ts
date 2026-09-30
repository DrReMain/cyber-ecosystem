import { Grid } from "antd";

export function useIsNarrowViewport(): boolean {
  const screens = Grid.useBreakpoint();
  return screens.lg === false;
}
