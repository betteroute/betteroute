import { MousePointerClick } from "lucide-react";

import { CopyButton } from "@/components/shared/copy-button";
import { Favicon } from "@/components/shared/favicon";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { LinkActions } from "@/features/link/components/actions";
import type { Link } from "@/features/link/types";
import { nFormatter } from "@/lib/number-utils";
import { timeAgo } from "@/lib/relative-time";
import { stripUrl } from "@/lib/url-utils";

type StatusInfo = {
  label: string;
  variant: "secondary" | "destructive" | "outline";
};

function getLinkStatus(link: Link): StatusInfo | null {
  if (!link.is_active) return { label: "Inactive", variant: "secondary" };
  if (link.expires_at && new Date(link.expires_at) < new Date())
    return { label: "Expired", variant: "destructive" };
  if (link.starts_at && new Date(link.starts_at) > new Date())
    return { label: "Scheduled", variant: "outline" };
  return null;
}

export function LinkCard({ link }: { link: Link }) {
  const status = getLinkStatus(link);

  return (
    <div className="flex items-center gap-3 rounded-lg border bg-card px-3 py-4">
      {/* Favicon */}
      <Favicon url={link.dest_url} />

      {/* Link info — grows to fill */}
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-1.5">
          <span className="truncate text-sm font-medium">
            {stripUrl(link.short_url || link.short_code)}
          </span>
          <CopyButton value={link.short_url || link.short_code} />
          {status && (
            <span className="shrink-0 rounded-md border border-dashed px-1.5 py-0.5 text-xs text-muted-foreground">
              {status.label}
            </span>
          )}
        </div>
        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
          <a
            href={link.dest_url}
            target="_blank"
            rel="noopener noreferrer"
            className="truncate hover:text-foreground"
          >
            {stripUrl(link.dest_url)}
          </a>
          <span className="shrink-0">·</span>
          <Tooltip>
            <TooltipTrigger asChild>
              <span className="shrink-0 cursor-default">
                {timeAgo(link.created_at)}
              </span>
            </TooltipTrigger>
            <TooltipContent>
              {new Date(link.created_at).toLocaleDateString("en-US", {
                weekday: "short",
                year: "numeric",
                month: "short",
                day: "numeric",
                hour: "2-digit",
                minute: "2-digit",
              })}
            </TooltipContent>
          </Tooltip>
        </div>
      </div>

      {/* Clicks — clickable, will navigate to analytics */}
      <Button
        variant="secondary"
        size="sm"
        className="shrink-0 gap-1 tabular-nums"
      >
        <MousePointerClick />
        {nFormatter(link.click_count)} clicks
      </Button>

      {/* Actions */}
      <LinkActions link={link} />
    </div>
  );
}
