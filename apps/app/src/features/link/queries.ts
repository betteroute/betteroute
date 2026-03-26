import { queryOptions } from "@tanstack/react-query";

import { api } from "@/lib/api";
import { PAGINATION } from "@/lib/constants";
import type { PaginatedResponse } from "@/types/common";
import type { CreateInput, UpdateInput } from "./schemas";
import type { Link } from "./types";

export interface LinkFilters {
  offset?: number;
  perPage?: number;
  search?: string;
  status?: string[];
}

export const linkKeys = {
  all: ["links"] as const,
  list: (slug: string, filters: LinkFilters = {}) =>
    [...linkKeys.all, slug, "list", filters] as const,
  detail: (slug: string, id: string) => [...linkKeys.all, slug, id] as const,
};

export const linkQueries = {
  list: (slug: string, filters: LinkFilters = {}) =>
    queryOptions({
      queryKey: linkKeys.list(slug, filters),
      queryFn: () => {
        const params: Record<string, string | number> = {
          per_page: filters.perPage ?? PAGINATION.DEFAULT_PER_PAGE,
        };
        if (filters.offset) params.offset = filters.offset;
        if (filters.search) params.search = filters.search;
        if (filters.status?.length) params.status = filters.status.join(",");

        return api
          .get(`workspaces/${slug}/links`, { searchParams: params })
          .json<PaginatedResponse<Link>>();
      },
    }),

  detail: (slug: string, id: string) =>
    queryOptions({
      queryKey: linkKeys.detail(slug, id),
      queryFn: () => api.get(`workspaces/${slug}/links/${id}`).json<Link>(),
    }),
};

export async function createLink(slug: string, input: CreateInput) {
  return api.post(`workspaces/${slug}/links`, { json: input }).json<Link>();
}

export async function updateLink(slug: string, id: string, input: UpdateInput) {
  return api
    .patch(`workspaces/${slug}/links/${id}`, { json: input })
    .json<Link>();
}

export async function deleteLink(slug: string, id: string) {
  await api.delete(`workspaces/${slug}/links/${id}`);
}
