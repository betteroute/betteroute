import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";

import { LoadingButton } from "@/components/shared/loading-button";
import { Field, FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useWorkspace } from "@/features/workspace/hooks";
import { updateWorkspace, workspaceKeys } from "@/features/workspace/queries";
import { updateSchema } from "@/features/workspace/schemas";
import { getFieldErrors } from "@/lib/errors";
import { resolveErrors } from "@/lib/form-errors";
import { slugify } from "@/lib/url-utils";

function WorkspaceNameForm() {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: (name: string) => updateWorkspace(workspace.slug, { name }),
    onSuccess: () => {
      queryClient.refetchQueries({ queryKey: workspaceKeys.all });
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  const form = useForm({
    defaultValues: { name: workspace.name },
    validators: {
      onSubmit: updateSchema.pick({ name: true }).required(),
    },
    onSubmit: async ({ value }) => {
      if (!value.name) return;
      await mutation.mutateAsync(value.name);
    },
  });

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-sm font-medium">Workspace name</h2>

        <p className="text-muted-foreground text-sm">
          This is the display name of your workspace.
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
        className="flex max-w-md items-end gap-3"
      >
        <form.Field name="name">
          {(field) => (
            <Field className="flex-1">
              <Input
                id={field.name}
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
                disabled={mutation.isPending}
                aria-invalid={
                  !!field.state.meta.errors.length || !!serverErrors?.name
                }
              />

              <FieldError
                errors={resolveErrors(
                  field.state.meta.errors,
                  serverErrors?.name,
                )}
              />
            </Field>
          )}
        </form.Field>

        <LoadingButton type="submit" loading={mutation.isPending}>
          Save
        </LoadingButton>
      </form>
    </section>
  );
}

function WorkspaceSlugForm() {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const mutation = useMutation({
    mutationFn: (slug: string) => updateWorkspace(workspace.slug, { slug }),
    onSuccess: async (ws) => {
      await queryClient.refetchQueries({ queryKey: workspaceKeys.all });
      navigate({ to: "/$slug/settings", params: { slug: ws.slug } });
    },
  });

  const serverErrors = getFieldErrors(mutation.error);

  const form = useForm({
    defaultValues: { slug: workspace.slug },
    validators: {
      onSubmit: updateSchema.pick({ slug: true }).required(),
    },
    onSubmit: async ({ value }) => {
      if (!value.slug) return;
      await mutation.mutateAsync(value.slug);
    },
  });

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-sm font-medium">Workspace slug</h2>

        <p className="text-muted-foreground text-sm">
          This is your workspace's URL namespace on Betteroute.
        </p>
      </div>

      <form
        onSubmit={(e) => {
          e.preventDefault();
          form.handleSubmit();
        }}
        className="flex max-w-md items-end gap-3"
      >
        <form.Field name="slug">
          {(field) => (
            <Field className="flex-1">
              <div className="flex">
                <span className="bg-muted text-muted-foreground border-input inline-flex shrink-0 items-center whitespace-nowrap rounded-l-md border border-r-0 px-3 text-sm">
                  betteroute.co/
                </span>

                <Input
                  id={field.name}
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

              <FieldError
                errors={resolveErrors(
                  field.state.meta.errors,
                  serverErrors?.slug,
                )}
              />
            </Field>
          )}
        </form.Field>

        <LoadingButton type="submit" loading={mutation.isPending}>
          Save
        </LoadingButton>
      </form>
    </section>
  );
}

export function GeneralForm() {
  return (
    <>
      <WorkspaceNameForm />
      <WorkspaceSlugForm />
    </>
  );
}
