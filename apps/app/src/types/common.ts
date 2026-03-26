/**
 * RFC 9457 Problem Details — matches the Go API's error response shape.
 *
 * Always present: type, title, status, instance, request_id
 * Conditional: detail (most errors), errors (422 only), retry_after (429 only)
 */
export interface ApiError {
  type: string;
  title: string;
  status: number;
  detail: string;
  instance?: string;
  request_id?: string;
  errors?: ApiFieldError[];
  retry_after?: number;
}

/** Per-field validation error (part of 422 responses). */
export interface ApiFieldError {
  field: string;
  message: string;
}

/** Paginated list envelope. */
export interface PaginatedResponse<T> {
  data: T[];
  pagination: Pagination;
}

export interface Pagination {
  per_page: number;
  has_more: boolean;
}
