// Query cache durations (in milliseconds)
export const QUERY_CACHE = {
  DEFAULT_STALE_TIME: 60 * 1000, // 1 minute
  SESSION_STALE_TIME: 5 * 60 * 1000, // 5 minutes
  DETAIL_STALE_TIME: 2 * 60 * 1000, // 2 minutes
  RETRY_ATTEMPTS: 1,
} as const;

// API timeouts (in milliseconds)
export const API_TIMEOUT = {
  DEFAULT: 15_000, // 15 seconds
} as const;

// Pagination defaults
export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_PER_PAGE: 20,
} as const;
