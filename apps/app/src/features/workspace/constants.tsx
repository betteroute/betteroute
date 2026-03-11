import { Crown, Eye, ShieldCheck, Users } from "lucide-react";
import { ROLE_INFO, ROLES } from "./types";

export const ROLE_FILTERS = [
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
