-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, last_fetched_at, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: GetFeeds :many
SELECT feeds.id, 
    feeds.created_at, 
    feeds.updated_at, 
    feeds.name, 
    feeds.url, 
    feeds.user_id, 
    feeds.last_fetched_at,
    users.name AS username
FROM feeds
JOIN users
ON feeds.user_id = users.id;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $2,
    updated_at = $3
WHERE ID = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;
