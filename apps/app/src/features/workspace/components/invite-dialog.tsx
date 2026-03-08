import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, ChevronsUpDown } from "lucide-react";
import { useState } from "react";

import { LoadingButton } from "@/components/shared/loading-button";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useWorkspace } from "@/features/workspace/hooks";
import { inviteMember, workspaceKeys } from "@/features/workspace/queries";
import { inviteSchema } from "@/features/workspace/schemas";
import { ASSIGNABLE_ROLES, ROLE_INFO } from "@/features/workspace/types";
import { getFieldErrors } from "@/lib/errors";
import { resolveErrors } from "@/lib/form-errors";

export function InviteDialog() {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const mutation = useMutation({
    mutationFn: (input: { email: string; role: string }) =>
      inviteMember(workspace.slug, input),
    onSuccess: () => {
      queryClient.refetchQueries({
        queryKey: workspaceKeys.invitations(workspace.slug),
      });
      setOpen(false);
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  const form = useForm({
    defaultValues: { email: "", role: "member" as string },
    validators: {
      onSubmit: inviteSchema,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({ email: value.email, role: value.role });
    },
  });

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        setOpen(v);
        if (!v) {
          form.reset();
          mutation.reset();
        }
      }}
    >
      <DialogTrigger asChild>
        <Button>Invite member</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Invite member</DialogTitle>
          <DialogDescription>
            Send an invitation to join this workspace.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
          className="space-y-4"
        >
          <form.Field name="role">
            {(field) => (
              <Field>
                <FieldLabel>Role</FieldLabel>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button
                      variant="outline"
                      className="w-full justify-between"
                      type="button"
                    >
                      {ROLE_INFO[field.state.value]?.label ?? field.state.value}
                      <ChevronsUpDown className="size-4 opacity-50" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent
                    align="start"
                    className="w-[var(--radix-dropdown-menu-trigger-width)]"
                  >
                    {ASSIGNABLE_ROLES.map((role) => (
                      <DropdownMenuItem
                        key={role}
                        onSelect={() => {
                          field.handleChange(role);
                        }}
                        className="flex items-center justify-between"
                      >
                        {ROLE_INFO[role]?.label ?? role}
                        {role === field.state.value && (
                          <Check className="size-4" />
                        )}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
                {ROLE_INFO[field.state.value] && (
                  <p className="text-xs text-muted-foreground">
                    {ROLE_INFO[field.state.value]?.description}
                  </p>
                )}
                <FieldError
                  errors={resolveErrors(
                    field.state.meta.errors,
                    serverErrors?.role,
                  )}
                />
              </Field>
            )}
          </form.Field>

          <form.Field name="email">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Email</FieldLabel>
                <Input
                  id={field.name}
                  type="email"
                  placeholder="colleague@example.com"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  disabled={mutation.isPending}
                  autoFocus
                  aria-invalid={
                    !!field.state.meta.errors.length || !!serverErrors?.email
                  }
                />
                <FieldError
                  errors={resolveErrors(
                    field.state.meta.errors,
                    serverErrors?.email,
                  )}
                />
              </Field>
            )}
          </form.Field>

          <DialogFooter>
            <LoadingButton type="submit" loading={mutation.isPending}>
              Send invitation
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
