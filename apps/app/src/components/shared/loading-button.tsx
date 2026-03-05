import { Button } from "@/components/ui/button";
import { Spinner } from "./spinner";

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
      {loading && <Spinner size="sm" />}
      {children}
    </Button>
  );
}
