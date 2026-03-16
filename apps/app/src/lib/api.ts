import ky from "ky";
import { env } from "@/env";
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
        if (
          response.status === 401 &&
          !window.location.pathname.startsWith("/login") &&
          !window.location.pathname.startsWith("/verify")
        ) {
          queryClient.clear();
          window.location.replace("/login");
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
