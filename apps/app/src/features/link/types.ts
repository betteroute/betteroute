export interface Link {
  id: string;
  workspace_id: string;
  created_by?: string;
  folder_id?: string;
  short_code: string;
  short_url: string;
  dest_url: string;
  title?: string;
  description?: string;

  // Status & Scheduling
  is_active: boolean;
  starts_at?: string;
  expires_at?: string;
  expiration_url?: string;

  // Click limits
  max_clicks?: number;

  // UTM parameters
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  utm_term?: string;
  utm_content?: string;

  // OG metadata overrides
  og_title?: string;
  og_description?: string;
  og_image?: string;

  // Analytics (denormalized)
  click_count: number;
  unique_click_count: number;
  last_clicked_at?: string;

  // Internal
  notes?: string;
  created_via: "web" | "api" | "import";

  // Timestamps
  created_at: string;
  updated_at: string;
}
