-- Migration: Add workflow format support
-- Description: Adds support for YAML workflows alongside JSON workflows
-- Date: 2026-01-10

-- Add workflow_format column to jobs table
ALTER TABLE jobs 
ADD COLUMN workflow_format ENUM('json', 'yaml') DEFAULT 'yaml' 
AFTER payload;

-- Add index for querying by format (optional, for analytics)
CREATE INDEX idx_jobs_workflow_format ON jobs(workflow_format);
