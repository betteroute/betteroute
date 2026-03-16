import { Search, X } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useDebouncedCallback } from "@/hooks/use-debounced-callback";

interface DebouncedSearchInputProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  delay?: number;
}

export function DebouncedSearchInput({
  value,
  onChange,
  placeholder = "Search…",
  delay = 300,
}: DebouncedSearchInputProps) {
  const [input, setInput] = useState(value);
  const debouncedOnChange = useDebouncedCallback(onChange, delay);

  // Sync external value → local (e.g. when filters are cleared externally)
  useEffect(() => {
    setInput(value);
  }, [value]);

  return (
    <div className="relative ml-auto w-64">
      <Search
        data-slot="icon"
        className="pointer-events-none absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground"
      />

      <Input
        value={input}
        onChange={(e) => {
          setInput(e.target.value);
          debouncedOnChange(e.target.value);
        }}
        placeholder={placeholder}
        className="pl-7"
      />

      {input && (
        <Button
          variant="ghost"
          size="icon-xs"
          className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground"
          onClick={() => {
            setInput("");
            debouncedOnChange.cancel();
            onChange("");
          }}
        >
          <X data-slot="icon" />
        </Button>
      )}
    </div>
  );
}
