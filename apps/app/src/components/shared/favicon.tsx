import { Link2 } from "lucide-react";
import { useState } from "react";
import { getFaviconUrl } from "@/lib/url-utils";
import { cn } from "@/lib/utils";

interface FaviconProps extends React.HTMLAttributes<HTMLDivElement> {
  url: string;
}

export function Favicon({ url, className, ...props }: FaviconProps) {
  const faviconUrl = getFaviconUrl(url);
  const [hasError, setHasError] = useState(false);

  return (
    <div
      className={cn(
        "flex size-8 shrink-0 items-center justify-center rounded-full bg-muted",
        className,
      )}
      {...props}
    >
      {faviconUrl && !hasError ? (
        <img
          src={faviconUrl}
          alt=""
          className="size-full rounded-full p-1"
          loading="lazy"
          onError={() => setHasError(true)}
        />
      ) : (
        <Link2 data-slot="icon" className="size-4 text-muted-foreground" />
      )}
    </div>
  );
}
