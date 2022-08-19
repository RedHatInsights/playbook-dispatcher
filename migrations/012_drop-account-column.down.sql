ALTER TABLE runs ADD COLUMN account VARCHAR(10);
CREATE INDEX runs_account_index ON runs (account);
