CREATE INDEX runs_account_index ON runs (account);
CREATE INDEX runs_labels_index ON runs USING GIN (labels JSONB_PATH_OPS);
