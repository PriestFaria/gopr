CREATE TABLE users
(
    id         TEXT PRIMARY KEY,
    username   TEXT        NOT NULL,
    team_id    TEXT,
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_team ON users (team_id);

CREATE TABLE team
(
    id   TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);


ALTER TABLE users
    ADD CONSTRAINT fk_users_team
        FOREIGN KEY (team_id)
            REFERENCES team (id)
            ON DELETE SET NULL;



CREATE TABLE pull_requests
(
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    author_id  TEXT NOT NULL,
    status     TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    merged_at  TIMESTAMPTZ,

    CONSTRAINT fk_pr_author
        FOREIGN KEY (author_id)
            REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_pr_author ON pull_requests (author_id);
CREATE INDEX idx_pr_status ON pull_requests (status);



CREATE TABLE pull_request_reviewer
(
    pull_request_id TEXT NOT NULL,
    reviewer_id     TEXT NOT NULL,

    PRIMARY KEY (pull_request_id, reviewer_id),

    CONSTRAINT fk_pr_reviewers_pr
        FOREIGN KEY (pull_request_id)
            REFERENCES pull_requests (id)
            ON DELETE CASCADE,

    CONSTRAINT fk_pr_reviewers_user
        FOREIGN KEY (reviewer_id)
            REFERENCES users (id)
            ON DELETE CASCADE
);

CREATE INDEX idx_pr_reviewer ON pull_request_reviewer (reviewer_id);
