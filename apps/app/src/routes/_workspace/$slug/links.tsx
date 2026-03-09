import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import {
  Ban,
  CalendarClock,
  ChevronLeft,
  ChevronRight,
  CircleCheck,
  Filter,
  Search,
  TimerOff,
  X,
} from "lucide-react";
import { useState } from "react";

import {
  type FilterDefinition,
  FilterSheet,
  type FilterValues,
} from "@/components/shared/filter-sheet";
import { PageHeader } from "@/components/shared/page-header";
import { PageLoader } from "@/components/shared/page-loader";

import { Button } from "@/components/ui/button";

import { Input } from "@/components/ui/input";

import { CreateLinkDialog } from "@/features/link/components/create-dialog";

import {
  LinksEmptyState,
  LinksFilteredEmptyState,
} from "@/features/link/components/empty-state";

import { LinkCard } from "@/features/link/components/link-card";

import { linkQueries } from "@/features/link/queries";

import { useWorkspace } from "@/features/workspace/hooks";

export const Route = createFileRoute("/_workspace/$slug/links")({
  component: LinksPage,
});

const LINK_FILTERS: FilterDefinition[] = [
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

function LinksPage() {
  const { workspace } = useWorkspace();

  const [page, setPage] = useState(1);

  const [search, setSearch] = useState("");

  const [filterValues, setFilterValues] = useState<FilterValues>({});

  const hasFilters = !!(search || Object.values(filterValues).some(Boolean));

  const query = useQuery(
    linkQueries.list(workspace.slug, {
      page,

      search: search || undefined,

      status: filterValues.status,
    }),
  );

  const links = query.data?.data ?? [];

  const pagination = query.data?.pagination;

  const isEmpty = !query.isLoading && links.length === 0;

  function clearAll() {
    setSearch("");

    setFilterValues({});

    setPage(1);
  }

  return (
    <>
      <PageHeader title="Links" actions={<CreateLinkDialog />} />

      <div className="flex flex-1 flex-col gap-4 p-4">
        {/* Toolbar — search + filters */}

        <div className="flex items-center gap-2">
          <FilterSheet
            filters={LINK_FILTERS}
            values={filterValues}
            onChange={(v) => {
              setFilterValues(v);

              setPage(1);
            }}
          />

          <div className="relative ml-auto w-64">
            <Search className="pointer-events-none absolute left-2 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />

            <Input
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);

                setPage(1);
              }}
              placeholder="Search links…"
              className="pl-7"
            />

            {search && (
              <Button
                variant="ghost"
                size="icon-xs"
                className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground"
                onClick={() => setSearch("")}
              >
                <X />
              </Button>
            )}
          </div>
        </div>

        {/* Link cards or empty state */}

        {query.isLoading ? (
          <PageLoader />
        ) : isEmpty ? (
          <div className="rounded-lg border">
            {hasFilters ? (
              <LinksFilteredEmptyState onClear={clearAll} />
            ) : (
              <LinksEmptyState />
            )}
          </div>
        ) : (
          <div className="space-y-2">
            {links.map((link) => (
              <LinkCard key={link.id} link={link} />
            ))}
          </div>
        )}

        {/* Pagination */}

        {pagination && pagination.total > 0 && (
          <div className="flex items-center justify-between px-2">
            <p className="text-sm text-muted-foreground">
              Showing {(pagination.page - 1) * pagination.per_page + 1}–
              {Math.min(
                pagination.page * pagination.per_page,

                pagination.total,
              )}{" "}
              of {pagination.total}
            </p>

            <div className="flex items-center gap-1">
              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(page - 1)}
                disabled={page <= 1}
              >
                <ChevronLeft />
                Previous
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={() => setPage(page + 1)}
                disabled={page >= pagination.total_pages}
              >
                Next
                <ChevronRight />
              </Button>
            </div>
          </div>
        )}
      </div>
    </>
  );
}
