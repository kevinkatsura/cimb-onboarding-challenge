-- ==========================================
-- DATABASE INITIALIZATION
-- ==========================================

DROP DATABASE IF EXISTS go_db_exercise;

CREATE DATABASE go_db_exercise;

\c go_db_exercise;

-- ==========================================
-- SCHEMA
-- ==========================================

CREATE SCHEMA IF NOT EXISTS app;

SET search_path TO app;

