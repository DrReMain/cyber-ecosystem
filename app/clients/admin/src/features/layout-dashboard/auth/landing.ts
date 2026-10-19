import { routeTree } from "#/routeTree.gen";
import { AREA_PATH } from "../area";
import { buildNavNodes, firstLeafOf, type NavRoute } from "../protocol/nav";

export function defaultLanding(permissions: readonly string[]): string {
  return (
    firstLeafOf(buildNavNodes(routeTree as unknown as NavRoute, permissions))?.path ?? AREA_PATH
  );
}
