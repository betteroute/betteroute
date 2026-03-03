/**
 * Maps TanStack Form field errors into the `{ message?: string }` shape
 * that shadcn's FieldError component expects. Plain string errors are also supported.
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
export function resolveErrors(
  clientErrors: Array<unknown> | undefined,
  serverError?: string,
): Array<{ message?: string }> | undefined {
  return (
    fieldErrors(clientErrors) ??
    (serverError ? [{ message: serverError }] : undefined)
  );
}
