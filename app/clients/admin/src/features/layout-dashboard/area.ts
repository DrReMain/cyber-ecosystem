import { AREAS } from "#/config";

export const AREA_NAMESPACE = "dashboard";
export const AREA_PATH = AREAS[AREA_NAMESPACE].home;
export const LOGIN_PATH = AREAS[AREA_NAMESPACE].login;
export const LOGIN_ROUTE_ID: `${typeof LOGIN_PATH}/` = `${LOGIN_PATH}/`;
