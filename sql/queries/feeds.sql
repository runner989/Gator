-- name: AddFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6
       )
RETURNING *;

-- name: DeleteFeeds :exec
DELETE FROM feeds;

-- name: DeleteFollows :exec
DELETE FROM feed_follows;

-- name: GetFeeds :many
SELECT name, url, user_id FROM feeds;

-- name: CreateFeedFollow :one
WITH inserted_follow AS (
    INSERT INTO feed_follows (user_id, feed_id)
    VALUES ($1, $2)
    RETURNING *
)
SELECT
    inserted_follow.*,
    users.name as user_name,
    feeds.name as feed_name
FROM inserted_follow
JOIN users ON users.id = inserted_follow.user_id
JOIN feeds ON feeds.id = inserted_follow.feed_id;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.*,
       users.name as user_name,
       feeds.name as feed_name
FROM feed_follows
JOIN users ON users.id = feed_follows.user_id
JOIN feeds ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1;

-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;

-- name: UnFollow :exec
WITH deleted_follow AS (
    DELETE FROM feed_follows
    WHERE feed_follows.user_id = $1
    AND feed_follows.feed_id = $2
    RETURNING *
)
SELECT
    deleted_follow.*,
    users.name as user_name,
    feeds.name as feed_name
FROM deleted_follow
JOIN users ON users.id = deleted_follow.user_id
JOIN feeds ON feeds.id = deleted_follow.feed_id;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at = Now(),
    last_fetched_at = Now()
WHERE feeds.id = $1;

-- name: GetNextFeedToFetch :one
SELECT *
FROM feeds
ORDER BY last_fetched_at NULLS FIRST,
         updated_at LIMIT 1;
