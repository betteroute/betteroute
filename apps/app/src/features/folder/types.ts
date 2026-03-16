// Folder represents a workspace folder for organizing links.
export interface Folder {
  id: string;
  workspace_id: string;
  created_by?: string;
  name: string;
  color: string;
  position: number;
  created_at: string;
  updated_at: string;
}
