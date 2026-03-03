import { cn } from "@/lib/utils";

export function Logo({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={cn("size-8", className)}
      aria-hidden="true"
    >
      <rect width="32" height="32" rx="8" className="fill-primary" />
      <path
        d="M8 16.5C8 12.358 11.358 9 15.5 9H18c2.21 0 4 1.79 4 4s-1.79 4-4 4h-3c-1.657 0-3 1.343-3 3s1.343 3 3 3h5"
        stroke="white"
        strokeWidth="2.5"
        strokeLinecap="round"
      />
    </svg>
  );
}
