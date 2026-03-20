import { useSuspenseQuery } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";

import { authQueries } from "@/features/auth/queries";
import { workspaceQueries } from "@/features/workspace/queries";

const workspaceSlugApi = getRouteApi("/_workspace/$slug");

/** Session context: { user, workspaces } */
export function useSession() {
  const { data: user } = useSuspenseQuery(authQueries.session());
  const { data: workspaces } = useSuspenseQuery(workspaceQueries.list());
  return { user, workspaces };
}

/** Workspace: { workspace, user, workspaces } */
export function useWorkspace() {
  const session = useSession();
  const { workspace } = workspaceSlugApi.useRouteContext();

  return { ...session, workspace };
}
