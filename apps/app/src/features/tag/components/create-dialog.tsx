import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { useState } from "react";

import { ColorPicker, getDefaultColor } from "@/components/shared/color-picker";
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
import { resolveErrors } from "@/lib/form-errors";

import { createTag, tagKeys, updateTag } from "../queries";
import { type CreateInput, createSchema, type UpdateInput } from "../schemas";
import type { Tag } from "../types";

interface CreateTagDialogProps {
  editTag?: Tag;
  onSuccess?: () => void;
}

export function CreateTagDialog({ editTag, onSuccess }: CreateTagDialogProps) {
  const { workspace } = useWorkspace();
  const queryClient = useQueryClient();
  const [open, setOpen] = useState(false);

  const isEdit = !!editTag;

  const mutation = useMutation({
    mutationFn: (input: CreateInput | UpdateInput) => {
      if (isEdit) {
        return updateTag(workspace.slug, editTag.id, input as UpdateInput);
      }
      return createTag(workspace.slug, input as CreateInput);
    },
    onSuccess: async () => {
      setOpen(false);
      onSuccess?.();
      await queryClient.invalidateQueries({
        queryKey: tagKeys.list(workspace.slug),
      });
    },
  });

  const form = useForm({
    defaultValues: {
      name: editTag?.name ?? "",
      color: editTag?.color ?? getDefaultColor(),
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
        if (v && !isEdit) {
          form.setFieldValue("color", getDefaultColor());
        }
        if (!v) {
          form.reset();
          mutation.reset();
        }
      }}
    >
      <DialogTrigger asChild>
        {editTag ? null : (
          <Button>
            <Plus />
            Create tag
          </Button>
        )}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{isEdit ? "Edit tag" : "Create tag"}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? "Update your tag details below."
              : "Create a new tag to organize your links."}
          </DialogDescription>
        </DialogHeader>
        <form
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            form.handleSubmit();
          }}
        >
          <form.Field name="name">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Name</FieldLabel>
                <Input
                  id={field.name}
                  value={field.state.value}
                  onChange={(e) => field.handleChange(e.target.value)}
                  onBlur={field.handleBlur}
                  disabled={mutation.isPending}
                  autoFocus
                  aria-invalid={!!field.state.meta.errors.length}
                />
                <FieldError errors={resolveErrors(field.state.meta.errors)} />
              </Field>
            )}
          </form.Field>

          <form.Field name="color">
            {(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Color</FieldLabel>
                <div className="mt-2">
                  <ColorPicker
                    value={field.state.value}
                    onChange={(color) => field.handleChange(color)}
                    disabled={mutation.isPending}
                  />
                </div>
                <FieldError errors={resolveErrors(field.state.meta.errors)} />
              </Field>
            )}
          </form.Field>

          <DialogFooter>
            <LoadingButton
              type="submit"
              loading={mutation.isPending}
              disabled={mutation.isPending}
            >
              {isEdit ? "Save changes" : "Create tag"}
            </LoadingButton>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
