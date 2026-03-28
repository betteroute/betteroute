-- name: UpsertDeepLink :one
INSERT INTO deep_links (
    link_id, platform_app_id, workspace_app_id,
    ios_deep_link, android_deep_link,
    ios_fallback_url, android_fallback_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (link_id) DO UPDATE SET
    platform_app_id      = EXCLUDED.platform_app_id,
    workspace_app_id     = EXCLUDED.workspace_app_id,
    ios_deep_link        = EXCLUDED.ios_deep_link,
    android_deep_link    = EXCLUDED.android_deep_link,
    ios_fallback_url     = EXCLUDED.ios_fallback_url,
    android_fallback_url = EXCLUDED.android_fallback_url,
    updated_at           = NOW()
RETURNING *;

-- name: FindDeepLink :one
SELECT * FROM deep_links
WHERE link_id = $1;

-- name: DeleteDeepLink :execrows
DELETE FROM deep_links
WHERE link_id = $1;

-- name: ResolveDeepLink :one
-- Redirect path: look up deep link URLs + android package for intent:// URLs.
-- One extra round-trip after ResolveLink, only when deepview is needed.
SELECT
    dl.ios_deep_link,
    dl.android_deep_link,
    COALESCE(dl.ios_fallback_url, wa.app_store_url) AS ios_fallback_url,
    COALESCE(dl.android_fallback_url, wa.play_store_url) AS android_fallback_url,
    COALESCE(pa.android_package, wa.package_name) AS android_package
FROM deep_links dl
LEFT JOIN platform_apps pa ON pa.id = dl.platform_app_id
LEFT JOIN workspace_apps wa ON wa.id = dl.workspace_app_id
WHERE dl.link_id = $1;
