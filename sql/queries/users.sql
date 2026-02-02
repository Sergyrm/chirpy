-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid()
    , NOW()
    , NOW()
    , $1
    , $2
)
RETURNING *
;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT id
    , created_at
    , updated_at
    , email
    , hashed_password
    , is_chirpy_red
FROM users
WHERE email = $1
;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens(token, created_at, updated_at, user_id, expires_at)
VALUES 
(
    $1
    , NOW()
    , NOW()
    , $2
    , NOW() + INTERVAL '1 second' * $3
)
RETURNING *
;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = NOW()
    , revoked_at = NOW()
WHERE token = $1
;

-- name: GetUserFromRefreshToken :one
SELECT rt.user_id
FROM refresh_tokens rt
WHERE rt.token = $1
    AND rt.revoked_at IS NULL
    AND rt.expires_at > NOW()
;

-- name: UpdateUserEmailPassword :one
UPDATE users
SET email = $2
    , hashed_password = $3
    , updated_at = NOW()
WHERE id = $1
RETURNING *
;

-- name: UpdateUserChirpyRed :one
UPDATE users
SET is_chirpy_red = $2
    , updated_at = NOW()
WHERE id = $1
RETURNING *
;