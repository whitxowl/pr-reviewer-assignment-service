-- +goose Up
CREATE TABLE IF NOT EXISTS teams
(
    team_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS users
(
    user_id   TEXT PRIMARY KEY CHECK (user_id ~ '^u[0-9]+$'),
    username  TEXT NOT NULL,
    team_name TEXT,
    is_active BOOLEAN DEFAULT TRUE,

    CONSTRAINT fk_user_team FOREIGN KEY (team_name) REFERENCES teams(team_name) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS pull_requests
(
    pull_request_id   TEXT PRIMARY KEY CHECK (pull_request_id ~ '^pr-[0-9]+$'),
    pull_request_name TEXT NOT NULL,
    author_id         TEXT NOT NULL,
    status            TEXT NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
    created_at        TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at         TIMESTAMP NULL,

    CONSTRAINT fk_pr_author FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers (
                                                      pull_request_id TEXT NOT NULL,
                                                      user_id         TEXT NOT NULL,

                                                      PRIMARY KEY (pull_request_id, user_id),

    CONSTRAINT fk_pr FOREIGN KEY (pull_request_id) REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    CONSTRAINT fk_reviewer FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS pull_request_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;