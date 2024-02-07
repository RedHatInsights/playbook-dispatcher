CREATE INDEX idx_btree_labels ON runs USING btree((labels->>'playbook-run'));
