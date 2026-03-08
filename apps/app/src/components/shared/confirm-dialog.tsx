import { useState } from "react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";

interface ConfirmDialogProps {
  /** The trigger element (button, menu item, etc.). */
  trigger: React.ReactNode;
  /** Dialog title. */
  title: string;
  /** Description explaining the action. */
  description: string;
  /** Label for the confirm button. Defaults to "Delete". */
  confirmLabel?: string;
  /** Label shown while confirming. Defaults to "Deleting…". */
  pendingLabel?: string;
  /** Callback when confirmed. Can be async — dialog stays open until resolved. */
  onConfirm: () => void | Promise<void>;
  /** Whether the confirm action is in progress. */
  pending?: boolean;
  /** Button variant. Defaults to "destructive". */
  variant?: "destructive" | "default";
}

export function ConfirmDialog({
  trigger,
  title,
  description,
  confirmLabel = "Delete",
  pendingLabel = "Deleting…",
  onConfirm,
  pending = false,
  variant = "destructive",
}: ConfirmDialogProps) {
  const [open, setOpen] = useState(false);

  return (
    <AlertDialog open={open} onOpenChange={setOpen}>
      <AlertDialogTrigger asChild>{trigger}</AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            variant={variant}
            disabled={pending}
            onClick={async (e) => {
              e.preventDefault();
              try {
                await onConfirm();
                setOpen(false);
              } catch {
                // Errors are handled by the caller (e.g. mutation onError / toasts).
              }
            }}
          >
            {pending ? pendingLabel : confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
