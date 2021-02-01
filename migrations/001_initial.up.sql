CREATE TYPE runs_status AS ENUM('running', 'success', 'failure', 'timeout');

CREATE TABLE runs (
    id uuid PRIMARY KEY,
    account varchar(10) NOT NULL,

    recipient uuid NOT NULL,
    correlation_id uuid NOT NULL,
    url varchar NOT NULL,

    labels jsonb NOT NULL default '{}',

    status runs_status NOT NULL default 'running',

    events jsonb NOT NULL default '[]',

    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    timeout integer NOT NULL
);

