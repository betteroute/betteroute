import { queryOptions } from "@tanstack/react-query";
import { env } from "@/env";
import { api } from "@/lib/api";
import { QUERY_CACHE } from "@/lib/constants";
import type { MagicLinkInput, VerifyMagicLinkInput } from "./schemas";
import type { User } from "./types";

export const authKeys = {
  all: ["auth"] as const,
  session: () => [...authKeys.all, "session"] as const,
};

export const authQueries = {
  session: () =>
    queryOptions({
      queryKey: authKeys.session(),
      queryFn: () => api.get("auth/me").json<User>(),
      staleTime: QUERY_CACHE.SESSION_STALE_TIME,
      retry: false,
    }),
};

export async function sendMagicLink(input: MagicLinkInput) {
  await api.post("auth/magic-link", { json: input });
}

export async function verifyMagicLink(input: VerifyMagicLinkInput) {
  return api.post("auth/verify-magic-link", { json: input }).json<User>();
}

export async function logout() {
  await api.post("auth/logout");
}

export function getOAuthURL(provider: "google" | "github") {
  const base = env.VITE_API_URL?.replace(/\/+$/, "") ?? "";
  return `${base}/auth/oauth/${provider}`;
}
