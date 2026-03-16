import { Check, X } from "lucide-react";
import { type ReactNode, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Command,
  CommandGroup,
  CommandItem,
  CommandList,
} from "@/components/ui/command";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";

export interface FilterOption {
  value: string;
  label: string;
  icon?: ReactNode;
}

export interface FilterDefinition {
  key: string;
  label: string;
  icon?: ReactNode;
  options: FilterOption[];
}

export type FilterValues = Record<string, string[] | undefined>;

interface FilterSheetProps {
  filters: FilterDefinition[];
  values: FilterValues;
  onChange: (values: FilterValues) => void;
}

export function FilterSheet({ filters, values, onChange }: FilterSheetProps) {
  const hasAnyFilter = filters.some((f) => (values[f.key]?.length ?? 0) > 0);

  return (
    <div className="flex flex-wrap items-center gap-2">
      {filters.map((filter) => (
        <FacetedFilter
          key={filter.key}
          filter={filter}
          selected={values[filter.key]}
          onSelect={(v) => onChange({ ...values, [filter.key]: v })}
        />
      ))}
      {hasAnyFilter && (
        <Button
          variant="ghost"
          size="sm"
          className="text-muted-foreground hover:text-destructive h-7 px-2"
          onClick={() => onChange({})}
        >
          <X data-slot="icon" className="mr-1" />
          Clear
        </Button>
      )}
    </div>
  );
}

function FacetedFilter({
  filter,
  selected,
  onSelect,
}: {
  filter: FilterDefinition;
  selected: string[] | undefined;
  onSelect: (value: string[] | undefined) => void;
}) {
  const [open, setOpen] = useState(false);
  const selectedSet = new Set(selected ?? []);

  function toggle(value: string) {
    const next = new Set(selectedSet);
    if (next.has(value)) {
      next.delete(value);
    } else {
      next.add(value);
    }
    onSelect(next.size > 0 ? Array.from(next) : undefined);
  }

  const selectedLabels = filter.options
    .filter((opt) => selectedSet.has(opt.value))
    .map((opt) => opt.label);

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className={selectedSet.size > 0 ? "border-solid" : "border-dashed"}
        >
          {filter.icon && (
            <span className="text-muted-foreground">{filter.icon}</span>
          )}
          {selectedSet.size > 0 ? (
            <span className="truncate max-w-24">
              {selectedLabels.join(", ")}
            </span>
          ) : (
            filter.label
          )}
          {selectedSet.size > 0 && (
            <Badge
              variant="default"
              className="ml-1.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] tabular-nums leading-none"
            >
              {selectedSet.size}
            </Badge>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="w-48 p-0">
        <Command>
          <CommandList>
            <CommandGroup>
              {filter.options.map((option) => (
                <CommandItem
                  key={option.value}
                  data-checked={selectedSet.has(option.value)}
                  onSelect={() => toggle(option.value)}
                  className="relative"
                >
                  {selectedSet.has(option.value) && (
                    <Check data-slot="icon" className="absolute right-2" />
                  )}
                  {option.icon && (
                    <span className="text-muted-foreground mr-0.5">
                      {option.icon}
                    </span>
                  )}
                  {option.label}
                </CommandItem>
              ))}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
