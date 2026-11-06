-- Migration 005: Add Event City / Detail Alamat (Kecamatan - Kota) to Projects

ALTER TABLE projects ADD COLUMN event_city TEXT NOT NULL DEFAULT '';
