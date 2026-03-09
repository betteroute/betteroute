import type { VariantProps } from "class-variance-authority";
import { Check, Copy } from "lucide-react";
import { useState } from "react";

import { Button, type buttonVariants } from "@/components/ui/button";
import { copyToClipboard } from "@/lib/clipboard";
import { cn } from "@/lib/utils";

interface CopyButtonProps
  extends React.ComponentProps<"button">,
    VariantProps<typeof buttonVariants> {
  /** The string value to be copied to the clipboard. */
  value: string;
}

export function CopyButton({
  value,
  className,
  variant = "ghost",
  size = "icon-xs",
  ...props
}: CopyButtonProps) {
  const [hasCopied, setHasCopied] = useState(false);

  async function handleCopy() {
    await copyToClipboard(value);
    setHasCopied(true);
    setTimeout(() => {
      setHasCopied(false);
    }, 2000); // Reset the icon after 2 seconds
  }

  return (
    <Button
      variant={variant}
      size={size}
      className={cn("shrink-0", "text-muted-foreground relative", className)}
      onClick={handleCopy}
      title="Copy to clipboard"
      {...props}
    >
      {hasCopied ? <Check /> : <Copy />}
    </Button>
  );
}
