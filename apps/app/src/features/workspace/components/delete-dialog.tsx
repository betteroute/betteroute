import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

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
import { Button } from "@/components/ui/button";
import { FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useWorkspace } from "@/features/workspace/hooks";
import { deleteWorkspace, workspaceKeys } from "@/features/workspace/queries";

export function DeleteWorkspaceDialog() {
    const { workspace } = useWorkspace();
    const queryClient = useQueryClient();
    const navigate = useNavigate();
    const [confirmValue, setConfirmValue] = useState("");

    const mutation = useMutation({
        mutationFn: () => deleteWorkspace(workspace.slug),
        onSuccess: async () => {
            await queryClient.refetchQueries({ queryKey: workspaceKeys.all });
            navigate({ to: "/" });
        },
    });

    const isConfirmed = confirmValue === workspace.slug;

    return (
        <section className="space-y-4 rounded-lg border border-destructive/20 p-4">
            <div>
                <h2 className="text-sm font-medium">Delete workspace</h2>

                <p className="text-muted-foreground text-sm">
                    Permanently delete this workspace, all its links, analytics,
                    and data. This action cannot be undone.
                </p>
            </div>

            <AlertDialog
                onOpenChange={(open) => {
                    if (!open) setConfirmValue("");
                }}
            >
                <AlertDialogTrigger asChild>
                    <Button variant="destructive" size="sm">
                        Delete workspace
                    </Button>
                </AlertDialogTrigger>

                <AlertDialogContent>
                    <AlertDialogHeader>
                        <AlertDialogTitle>Delete workspace</AlertDialogTitle>

                        <AlertDialogDescription>
                            This will permanently delete the{" "}
                            <strong>{workspace.name}</strong> workspace and all
                            its data. This action cannot be undone.
                        </AlertDialogDescription>
                    </AlertDialogHeader>

                    <div className="space-y-2">
                        <FieldLabel htmlFor="confirm-delete">
                            Type <strong>{workspace.slug}</strong> to confirm
                        </FieldLabel>

                        <Input
                            id="confirm-delete"
                            value={confirmValue}
                            onChange={(e) => setConfirmValue(e.target.value)}
                            placeholder={workspace.slug}
                            autoComplete="off"
                        />
                    </div>

                    <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>

                        <AlertDialogAction
                            variant="destructive"
                            disabled={!isConfirmed || mutation.isPending}
                            onClick={(e) => {
                                e.preventDefault();
                                mutation.mutate();
                            }}
                        >
                            {mutation.isPending
                                ? "Deleting…"
                                : "Delete workspace"}
                        </AlertDialogAction>
                    </AlertDialogFooter>
                </AlertDialogContent>
            </AlertDialog>
        </section>
    );
}
