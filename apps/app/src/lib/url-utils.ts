/**
 * Strips the protocol (http://, https://) and trailing slash from a URL.
 * e.g. "https://www.pixelship.dev/#contact" → "pixelship.dev/#contact"
 */
export function stripUrl(url: string): string {
    return url.replace(/^https?:\/\//, "").replace(/\/$/, "");
}

/**
 * Returns a Google Favicon URL for a given valid URL string.
 */
export function getFaviconUrl(url: string): string {
    try {
        const hostname = new URL(url).hostname;
        // DuckDuckGo's favicon service returns a 404 if the favicon is not found.
        // This allows our <img onError> handler to correctly trigger and show the Link2 fallback!
        return `https://icons.duckduckgo.com/ip3/${hostname}.ico`;
    } catch {
        return "";
    }
}

/**
 * Converts a string into a URL-safe slug.
 * e.g. "Acme, Inc." → "acme-inc"
 */
export function slugify(value: string): string {
    return value
        .toLowerCase()
        .trim()
        .replace(/[^\w\s-]/g, "")
        .replace(/[\s_]+/g, "-")
        .replace(/^-+|-+$/g, "")
        .slice(0, 50);
}
