CREATE TABLE IF NOT EXISTS scan_jobs (
    id UUID PRIMARY KEY,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    scan_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ NULL,
    error TEXT NOT NULL DEFAULT '',
    results INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_jobs_asset_id ON scan_jobs(asset_id);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_status ON scan_jobs(status);
CREATE INDEX IF NOT EXISTS idx_scan_jobs_scan_type ON scan_jobs(scan_type);

CREATE TABLE IF NOT EXISTS scan_results (
    id UUID PRIMARY KEY,
    scan_job_id UUID NOT NULL REFERENCES scan_jobs(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    result_type VARCHAR(50) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scan_results_scan_job_id ON scan_results(scan_job_id);
CREATE INDEX IF NOT EXISTS idx_scan_results_asset_id ON scan_results(asset_id);
CREATE INDEX IF NOT EXISTS idx_scan_results_result_type ON scan_results(result_type);