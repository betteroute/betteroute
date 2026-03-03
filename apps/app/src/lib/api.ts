import ky, { HTTPError } from "ky";
import { toast } from "sonner";
import { env } from "@/env";
import type { ApiError } from "@/types/common";

/** HTTPError with the RFC 9457 body already parsed and attached. */
export interface ApiHTTPError extends HTTPError {
  apiError: ApiError;
}

/** Type-guard: narrows an unknown error to an ApiHTTPError. */
export function isApiError(error: unknown): error is ApiHTTPError {
  return error instanceof HTTPError && "apiError" in error;
}

/** Extract per-field validation errors from an API error (422 responses). */
export function getFieldErrors(
  error: unknown,
): Record<string, string> | undefined {
  if (!isApiError(error)) return undefined;
  const fields = error.apiError.errors;
  if (!fields?.length) return undefined;
  const map: Record<string, string> = {};
  for (const f of fields) {
    map[f.field] = f.message;
  }
  return map;
}

export const api = ky.create({
  prefixUrl: env.VITE_API_URL,
  credentials: "include",
  timeout: 15_000,
  hooks: {
    afterResponse: [
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

/** Default mutation error handler — shows a toast with the API error detail. */
export function onMutationError(error: unknown) {
  if (isApiError(error)) {
    toast.error(error.apiError.detail || error.apiError.title);
  } else {
    toast.error("Something went wrong. Please try again.");
  }
}
