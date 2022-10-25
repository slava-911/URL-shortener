BEGIN;

SET statement_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = ON;
SET check_function_bodies = FALSE;
SET client_min_messages = WARNING;
SET search_path = public, extensions;
SET default_tablespace = '';
SET default_with_oids = FALSE;

-- EXTENSIONS --

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- TABLES --

CREATE TABLE users
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    email         TEXT NOT NULL UNIQUE,
    password      TEXT NOT NULL
);

CREATE TABLE links
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_version  TEXT NOT NULL,
    short_version TEXT NOT NULL UNIQUE,
    description   TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'utc'),
    clicked       INT,
    user_id       UUID NOT NULL,
    CONSTRAINT user_fk FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

COMMIT;
