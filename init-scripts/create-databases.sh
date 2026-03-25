#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    SELECT 'CREATE DATABASE workflow_designer'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'workflow_designer')\gexec
    GRANT ALL PRIVILEGES ON DATABASE workflow_designer TO temporal;
EOSQL
