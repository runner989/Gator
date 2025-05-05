-- name: CreatePost :exec
INSERT INTO posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetPostsForUser :many
SELECT p.*
FROM posts AS p
   JOIN feed_follows AS ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1                 -- <-- user ID here
ORDER BY
    p.published_at DESC NULLS LAST, -- newest first
    p.created_at  DESC
LIMIT  $2;
