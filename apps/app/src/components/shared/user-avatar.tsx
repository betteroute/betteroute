import type { ComponentProps, ReactNode } from "react";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { getInitials } from "@/lib/initials";
import { cn } from "@/lib/utils";

interface UserAvatarProps extends ComponentProps<typeof Avatar> {
  name: string;
  src?: string | null;
  /** Optional icon to display instead of initials if no image is provided. */
  fallbackIcon?: ReactNode;
}

export function UserAvatar({
  name,
  src,
  fallbackIcon,
  className,
  ...props
}: UserAvatarProps) {
  return (
    <Avatar className={cn("size-8 shrink-0", className)} {...props}>
      {src && <AvatarImage src={src} alt={name} />}
      <AvatarFallback className="text-xs bg-muted text-muted-foreground border border-border/50">
        {fallbackIcon || getInitials(name)}
      </AvatarFallback>
    </Avatar>
  );
}
