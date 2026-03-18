import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
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
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useWorkspace } from "@/features/workspace/hooks";
import { getFieldErrors, resolveFieldErrors } from "@/lib/errors";

import { createLink, linkKeys } from "../queries";
import { type CreateInput, createSchema } from "../schemas";

export function CreateLinkDialog() {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const mutation = useMutation({
    mutationFn: (input: CreateInput) => createLink(workspace.slug, input),
    onSuccess: async () => {
      setOpen(false);
      await queryClient.invalidateQueries({
        queryKey: linkKeys.list(workspace.slug),
      });
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  const form = useForm({
    defaultValues: {
      dest_url: "",
      short_code: "",
      title: "",
      description: "",
    },
    validators: {
      onSubmit: createSchema,
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync(value);
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
        <Button>
          <Plus data-slot="icon" />
          Create link
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create link</DialogTitle>
          <DialogDescription>
            Shorten a URL and track its performance.
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
          className="space-y-4"
        >
          <form.Field
            name="dest_url"
            validators={{
              onChange: createSchema.shape.dest_url,
            }}
          >
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Destination URL</FieldLabel>
                <Input
                  id={field.name}
                  type="url"
                  placeholder="https://example.com"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  disabled={mutation.isPending}
                  autoFocus
                  aria-invalid={
                    !!field.state.meta.errors.length || !!serverErrors?.dest_url
                  }
                />
                <FieldError
                  errors={resolveFieldErrors(
                    field.state.meta.errors,
                    serverErrors?.dest_url,
                  )}
                />
              </Field>
            )}
          </form.Field>

          <form.Field name="short_code">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>
                  Short code{" "}
                  <span className="text-muted-foreground font-normal">
                    (optional)
                  </span>
                </FieldLabel>
                <Input
                  id={field.name}
                  placeholder="my-link"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  disabled={mutation.isPending}
                  aria-invalid={
                    !!field.state.meta.errors.length ||
                    !!serverErrors?.short_code
                  }
                />
                <FieldError
                  errors={resolveFieldErrors(
                    field.state.meta.errors,
                    serverErrors?.short_code,
                  )}
                />
                <p className="text-xs text-muted-foreground">
                  Leave empty to auto-generate a random code.
                </p>
              </Field>
            )}
          </form.Field>

          <form.Field name="title">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>
                  Title{" "}
                  <span className="text-muted-foreground font-normal">
                    (optional)
                  </span>
                </FieldLabel>
                <Input
                  id={field.name}
                  placeholder="My marketing campaign"
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  disabled={mutation.isPending}
                  aria-invalid={
                    !!field.state.meta.errors.length || !!serverErrors?.title
                  }
                />
                <FieldError
                  errors={resolveFieldErrors(
                    field.state.meta.errors,
                    serverErrors?.title,
                  )}
                />
              </Field>
            )}
          </form.Field>

          <DialogFooter>
            <LoadingButton type="submit" loading={mutation.isPending}>
              Create link
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
