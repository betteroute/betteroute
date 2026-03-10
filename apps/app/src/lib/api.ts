import ky from "ky";
import { toast } from "sonner";
import { env } from "@/env";
import { authKeys } from "@/features/auth/queries";
import { queryClient } from "@/integrations/tanstack-query/root-provider";
import { API_TIMEOUT } from "@/lib/constants";
import type { ApiError } from "@/types/common";
import type { ApiHTTPError } from "./errors";

export const api = ky.create({
  prefixUrl: env.VITE_API_URL,
  credentials: "include",
  timeout: API_TIMEOUT.DEFAULT,
  hooks: {
    afterResponse: [
      async (_request, _options, response) => {
        if (response.status === 401) {
          queryClient.removeQueries({ queryKey: authKeys.all });
        }
      },
      (_request, _options, response) => {
        if (response.status === 429) {
          const retryHeader = response.headers.get("Retry-After");
          const seconds = retryHeader ? Number.parseInt(retryHeader, 10) : 60;
          toast.error(`Too many requests. Try again in ${seconds}s.`);
        }
      },
    ],
    beforeError: [
      async (error) => {
        try {
          const body = (await error.response.json()) as ApiError;
          (error as ApiHTTPError).apiError = body;
        } catch {
          (error as ApiHTTPError).apiError = {
            type: "client-error",
            title: "Something went wrong",
            status: error.response.status,
            detail: error.response.statusText || "An unexpected error occurred",
          };
        }
        return error;
      },
    ],
  },
});
