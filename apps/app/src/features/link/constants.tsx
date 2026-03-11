import {
  Ban,
  CalendarClock,
  CircleCheck,
  Filter,
  TimerOff,
} from "lucide-react";
import type { FilterDefinition } from "@/components/shared/filter-sheet";

export const LINK_FILTERS: FilterDefinition[] = [
  {
    key: "status",
    label: "Status",
    icon: <Filter />,
    options: [
      { value: "active", label: "Active", icon: <CircleCheck /> },
      { value: "inactive", label: "Inactive", icon: <Ban /> },
      { value: "expired", label: "Expired", icon: <TimerOff /> },
      { value: "scheduled", label: "Scheduled", icon: <CalendarClock /> },
    ],
  },
];
