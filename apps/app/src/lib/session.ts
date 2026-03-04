import type { QueryClient } from "@tanstack/react-query";
import { isRedirect, redirect } from "@tanstack/react-router";
import { authQueries } from "@/features/auth/queries";
import type { User } from "@/features/auth/types";
import { workspaceQueries } from "@/features/workspace/queries";
import type { WorkspaceWithRole } from "@/features/workspace/types";

export interface SessionContext {
  user: User;
  workspaces: WorkspaceWithRole[];
}

/**
 * Ensures the user has a valid session. Redirects to /login on failure.
 * Called once in `_workspace.tsx` — the single auth checkpoint.
 */
export async function ensureSession(
  queryClient: QueryClient,
  redirectTo?: string,
): Promise<SessionContext> {
  try {
    const [user, workspaces] = await Promise.all([
      queryClient.ensureQueryData(authQueries.session()),
      queryClient.ensureQueryData(workspaceQueries.list()),
    ]);
    return { user, workspaces };
  } catch (error) {
    if (isRedirect(error)) throw error;
    throw redirect({
      to: "/login",
      search: redirectTo ? { redirect: redirectTo } : undefined,
    });
  }
}

const LAST_WORKSPACE_KEY = "last_workspace";

/**
 * Resolves which workspace slug to navigate to.
 * Priority: localStorage hint → first available → undefined (onboarding).
 *
 * localStorage is the right tool — this is a per-device UI preference.
 * Stale entries (deleted workspace, removed membership) auto-clean.
 */
export function resolveDefaultWorkspace(
  workspaces: WorkspaceWithRole[],
): string | undefined {
  const stored = localStorage.getItem(LAST_WORKSPACE_KEY);
  if (stored) {
    const match = workspaces.find((w) => w.slug === stored);
    if (match) return match.slug;
    localStorage.removeItem(LAST_WORKSPACE_KEY);
  }
  return workspaces[0]?.slug;
}

/** Persist the current workspace as a per-device preference. */
export function setLastWorkspace(slug: string): void {
  localStorage.setItem(LAST_WORKSPACE_KEY, slug);
}
