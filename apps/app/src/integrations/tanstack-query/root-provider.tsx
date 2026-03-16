import {
  MutationCache,
  QueryCache,
  QueryClient,
  QueryClientProvider,
} from "@tanstack/react-query";
import type { ReactNode } from "react";
import { QUERY_CACHE } from "@/lib/constants";
import { isApiError, onGlobalError } from "@/lib/errors";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: QUERY_CACHE.DEFAULT_STALE_TIME,
      retry: (failureCount, error) => {
        // Never retry 401/403/404 — they won't succeed on retry
        if (isApiError(error)) {
          const status = error.apiError.status;
          if (status === 401 || status === 403 || status === 404) return false;
        }
        return failureCount < QUERY_CACHE.RETRY_ATTEMPTS;
      },
    },
  },
  queryCache: new QueryCache({
    onError: (error, query) => {
      // Only show toast for queries that already have data (background refetch failed)
      // Don't toast on initial load failures — the UI should handle those
      if (query.state.data !== undefined) {
        onGlobalError(error);
      }
    },
  }),
  mutationCache: new MutationCache({
    onError: (error) => {
      // 422 field errors are shown inline by the form — don't double-toast
      if (isApiError(error) && error.apiError.status === 422) return;
      onGlobalError(error);
    },
  }),
});

let context: { queryClient: QueryClient } | undefined;

export function getContext() {
  if (context) return context;

  context = { queryClient };

  return context;
}

export default function TanStackQueryProvider({
  children,
}: {
  children: ReactNode;
}) {
  const { queryClient } = getContext();

  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
}
