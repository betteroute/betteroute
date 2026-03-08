import { useEffect, useState } from "react";

const MS_PER_MINUTE = 60000;
const MS_PER_HOUR = 3600000;
const MS_PER_DAY = 86400000;

/** * Returns a compact relative time string: "11h", "2d", "3mo", "1y".
 */
export function timeAgo(date: string | Date | undefined | null): string {
  if (!date) return "";

  const diffMs = Date.now() - new Date(date).getTime();

  if (diffMs < MS_PER_MINUTE) return "now"; // Less than 1 minute

  const days = Math.floor(diffMs / MS_PER_DAY);

  if (days >= 365) return `${Math.floor(days / 365)}y`;
  if (days >= 30) return `${Math.floor(days / 30)}mo`;
  if (days >= 7) return `${Math.floor(days / 7)}w`;
  if (days >= 1) return `${days}d`;

  const hours = Math.floor(diffMs / MS_PER_HOUR);
  if (hours >= 1) return `${hours}h`;

  const minutes = Math.floor(diffMs / MS_PER_MINUTE);
  return `${minutes}m`;
}

/**
 * React hook that provides automatically updating relative time strings.
 * Updates every minute for real-time display.
 */
export function useRelativeTime(date: string | Date | undefined | null) {
  const [time, setTime] = useState(() => timeAgo(date));

  useEffect(() => {
    const interval = setInterval(() => {
      setTime(timeAgo(date));
    }, MS_PER_MINUTE); // Update every minute

    return () => clearInterval(interval);
  }, [date]);

  return time;
}

/**
 * Returns relative time for future dates: "Expires today", "Expires tomorrow", "Expires in 3d"
 */
export function expiresIn(date: string | Date | undefined | null): string {
  if (!date) return "";

  const diffMs = new Date(date).getTime() - Date.now();

  if (diffMs <= 0) return "Expired";

  const days = Math.floor(diffMs / MS_PER_DAY);
  if (days >= 1) return `Expires in ${days}d`;

  const hours = Math.floor(diffMs / MS_PER_HOUR);
  if (hours >= 1) return `Expires in ${hours}h`;

  const minutes = Math.floor(diffMs / MS_PER_MINUTE);
  return `Expires in ${minutes || 1}m`;
}
