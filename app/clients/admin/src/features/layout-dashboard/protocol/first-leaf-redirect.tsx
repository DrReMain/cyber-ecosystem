import { Navigate } from "@tanstack/react-router";
import { useMemo } from "react";
import { AREA_PATH } from "../area";
import { firstLeafOf, useNavIndex } from "./nav";

export function FirstLeafRedirect({ of }: Readonly<{ of: string }>) {
  const navIndex = useNavIndex();
  const target = useMemo(() => {
    const node = navIndex.get(of);
    return node ? (firstLeafOf([node])?.path ?? AREA_PATH) : AREA_PATH;
  }, [navIndex, of]);
  return <Navigate replace to={target as never} />;
}

export function firstLeafRedirect(of: string) {
  // biome-ignore lint/nursery/noComponentHookFactories: route options evaluate once at module scope, so the component identity is stable for the app lifetime.
  return function GroupIndexRedirect() {
    return <FirstLeafRedirect of={of} />;
  };
}
