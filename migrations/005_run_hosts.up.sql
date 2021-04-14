CREATE TABLE run_hosts (
    id uuid PRIMARY KEY,
    run_id uuid REFERENCES runs,

    inventory_id uuid,
    host varchar NOT NULL,

    status runs_status NOT NULL default 'running',
    log text NOT NULL default '',

    events jsonb NOT NULL default '[]',

    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,

    UNIQUE (run_id, host)
);
