-- 0000_init.up.sql
-- Phase 1.1 / Module 1: Authentication (foundation)
-- Creates the schema_migrations bookkeeping table, the btree_gist extension
-- (needed later for tax_rates and fiscal_years daterange constraints), and the
-- reusable set_updated_at() trigger function used by every table with
-- `updated_at TIMESTAMPTZ NOT NULL`.
--
-- Idempotent: extensions and the migrations table are guarded with IF NOT EXISTS.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT PRIMARY KEY,
    name       TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;
