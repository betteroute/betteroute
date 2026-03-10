import { queryOptions } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { QUERY_CACHE } from "@/lib/constants";
import type { CreateInput, UpdateInput } from "./schemas";
import type { Folder } from "./types";

export const folderKeys = {
  all: ["folders"] as const,
  list: (slug: string) => [...folderKeys.all, slug, "list"] as const,
  detail: (slug: string, id: string) => [...folderKeys.all, slug, id] as const,
};

export const folderQueries = {
  list: (slug: string) =>
    queryOptions({
      queryKey: folderKeys.list(slug),
      queryFn: () => api.get(`workspaces/${slug}/folders`).json<Folder[]>(),
      staleTime: QUERY_CACHE.DEFAULT_STALE_TIME,
    }),

  detail: (slug: string, id: string) =>
    queryOptions({
      queryKey: folderKeys.detail(slug, id),
      queryFn: () => api.get(`workspaces/${slug}/folders/${id}`).json<Folder>(),
      staleTime: QUERY_CACHE.DETAIL_STALE_TIME,
    }),
};

export async function createFolder(slug: string, input: CreateInput) {
  return api.post(`workspaces/${slug}/folders`, { json: input }).json<Folder>();
}

export async function updateFolder(
  slug: string,
  id: string,
  input: UpdateInput,
) {
  return api
    .patch(`workspaces/${slug}/folders/${id}`, { json: input })
    .json<Folder>();
}

export async function deleteFolder(slug: string, id: string) {
  await api.delete(`workspaces/${slug}/folders/${id}`);
}
