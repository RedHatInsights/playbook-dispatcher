CREATE INDEX runs_btree_labels_playbook_run_index ON runs USING btree((labels->>'playbook-run'));
