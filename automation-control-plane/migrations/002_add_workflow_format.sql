-- Migration: Add workflow format support
-- Description: Adds support for YAML workflows alongside JSON workflows
-- Date: 2026-01-10

-- Add workflow_format column to jobs table
ALTER TABLE jobs 
ADD COLUMN workflow_format ENUM('json', 'yaml') DEFAULT 'yaml' 
AFTER workflow
COMMENT 'Workflow definition format';

-- Update existing jobs to explicitly mark as JSON (optional, for clarity)
-- Uncomment if you want to explicitly mark existing workflows
-- UPDATE jobs SET workflow_format = 'json' WHERE workflow_format IS NULL;

-- Add index for querying by format (optional, for analytics)
CREATE INDEX idx_jobs_workflow_format ON jobs(workflow_format);

-- Add comment to workflow column for clarity
ALTER TABLE jobs 
MODIFY COLUMN workflow TEXT 
COMMENT 'Workflow definition (JSON or YAML format, see workflow_format column)';
