import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Crown, Eye, Search, ShieldCheck, Users, X } from "lucide-react";
import { useMemo, useState } from "react";

import {
  FilterSheet,
  type FilterValues,
} from "@/components/shared/filter-sheet";

import { PageLoader } from "@/components/shared/page-loader";

import { Button } from "@/components/ui/button";

import { Input } from "@/components/ui/input";

import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { InvitationRow } from "@/features/workspace/components/invitation-row";

import { InviteDialog } from "@/features/workspace/components/invite-dialog";

import { MemberRow } from "@/features/workspace/components/member-row";
import { useWorkspace } from "@/features/workspace/hooks";
import { workspaceQueries } from "@/features/workspace/queries";
import {
  type Invitation,
  type Member,
  ROLE_INFO,
  ROLES,
} from "@/features/workspace/types";

export const Route = createFileRoute("/_workspace/$slug/settings/members")({
  staticData: { title: "Members" },

  component: MembersSettingsPage,
});

const ROLE_FILTERS = [
  {
    key: "role",

    label: "Role",

    icon: <ShieldCheck />,

    options: ROLES.map((role) => ({
      value: role,

      label: ROLE_INFO[role]?.label ?? role,

      icon:
        role === "owner" ? (
          <Crown />
        ) : role === "admin" ? (
          <ShieldCheck />
        ) : role === "member" ? (
          <Users />
        ) : (
          <Eye />
        ),
    })),
  },
];

function MembersSettingsPage() {
  const { workspace } = useWorkspace();

  const [search, setSearch] = useState("");

  const [filters, setFilters] = useState<FilterValues>({});

  const membersQuery = useQuery(workspaceQueries.members(workspace.slug));

  const invitationsQuery = useQuery(
    workspaceQueries.invitations(workspace.slug),
  );

  const filteredMembers = useMemo(() => {
    if (!membersQuery.data) return [];

    return membersQuery.data.filter((member) => {
      const searchLower = search.toLowerCase();

      const matchesSearch =
        !search ||
        member.name?.toLowerCase().includes(searchLower) ||
        member.email?.toLowerCase().includes(searchLower);

      const selectedRoles = filters.role;

      const matchesRole =
        !selectedRoles?.length || selectedRoles.includes(member.role);

      return matchesSearch && matchesRole;
    });
  }, [membersQuery.data, search, filters]);

  const filteredInvitations = useMemo(() => {
    if (!invitationsQuery.data) return [];

    return invitationsQuery.data.filter((invitation) => {
      const searchLower = search.toLowerCase();

      const matchesSearch =
        !search || invitation.email.toLowerCase().includes(searchLower);

      const selectedRoles = filters.role;

      const matchesRole =
        !selectedRoles?.length || selectedRoles.includes(invitation.role);

      return matchesSearch && matchesRole;
    });
  }, [invitationsQuery.data, search, filters]);

  return (
    <div className="space-y-6 p-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-lg font-semibold">Members</h1>

          <p className="text-muted-foreground text-sm">
            Invite team members and manage roles.
          </p>
        </div>

        <InviteDialog />
      </div>

      <div className="flex items-center gap-2">
        <FilterSheet
          filters={ROLE_FILTERS}
          values={filters}
          onChange={setFilters}
        />

        <div className="relative ml-auto">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />

          <Input
            placeholder="Search members…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8"
          />

          {search && (
            <Button
              variant="ghost"
              size="icon-xs"
              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground"
              onClick={() => setSearch("")}
            >
              <X className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      </div>

      <section>
        {membersQuery.isLoading ? (
          <PageLoader />
        ) : filteredMembers.length ? (
          <div className="space-y-4">
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-full">Member</TableHead>

                    <TableHead className="w-[150px]">Role</TableHead>

                    <TableHead className="w-[50px]" />
                  </TableRow>
                </TableHeader>

                <TableBody>
                  {filteredMembers.map((member: Member) => (
                    <MemberRow key={member.user_id} member={member} />
                  ))}
                </TableBody>
              </Table>
            </div>

            {filteredMembers.length > 10 && (
              <div className="flex items-center justify-between px-2">
                <p className="text-sm text-muted-foreground">
                  Showing {filteredMembers.length > 0 ? 1 : 0}–
                  {filteredMembers.length} of {filteredMembers.length}
                </p>
              </div>
            )}
          </div>
        ) : (
          <div className="text-muted-foreground text-sm py-8 text-center">
            {search || filters.role?.length
              ? "No members match your filters."
              : "No members found."}
          </div>
        )}
      </section>

      {filteredInvitations.length > 0 && (
        <section className="space-y-3">
          <h2 className="text-sm font-medium">
            Pending invitations{" "}
            <span className="text-muted-foreground font-normal">
              ({filteredInvitations.length})
            </span>
          </h2>

          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-full">Email</TableHead>

                  <TableHead className="w-[150px]">Role</TableHead>

                  <TableHead className="w-[50px]" />
                </TableRow>
              </TableHeader>

              <TableBody>
                {filteredInvitations.map((invitation: Invitation) => (
                  <InvitationRow key={invitation.id} invitation={invitation} />
                ))}
              </TableBody>
            </Table>
          </div>
        </section>
      )}
    </div>
  );
}
