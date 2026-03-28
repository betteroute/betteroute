-- name: FindPlatformAppByID :one
SELECT * FROM platform_apps
WHERE id = $1;

-- name: ListPlatformApps :many
SELECT * FROM platform_apps
ORDER BY name;

-- name: FindPlatformAppByURLPattern :one
-- Finds the first platform app whose url_patterns array contains the given hostname.
SELECT * FROM platform_apps
WHERE url_patterns @> ARRAY[$1::TEXT]
LIMIT 1;

-- name: InsertPlatformApp :one
INSERT INTO platform_apps (
    id, name, icon_url, url_patterns,
    ios_scheme, android_scheme,
    ios_app_id, ios_bundle_id, ios_team_id,
    android_package, android_sha256
) VALUES (
    $1, $2, $3, $4,
    $5, $6,
    $7, $8, $9,
    $10, $11
) RETURNING *;
