-- name: CreateUser :one
INSERT INTO users (username, email, password, created, updated)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, username, email, created, updated;

-- name: GetUser :one
SELECT id, username, email, created, updated
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, email, created, updated
FROM users
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: UpdateUser :one
UPDATE users
SET username = $1, email = $2, password = $3, updated = $4
WHERE id = $5
RETURNING id, username, email, created, updated;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
RETURNING id;  -- You can use RETURNING to get the id of the deleted row (optional)

-- name: CreateBlog :one
INSERT INTO blogs (title, content, user_id, created, updated)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, content, user_id, created, updated;

-- name: GetBlog :one
SELECT id, title, content, user_id, created, updated
FROM blogs
WHERE id = $1;

-- name: ListBlogs :many
SELECT id, title, content, user_id, created, updated
FROM blogs
ORDER BY id
LIMIT $1 OFFSET $2;

-- name: UpdateBlog :one
UPDATE blogs
SET title = $1, content = $2, user_id = $3, updated = $4
WHERE id = $5
RETURNING id, title, content, user_id, created, updated;

-- name: DeleteBlog :exec
DELETE FROM blogs
WHERE id = $1
RETURNING id;  -- You can use RETURNING to get the id of the deleted row (optional)

-- name: GetTotalUsersCount :one
SELECT COUNT(*) AS total FROM users;

-- name: GetUserByUsernameOrEmail :one
SELECT id, username, email, created, updated, password
FROM users
WHERE username = $1 OR email = $1;

-- name: CreateUserProfile :one
INSERT INTO user_profiles (user_id, profile_image)
VALUES ($1, $2)
RETURNING id, user_id, profile_image;

-- name: GetUserProfileByUserId :one
SELECT id, user_id, profile_image, created, updated
FROM user_profiles
WHERE user_id = $1;