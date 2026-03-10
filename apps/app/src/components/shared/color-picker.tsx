import { Check } from "lucide-react";
import { Input } from "@/components/ui/input";

import { cn } from "@/lib/utils";

export const PRESET_COLORS = [
  "#E5484D", // Red
  "#F76B15", // Orange
  "#F5D90A", // Yellow
  "#46A758", // Green
  "#00A2C7", // Cyan
  "#0090FF", // Blue
  "#8E4EC6", // Purple
  "#D6409F", // Pink
  "#3E63DD", // Indigo
  "#10B981", // Emerald
  "#F59E0B", // Amber
  "#64748B", // Slate
  "#F43F5E", // Rose
];

/** Returns a random color from the preset palette. */
export function getDefaultColor() {
  return PRESET_COLORS[Math.floor(Math.random() * PRESET_COLORS.length)];
}

interface ColorPickerProps {
  value?: string;
  onChange: (value: string) => void;
  disabled?: boolean;
}

export function ColorPicker({ value, onChange, disabled }: ColorPickerProps) {
  const selectedColor = value ?? PRESET_COLORS[0];

  return (
    <>
      <div className="flex flex-wrap gap-2">
        {PRESET_COLORS.map((color) => {
          const isSelected =
            !!selectedColor &&
            selectedColor.toUpperCase() === color.toUpperCase();
          return (
            <button
              key={color}
              type="button"
              className={cn(
                "flex size-6 items-center justify-center rounded-full transition-all",
                "hover:scale-110 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-1",
                disabled && "cursor-not-allowed opacity-50",
                isSelected && "ring-2 ring-ring ring-offset-2",
              )}
              style={{ backgroundColor: color }}
              onClick={() => {
                if (!disabled) {
                  onChange(color);
                }
              }}
              aria-label={`Select color ${color}`}
              aria-pressed={isSelected}
            >
              {isSelected && (
                <Check className="size-3.5 text-white drop-shadow-sm mix-blend-difference" />
              )}
            </button>
          );
        })}
      </div>

      <div className="mt-4 flex items-center gap-3">
        <div className="text-muted-foreground mr-1 text-xs font-medium tracking-wider">
          CUSTOM HEX
        </div>
        <Input
          type="text"
          value={selectedColor}
          onChange={(e) => {
            if (!disabled) {
              let val = e.target.value;
              if (val && !val.startsWith("#")) {
                val = `#${val}`;
              }
              onChange(val.substring(0, 7));
            }
          }}
          disabled={disabled}
          className="w-24 text-xs font-mono uppercase"
          placeholder="#8E8E93"
          maxLength={7}
        />
        <div
          className="size-7 shrink-0 rounded-md border"
          style={{ backgroundColor: selectedColor || "transparent" }}
          aria-hidden="true"
        />
      </div>
    </>
  );
}
