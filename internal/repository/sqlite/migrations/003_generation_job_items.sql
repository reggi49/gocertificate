-- Migration 002: Generation Job Items and Job Indexes

CREATE TABLE IF NOT EXISTS generation_job_items (
    id TEXT PRIMARY KEY,
    job_id TEXT NOT NULL,
    recipient_id TEXT NOT NULL,
    recipient_name TEXT NOT NULL,
    certificate_id TEXT,
    output_filename TEXT,
    status TEXT NOT NULL,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES generation_jobs(id) ON DELETE CASCADE,
    FOREIGN KEY (recipient_id) REFERENCES recipients(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_job_items_job_id ON generation_job_items(job_id);
CREATE INDEX IF NOT EXISTS idx_job_items_status ON generation_job_items(job_id, status);
CREATE INDEX IF NOT EXISTS idx_generation_jobs_project ON generation_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_generation_jobs_status ON generation_jobs(status);
