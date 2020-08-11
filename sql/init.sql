--- Schema
CREATE SCHEMA IF NOT EXISTS __migrations;

--- Primary objects
CREATE TABLE IF NOT EXISTS __migrations.history
(
    id            BIGSERIAL PRIMARY KEY,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name          TEXT      NOT NULL UNIQUE,
    version       INT       NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS __migrations.state
(
    version    INT       NOT NULL UNIQUE,
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS __migrations.audit
(
    id             BIGSERIAL PRIMARY KEY,
    registered_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    action         TEXT      NOT NULL,
    migration_name TEXT      NOT NULL
);

--- Indexes
CREATE INDEX ON __migrations.audit (action);

CREATE INDEX ON __migrations.audit (migration_name);

--- Routines

--- Returns migration JSON object
CREATE OR REPLACE FUNCTION __migrations.get_object(p_id BIGINT) RETURNS TEXT AS
$$
DECLARE
    l_migration TEXT;
BEGIN
    SELECT to_json(a_1) FROM __migrations.history a_1 WHERE id = p_id INTO l_migration;

    RETURN l_migration;
END;
$$ LANGUAGE plpgsql;

--- Adds new migration
CREATE OR REPLACE FUNCTION __migrations.add(p_name TEXT) RETURNS TEXT AS
$$
DECLARE
    l_last_version INT;
    l_migration    BIGINT;
BEGIN
    SELECT version FROM __migrations.history ORDER BY version DESC LIMIT 1 INTO l_last_version;

    INSERT INTO __migrations.history (name, version)
    VALUES (p_name, coalesce(l_last_version + 1, 1))
    RETURNING id INTO l_migration;

    PERFORM __migrations.audit_add(p_name, 'C');

    RETURN __migrations.get_object(l_migration);
END;
$$ LANGUAGE plpgsql;

--- Removes a migration
CREATE OR REPLACE FUNCTION __migrations.remove(p_name TEXT) RETURNS VOID AS
$$
BEGIN
    DELETE FROM __migrations.history WHERE name = p_name;

    PERFORM __migrations.audit_add(p_name, 'R');
END;
$$ LANGUAGE plpgsql;

--- Returns migration audit JSON object
CREATE OR REPLACE FUNCTION __migrations.audit_get_object(p_id BIGINT) RETURNS TEXT AS
$$
DECLARE
    l_migration_audit TEXT;
BEGIN
    SELECT to_json(a_1) FROM __migrations.history a_1 WHERE id = p_id INTO l_migration_audit;

    RETURN l_migration_audit;
END;
$$ LANGUAGE plpgsql;

--- Adds new migration audit
CREATE OR REPLACE FUNCTION __migrations.audit_add(p_name TEXT, p_action TEXT DEFAULT 'C',
                                                  p_timestamp TIMESTAMP DEFAULT clock_timestamp()) RETURNS TEXT AS
$$
DECLARE
    l_migration_audit BIGINT;
BEGIN
    INSERT INTO __migrations.audit (action, migration_name, registered_at)
    VALUES (p_action, p_name, p_timestamp)
    RETURNING id INTO l_migration_audit;

    RETURN __migrations.audit_get_object(l_migration_audit);
END;
$$ LANGUAGE plpgsql;

--- Returns current active version
CREATE OR REPLACE FUNCTION __migrations.get_current_version() RETURNS INT AS
$$
DECLARE
    l_current_version INT;
BEGIN
    SELECT coalesce(max(version), 0) FROM __migrations.state INTO l_current_version;

    RETURN l_current_version;
END;
$$ LANGUAGE plpgsql;

--- Upgrades to specific version (by version number)
CREATE OR REPLACE FUNCTION __migrations.up(p_version INT DEFAULT NULL) RETURNS INT AS
$$
DECLARE
    l_version         INT;
    l_current_version INT;
BEGIN
    SELECT coalesce(p_version, max(version)) FROM __migrations.history INTO l_version;

    SELECT __migrations.get_current_version() INTO l_current_version;

    FOR i IN (l_current_version + 1)..l_version
        LOOP
            INSERT INTO __migrations.state (version, updated_at) VALUES (i, clock_timestamp());

            PERFORM __migrations.audit_add((SELECT name FROM __migrations.history WHERE version = i), 'U');
        END LOOP;

    RETURN __migrations.get_current_version();
END;
$$ LANGUAGE plpgsql;

--- Upgrades to specific migration (by migration name)
CREATE OR REPLACE FUNCTION __migrations.up(p_name TEXT DEFAULT NULL) RETURNS INT AS
$$
DECLARE
    l_version         INT;
    l_current_version INT;
BEGIN
    SELECT coalesce((SELECT version FROM __migrations.history WHERE name = p_name), max(version))
    FROM __migrations.history
    INTO l_version;

    SELECT __migrations.get_current_version() INTO l_current_version;

    FOR i IN (l_current_version + 1)..l_version
        LOOP
            INSERT INTO __migrations.state (version, updated_at) VALUES (i, clock_timestamp());

            PERFORM __migrations.audit_add((SELECT name FROM __migrations.history WHERE version = i), 'U');
        END LOOP;

    RETURN __migrations.get_current_version();
END;
$$ LANGUAGE plpgsql;

--- Downgrades to specific version (by version number)
CREATE OR REPLACE FUNCTION __migrations.down(p_version INT DEFAULT NULL) RETURNS INT AS
$$
DECLARE
    l_version         INT;
    l_current_version INT;
BEGIN
    SELECT coalesce(p_version, 0) INTO l_version;

    SELECT __migrations.get_current_version() INTO l_current_version;

    FOR i IN REVERSE l_current_version..(l_version + 1)
        LOOP
            DELETE FROM __migrations.state WHERE version = i;

            PERFORM __migrations.audit_add((SELECT name FROM __migrations.history WHERE version = i), 'D');
        END LOOP;

    RETURN __migrations.get_current_version();
END;
$$ LANGUAGE plpgsql;

--- Downgrade to specific migration (by migration name)
CREATE OR REPLACE FUNCTION __migrations.down(p_name TEXT DEFAULT NULL) RETURNS INT AS
$$
DECLARE
    l_version         INT;
    l_current_version INT;
BEGIN
    SELECT coalesce((SELECT version FROM __migrations.history WHERE name = p_name), 0) INTO l_version;

    SELECT __migrations.get_current_version() INTO l_current_version;

    FOR i IN REVERSE l_current_version..(l_version + 1)
        LOOP
            DELETE FROM __migrations.state WHERE version = i;

            PERFORM __migrations.audit_add((SELECT name FROM __migrations.history WHERE version = i), 'D');
        END LOOP;

    RETURN __migrations.get_current_version();
END;
$$ LANGUAGE plpgsql;

--- Returns active migrations as JSON array
CREATE OR REPLACE FUNCTION __migrations.get_active() RETURNS TEXT AS
$$
DECLARE
    l_active_migrations TEXT;
BEGIN
    SELECT json_agg(a_1)
    FROM __migrations.history a_1
             JOIN __migrations.state a_2 ON a_1.version = a_2.version
    INTO l_active_migrations;

    RETURN coalesce(l_active_migrations, json_build_array()::TEXT);
END;
$$ LANGUAGE plpgsql;

--- Returns number of active migrations
CREATE OR REPLACE FUNCTION __migrations.get_active_count() RETURNS TEXT AS
$$
DECLARE
    l_count INT;
BEGIN
    SELECT count(a_1)
    FROM __migrations.history a_1
             JOIN __migrations.state a_2 ON a_1.version = a_2.version
    INTO l_count;

    RETURN l_count;
END;
$$ LANGUAGE plpgsql;

--- Returns all migrations as JSON array
CREATE OR REPLACE FUNCTION __migrations.get_all() RETURNS TEXT AS
$$
DECLARE
    l_migrations TEXT;
BEGIN
    SELECT json_agg(a_1)
    FROM __migrations.history a_1
    INTO l_migrations;

    RETURN coalesce(l_migrations, json_build_array()::TEXT);
END;
$$ LANGUAGE plpgsql;

--- Returns migration names between specific versions (from, to]
CREATE OR REPLACE FUNCTION __migrations.get_diff(p_version_from INT, p_version_to INT) RETURNS TEXT[] AS
$$
DECLARE
    l_migrations TEXT[];
BEGIN
    SELECT array_agg(a_1.name)
    FROM __migrations.history a_1
    WHERE a_1.version BETWEEN p_version_from + 1 AND p_version_to
    GROUP BY a_1.version
    ORDER BY a_1.version
    INTO l_migrations;

    RETURN l_migrations;
END;
$$ LANGUAGE plpgsql
