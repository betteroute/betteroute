import { queryOptions } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { QUERY_CACHE } from "@/lib/constants";
import type { CreateInput, UpdateInput } from "./schemas";
import type { Invitation, Member, WorkspaceWithRole } from "./types";

export const workspaceKeys = {
  all: ["workspaces"] as const,
  list: () => [...workspaceKeys.all, "list"] as const,
  detail: (slug: string) => [...workspaceKeys.all, slug] as const,
  members: (slug: string) =>
    [...workspaceKeys.detail(slug), "members"] as const,
  invitations: (slug: string) =>
    [...workspaceKeys.detail(slug), "invitations"] as const,
};

export const workspaceQueries = {
  list: () =>
    queryOptions({
      queryKey: workspaceKeys.list(),
      queryFn: () => api.get("workspaces").json<WorkspaceWithRole[]>(),
      staleTime: QUERY_CACHE.SESSION_STALE_TIME,
    }),

  detail: (slug: string) =>
    queryOptions({
      queryKey: workspaceKeys.detail(slug),
      queryFn: () => api.get(`workspaces/${slug}`).json<WorkspaceWithRole>(),
    }),

  members: (slug: string) =>
    queryOptions({
      queryKey: workspaceKeys.members(slug),
      queryFn: () => api.get(`workspaces/${slug}/members`).json<Member[]>(),
    }),

  invitations: (slug: string) =>
    queryOptions({
      queryKey: workspaceKeys.invitations(slug),
      queryFn: () =>
        api.get(`workspaces/${slug}/invitations`).json<Invitation[]>(),
    }),
};

export async function createWorkspace(input: CreateInput) {
  return api.post("workspaces", { json: input }).json<WorkspaceWithRole>();
}

export async function updateWorkspace(slug: string, input: UpdateInput) {
  return api
    .patch(`workspaces/${slug}`, { json: input })
    .json<WorkspaceWithRole>();
}

export async function deleteWorkspace(slug: string) {
  await api.delete(`workspaces/${slug}`);
}

export async function updateMemberRole(
  slug: string,
  userId: string,
  role: string,
) {
  await api.patch(`workspaces/${slug}/members/${userId}`, {
    json: { role },
  });
}

export async function removeMember(slug: string, userId: string) {
  await api.delete(`workspaces/${slug}/members/${userId}`);
}

export async function inviteMember(
  slug: string,
  input: { email: string; role: string },
) {
  return api
    .post(`workspaces/${slug}/invitations`, { json: input })
    .json<Invitation>();
}

export async function cancelInvitation(slug: string, invitationId: string) {
  await api.delete(`workspaces/${slug}/invitations/${invitationId}`);
}

export async function acceptInvitation(token: string) {
  return api
    .post("workspaces/accept-invitation", { json: { token } })
    .json<WorkspaceWithRole>();
}

