export const ROLES = ["viewer", "member", "admin", "owner"] as const;
export type WorkspaceRole = (typeof ROLES)[number];

const ROLE_RANK: Record<WorkspaceRole, number> = {
  viewer: 1,
  member: 2,
  admin: 3,
  owner: 4,
};

/** Returns true if `userRole` meets or exceeds `minRole`. */
export function hasRole(userRole: WorkspaceRole, minRole: WorkspaceRole) {
  return ROLE_RANK[userRole] >= ROLE_RANK[minRole];
}

/** Assignable roles — everything except `owner` (owner is set on creation). */
export const ASSIGNABLE_ROLES = ROLES.filter(
  (r) => r !== "owner",
) as WorkspaceRole[];

export const ROLE_INFO: Record<string, { label: string; description: string }> =
  {
    viewer: {
      label: "Viewer",
      description: "Can view links, analytics, and workspace activity.",
    },
    member: {
      label: "Member",
      description: "Can create, edit, and manage links and tags.",
    },
    admin: {
      label: "Admin",
      description: "Can manage members, billing, and workspace settings.",
    },
    owner: {
      label: "Owner",
      description: "Full control including ownership transfer and deletion.",
    },
  };

export interface Workspace {
  id: string;
  name: string;
  slug: string;
  status: string;
  plan_id: string;
  created_at: string;
  updated_at: string;
}

export interface WorkspaceWithRole extends Workspace {
  role: WorkspaceRole;
}

export interface Member {
  user_id: string;
  name: string;
  email: string;
  avatar_url?: string;
  role: WorkspaceRole;
  joined_at: string;
}

export interface Invitation {
  id: string;
  workspace_id: string;
  email: string;
  role: WorkspaceRole;
  expires_at: string;
  created_at: string;
}
