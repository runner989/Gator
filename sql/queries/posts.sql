-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPostsForUser :many
SELECT p.*
FROM posts AS p
JOIN feed_follows AS ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1
ORDER BY
    p.published_at DESC NULLS LAST,
    p.created_at  DESC
LIMIT  $2;

-- name: GetPostsForUserPaginated :many
SELECT p.*
FROM posts AS p
JOIN feed_follows AS ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1
ORDER BY
    CASE WHEN $3::text = 'time' THEN p.published_at END DESC,
    CASE WHEN $3::text = 'title' THEN p.title END ASC,
    p.published_at DESC
LIMIT $2
OFFSET $4;
