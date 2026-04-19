-- Rule 1: each published version of a schema must be unique per tenant.
-- Prevents creating two rows with the same (tenant, schema_key, version),
-- regardless of lifecycle. A tenant can have many versions of the same schema
-- (1.0.0 active, 1.1.0 active, 2.0.0 draft) but never two rows for the
-- exact same version string.
--
-- This index was created in migration 000002; it is documented here for
-- reference alongside Rule 2 below.
--
-- CREATE UNIQUE INDEX entity_schemas_tenant_schema_uni_idx
--     ON entity_schemas (tenant_id, schema_key, schema_version);

-- Rule 2: at most one draft per schema per tenant.
-- A draft represents work-in-progress on the next version of a schema.
-- Only one change can be in flight at a time — if a new draft is needed,
-- the existing one must be deleted first. This prevents two conflicting
-- drafts (e.g. a 1.1.0 draft and a 2.0.0 draft) from existing simultaneously
-- and removes ambiguity about which one gets published.
--
-- The WHERE clause makes this a partial index so it only applies to rows
-- with lifecycle = 'draft' and has no effect on active or deprecated schemas.
CREATE UNIQUE INDEX entity_schemas_single_draft_idx
    ON entity_schemas (tenant_id, schema_key)
    WHERE lifecycle = 'draft';
