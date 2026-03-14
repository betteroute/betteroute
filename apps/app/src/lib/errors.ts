import { type HTTPError, isHTTPError } from "ky";
import { toast } from "sonner";
import type { ApiError } from "@/types/common";

/** HTTPError with the RFC 9457 body already parsed and attached. */
export interface ApiHTTPError extends HTTPError {
  apiError: ApiError;
}

/** Type-guard: narrows an unknown error to an ApiHTTPError. */
export function isApiError(error: unknown): error is ApiHTTPError {
  return isHTTPError(error) && "apiError" in error;
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

/** Global error handler — shows a toast with the API error detail. */
export function onGlobalError(error: unknown) {
  if (isApiError(error)) {
    const { status, detail, title, retry_after } = error.apiError;
    if (status === 429) {
      toast.error(`Too many requests. Try again in ${retry_after ?? 60}s.`);
    } else {
      toast.error(detail || title);
    }
  } else {
    toast.error("Something went wrong. Please try again.");
  }
}

/**
 * Maps TanStack Form field errors into the `{ message?: string }` shape
 * that the FieldError component expects. Plain string errors are also supported.
 */
function fieldErrors(
  errors: Array<unknown> | undefined,
): Array<{ message?: string }> | undefined {
  if (!errors?.length) return undefined;
  return errors.map((e) => ({
    message: typeof e === "string" ? e : (e as { message?: string }).message,
  }));
}

/**
 * Merges client-side validation errors with server field errors.
 * Returns the client errors if present, otherwise the server error for that field.
 */
export function resolveFieldErrors(
  clientErrors: Array<unknown> | undefined,
  serverError?: string,
): Array<{ message?: string }> | undefined {
  return (
    fieldErrors(clientErrors) ??
    (serverError ? [{ message: serverError }] : undefined)
  );
}
