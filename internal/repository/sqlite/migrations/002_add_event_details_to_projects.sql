-- Migration 002: Add Event Details (Tanggal, Tempat, Alamat) to Projects

ALTER TABLE projects ADD COLUMN event_date TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN event_location TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN event_address TEXT NOT NULL DEFAULT '';

