import { queryOptions } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { QUERY_CACHE } from "@/lib/constants";
import type { CreateInput, UpdateInput } from "./schemas";
import type { Tag } from "./types";

export const tagKeys = {
  all: ["tags"] as const,
  list: (slug: string) => [...tagKeys.all, slug, "list"] as const,
  detail: (slug: string, id: string) => [...tagKeys.all, slug, id] as const,
};

export const tagQueries = {
  list: (slug: string) =>
    queryOptions({
      queryKey: tagKeys.list(slug),
      queryFn: () => api.get(`workspaces/${slug}/tags`).json<Tag[]>(),
      staleTime: QUERY_CACHE.DEFAULT_STALE_TIME,
    }),

  detail: (slug: string, id: string) =>
    queryOptions({
      queryKey: tagKeys.detail(slug, id),
      queryFn: () => api.get(`workspaces/${slug}/tags/${id}`).json<Tag>(),
      staleTime: QUERY_CACHE.DETAIL_STALE_TIME,
    }),
};

export async function createTag(slug: string, input: CreateInput) {
  return api.post(`workspaces/${slug}/tags`, { json: input }).json<Tag>();
}

export async function updateTag(slug: string, id: string, input: UpdateInput) {
  return api
    .patch(`workspaces/${slug}/tags/${id}`, { json: input })
    .json<Tag>();
}

export async function deleteTag(slug: string, id: string) {
  await api.delete(`workspaces/${slug}/tags/${id}`);
}
