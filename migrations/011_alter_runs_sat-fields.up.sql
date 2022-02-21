ALTER TABLE runs
    ADD COLUMN playbook_name varchar,
    ADD COLUMN playbook_run_url varchar,
    ADD COLUMN principal varchar,
    ADD COLUMN sat_id uuid,
    ADD COLUMN sat_org_id varchar
