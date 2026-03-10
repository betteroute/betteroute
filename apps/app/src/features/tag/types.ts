// Tag represents a workspace tag for categorizing links.
export interface Tag {
  id: string;
  workspace_id: string;
  created_by?: string;
  name: string;
  color: string;
  created_at: string;
  updated_at: string;
}
