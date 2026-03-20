import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { LoadingButton } from "@/components/shared/loading-button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { createWorkspace, workspaceKeys } from "@/features/workspace/queries";
import { createSchema } from "@/features/workspace/schemas";
import { getFieldErrors, resolveFieldErrors } from "@/lib/errors";
import { slugify } from "@/lib/url-utils";

export function CreateWorkspaceForm({
  onSuccess,
  onAfterCreate,
  autoFocus = true,
}: {
  onSuccess?: () => void;
  onAfterCreate?: (slug: string) => void;
  autoFocus?: boolean;
}) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createWorkspace,
    onSuccess: async (ws) => {
      onSuccess?.();
      await queryClient.invalidateQueries({ queryKey: workspaceKeys.list() });
      if (onAfterCreate) {
        onAfterCreate(ws.slug);
      } else {
        navigate({ to: "/$slug", params: { slug: ws.slug } });
      }
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  const form = useForm({
    defaultValues: { name: "", slug: "" },
    validators: {
      onSubmit: createSchema.required({ slug: true }),
    },
    onSubmit: async ({ value }) => {
      await mutation.mutateAsync({
        name: value.name,
        slug: value.slug,
      });
    },
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
      className="flex flex-col gap-4"
    >
      <form.Field name="name">
        {(field) => (
          <Field>
            <FieldLabel htmlFor={field.name}>Workspace name</FieldLabel>
            <Input
              id={field.name}
              placeholder="Acme, Inc."
              value={field.state.value}
              onChange={(e) => {
                field.handleChange(e.target.value);
                const slugField = form.getFieldValue("slug");
                const autoSlug = slugify(field.state.value);
                if (!slugField || slugField === autoSlug) {
                  form.setFieldValue("slug", slugify(e.target.value));
                }
              }}
              onBlur={field.handleBlur}
              disabled={mutation.isPending}
              autoFocus={autoFocus}
              aria-invalid={
                !!field.state.meta.errors.length || !!serverErrors?.name
              }
            />
            <FieldError
              errors={resolveFieldErrors(
                field.state.meta.errors,
                serverErrors?.name,
              )}
            />
          </Field>
        )}
      </form.Field>

      <form.Field name="slug">
        {(field) => (
          <Field>
            <FieldLabel htmlFor={field.name}>Workspace URL</FieldLabel>
            <div className="flex">
              <span className="bg-muted text-muted-foreground border-input inline-flex shrink-0 items-center whitespace-nowrap rounded-l-md border border-r-0 px-3 text-sm">
                betteroute.co/
              </span>
              <Input
                id={field.name}
                placeholder="acme"
                value={field.state.value}
                onChange={(e) => field.handleChange(slugify(e.target.value))}
                onBlur={field.handleBlur}
                disabled={mutation.isPending}
                className="rounded-l-none"
                aria-invalid={
                  !!field.state.meta.errors.length || !!serverErrors?.slug
                }
              />
            </div>
            {field.state.value && !field.state.meta.errors.length && (
              <p className="text-muted-foreground text-xs">
                Your workspace: betteroute.co/
                <span className="text-foreground font-medium">
                  {field.state.value}
                </span>
              </p>
            )}
            <FieldError
              errors={resolveFieldErrors(
                field.state.meta.errors,
                serverErrors?.slug,
              )}
            />
          </Field>
        )}
      </form.Field>

      <LoadingButton
        type="submit"
        size="lg"
        className="mt-1 w-full"
        loading={mutation.isPending}
      >
        Create workspace
      </LoadingButton>
    </form>
  );
}
