// Query cache durations (in milliseconds)
export const QUERY_CACHE = {
  DEFAULT_STALE_TIME: 2 * 60 * 1000, // 2 minutes (Standard for Details/Lists)
  SESSION_STALE_TIME: 5 * 60 * 1000, // 5 minutes (Static Infrastructure)
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
