-- ── enums ────────────────────────────────────────────────────────────────────

CREATE TYPE task_status AS ENUM ('open', 'in_progress', 'done');

-- ── tasks ────────────────────────────────────────────────────────────────────

CREATE TABLE tasks (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    store_id     UUID        NOT NULL,
    title        TEXT        NOT NULL,
    status       task_status NOT NULL DEFAULT 'open',
    assignee_id  UUID,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_store_id         ON tasks (store_id);
CREATE INDEX idx_tasks_store_id_status  ON tasks (store_id, status);
CREATE INDEX idx_tasks_assignee_id      ON tasks (assignee_id) WHERE assignee_id IS NOT NULL;

-- ── outbox ───────────────────────────────────────────────────────────────────

CREATE TABLE outbox (
    id             BIGSERIAL   PRIMARY KEY,
    aggregate_id   UUID        NOT NULL,
    event_type     TEXT        NOT NULL,
    payload        JSONB       NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    process_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at   TIMESTAMPTZ,
    failed_at      TIMESTAMPTZ,
    attempts       INT         NOT NULL DEFAULT 0
);

-- Worker polls: unprocessed rows ready to deliver, oldest first.
-- FOR UPDATE SKIP LOCKED friendly.
CREATE INDEX idx_outbox_unprocessed
    ON outbox (process_at)
    WHERE processed_at IS NULL AND failed_at IS NULL;

-- ── seed data (optional, useful for smoke testing) ───────────────────────────

INSERT INTO tasks (id, store_id, title, status) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'b0000000-0000-0000-0000-000000000001', 'Set up CI pipeline',      'open'),
    ('a0000000-0000-0000-0000-000000000002', 'b0000000-0000-0000-0000-000000000001', 'Write API documentation',  'in_progress'),
    ('a0000000-0000-0000-0000-000000000003', 'b0000000-0000-0000-0000-000000000001', 'Deploy to staging',        'done'),
    ('a0000000-0000-0000-0000-000000000004', 'b0000000-0000-0000-0000-000000000002', 'Review pull requests',     'open'),
    ('a0000000-0000-0000-0000-000000000005', 'b0000000-0000-0000-0000-000000000002', 'Fix login bug',            'open');
