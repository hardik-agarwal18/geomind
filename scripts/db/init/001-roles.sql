-- Read credentials from the container environment.

\getenv app_db_user APP_DB_USER
\getenv app_db_password APP_DB_PASSWORD

\getenv migration_db_user MIGRATION_DB_USER
\getenv migration_db_password MIGRATION_DB_PASSWORD

\getenv database_name POSTGRES_DB

-- Migration owner.
CREATE ROLE :"migration_db_user"
WITH
    LOGIN
    PASSWORD :'migration_db_password'
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT;

-- Runtime application user.
CREATE ROLE :"app_db_user"
WITH
    LOGIN
    PASSWORD :'app_db_password'
    NOSUPERUSER
    NOCREATEDB
    NOCREATEROLE
    NOINHERIT;

-- Restrict database access.
REVOKE CONNECT, TEMP
ON DATABASE :"database_name"
FROM PUBLIC;

GRANT CONNECT
ON DATABASE :"database_name"
TO :"migration_db_user", :"app_db_user";

-- Migration role owns the public schema.
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

ALTER SCHEMA public
OWNER TO :"migration_db_user";

-- Application may use the schema but cannot create objects.
GRANT USAGE ON SCHEMA public TO :"app_db_user";

-- Future migration-created tables automatically grant
-- application-level DML privileges.
ALTER DEFAULT PRIVILEGES
FOR ROLE :"migration_db_user"
IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE
ON TABLES TO :"app_db_user";

ALTER DEFAULT PRIVILEGES
FOR ROLE :"migration_db_user"
IN SCHEMA public
GRANT USAGE, SELECT
ON SEQUENCES TO :"app_db_user";

\unset app_db_password
\unset migration_db_password
