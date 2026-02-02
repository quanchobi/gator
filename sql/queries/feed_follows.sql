-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3
    )
    RETURNING *
) 
SELECT inserted_feed_follow.*,
    feeds.name AS feedname,
    users.name AS username
FROM inserted_feed_follow
JOIN users
ON inserted_feed_follow.user_id = users.id
JOIN feeds
ON inserted_feed_follow.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.id,
    feed_follows.user_id,
    users.name AS username,
    feed_follows.feed_id,
    feeds.name AS feedname,
    feeds.url AS url
FROM feed_follows
JOIN feeds
ON feed_follows.feed_id = feeds.id
JOIN users
ON feed_follows.user_id = users.id
WHERE feed_follows.user_id = $1;
