import { Loader2 } from "lucide-react";

import { Button } from "@/components/ui/button";

export function LoadingButton({
  loading,
  children,
  className,
  disabled,
  ...props
}: React.ComponentProps<typeof Button> & { loading?: boolean }) {
  return (
    <Button
      disabled={loading || disabled}
      aria-busy={loading || undefined}
      className={className}
      {...props}
    >
      {loading && <Loader2 className="mr-2 size-4 animate-spin" />}
      {children}
    </Button>
  );
}
