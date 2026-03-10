import { HTTPError } from "ky";
import { toast } from "sonner";
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

/** Default mutation error handler — shows a toast with the API error detail. */
export function onMutationError(error: unknown) {
  if (isApiError(error)) {
    toast.error(error.apiError.detail || error.apiError.title);
  } else {
    toast.error("Something went wrong. Please try again.");
  }
}
