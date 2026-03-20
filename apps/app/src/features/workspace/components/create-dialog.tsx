import { useState } from "react";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { CreateWorkspaceForm } from "@/features/workspace/components/create-form";
import { PlanStep } from "@/features/workspace/components/plan-step";

export function CreateWorkspaceDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const [step, setStep] = useState<"form" | "plan">("form");
  const [newSlug, setNewSlug] = useState("");

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      setTimeout(() => {
        setStep("form");
        setNewSlug("");
      }, 300);
    }
    onOpenChange(isOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-lg">
        {step === "form" ? (
          <>
            <DialogHeader>
              <DialogTitle>Create workspace</DialogTitle>
              <DialogDescription>
                Add a new workspace for your team or project.
              </DialogDescription>
            </DialogHeader>
            <CreateWorkspaceForm
              onAfterCreate={(slug) => {
                setNewSlug(slug);
                setStep("plan");
              }}
              autoFocus
            />
          </>
        ) : (
          <div className="py-2">
            <PlanStep
              slug={newSlug}
              onContinue={() => handleOpenChange(false)}
            />
          </div>
        )}
      </DialogContent>
    </Dialog>
  );
}
