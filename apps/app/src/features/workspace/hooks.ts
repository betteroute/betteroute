import { getRouteApi } from "@tanstack/react-router";

const workspaceApi = getRouteApi("/_workspace");
const workspaceSlugApi = getRouteApi("/_workspace/$slug");

/** Session context: { user, workspaces } from the _workspace layout guard. */
export function useSession() {
  return workspaceApi.useRouteContext();
}

/** Resolved workspace: { workspace } + inherited session context. */
export function useWorkspace() {
  return workspaceSlugApi.useRouteContext();
}
