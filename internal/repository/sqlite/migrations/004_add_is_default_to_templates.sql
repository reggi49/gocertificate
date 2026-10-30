-- Add is_default column to templates table
ALTER TABLE templates ADD COLUMN is_default INTEGER NOT NULL DEFAULT 0;
