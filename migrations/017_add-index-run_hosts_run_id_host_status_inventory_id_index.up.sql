CREATE INDEX CONCURRENTLY IF NOT EXISTS run_hosts_run_id_host_status_inventory_id_index ON run_hosts (run_id) INCLUDE (host, status, inventory_id);
