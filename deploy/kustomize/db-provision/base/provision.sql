\set ON_ERROR_STOP on

-- Create role if missing
SELECT format('CREATE ROLE %I LOGIN PASSWORD %L;', :'app_user', :'app_pass')
WHERE NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user')
\gexec

-- Ensure password matches desired value
ALTER ROLE :"app_user" PASSWORD :'app_pass';

-- Create database if missing
SELECT format('CREATE DATABASE %I OWNER %I;', :'app_db', :'app_user')
WHERE NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = :'app_db')
\gexec
