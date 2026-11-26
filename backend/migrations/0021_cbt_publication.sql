-- PUB-05A: trusted, revision-scoped inputs from an approved operator or an
-- approved image workflow. No client API may insert or change these rows.
CREATE TABLE IF NOT EXISTS app.cbt_publication_inputs (
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    own_product_id uuid NOT NULL REFERENCES app.own_products(id),
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id),
    product_revision bigint NOT NULL CHECK (product_revision > 0),
    site_id text NOT NULL CHECK (site_id IN ('MLA','MLB','MLC','MLM','MCO')),
    parent_seller_id text NOT NULL,
    site_seller_id text NOT NULL,
    listing_type_id text NOT NULL CHECK (listing_type_id IN ('gold_pro','gold_special')),
    pictures_by_variant jsonb NOT NULL CHECK (jsonb_typeof(pictures_by_variant) = 'object'),
    approved_until timestamptz NOT NULL,
    PRIMARY KEY (owner_user_id, own_product_id, store_id, product_revision, site_id)
);

-- One create attempt owns the product/store scope, including its Family and
-- single-item path. Adding a site or variant needs a separate official flow.
CREATE TABLE IF NOT EXISTS app.cbt_publication_attempts (
    request_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id uuid NOT NULL REFERENCES app.users(id),
    own_product_id uuid NOT NULL REFERENCES app.own_products(id),
    store_id uuid NOT NULL REFERENCES app.meli_store_connections(id),
    product_revision bigint NOT NULL CHECK (product_revision > 0),
    site_id text NOT NULL CHECK (site_id IN ('MLA','MLB','MLC','MLM','MCO')),
    logistic_type text NOT NULL CHECK (logistic_type = 'remote'),
    parent_seller_id text NOT NULL,
    site_seller_id text NOT NULL,
    title_mode text NOT NULL CHECK (title_mode IN ('platform_auto','localized')),
    site_family_name text NOT NULL,
    payload bytea NOT NULL,
    payload_sha256 bytea NOT NULL CHECK (octet_length(payload_sha256) = 32),
    candidates jsonb,
    platform_http_status integer,
    state text NOT NULL DEFAULT 'submitting' CHECK (state IN ('submitting','pending_verify','success')),
    submitted_at timestamptz NOT NULL DEFAULT now(),
    active_request_deadline timestamptz NOT NULL DEFAULT now() + interval '25 seconds',
    completed_at timestamptz,
    UNIQUE (owner_user_id, own_product_id, store_id),
    UNIQUE (request_id, store_id)
);

-- One durable identity per stable variant. Its Siteless and site IDs share a
-- row because this slice supports exactly one site and one logistic type.
CREATE TABLE IF NOT EXISTS app.cbt_publication_identities (
    request_id uuid NOT NULL,
    store_id uuid NOT NULL,
    variant_key text NOT NULL CHECK (variant_key = 'single' OR variant_key ~ '^[0-9a-f]{32}$'),
    seller_sku text NOT NULL,
    verified_listing jsonb,
    title_verified boolean NOT NULL DEFAULT false,
    last_verified_at timestamptz,
    PRIMARY KEY (request_id, variant_key),
    UNIQUE (request_id, seller_sku),
    FOREIGN KEY (request_id, store_id) REFERENCES app.cbt_publication_attempts(request_id, store_id)
);
CREATE UNIQUE INDEX IF NOT EXISTS cbt_pub_siteless_up ON app.cbt_publication_identities
    (store_id, (verified_listing->>'SitelessUPID')) WHERE verified_listing IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS cbt_pub_site_up ON app.cbt_publication_identities
    (store_id, (verified_listing->>'SiteUPID')) WHERE verified_listing IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS cbt_pub_item ON app.cbt_publication_identities
    (store_id, (verified_listing->>'ItemID')) WHERE verified_listing IS NOT NULL;

ALTER TABLE app.cbt_publication_inputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.cbt_publication_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE app.cbt_publication_identities ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON app.cbt_publication_inputs, app.cbt_publication_attempts, app.cbt_publication_identities FROM PUBLIC;

-- Product edits already lock the product row. Reservation takes the same row
-- lock before inserting its identities, so legacy and current writers agree.
CREATE OR REPLACE FUNCTION app.cbt_keep_publication_variants() RETURNS trigger
LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM app.cbt_publication_attempts a
        JOIN app.cbt_publication_identities i USING (request_id, store_id)
        WHERE a.owner_user_id = OLD.owner_user_id AND a.own_product_id = OLD.id
        AND CASE WHEN i.variant_key = 'single' THEN
            jsonb_array_length(COALESCE(NULLIF(NEW.listing_fields->'variants', 'null'::jsonb), '[]'::jsonb)) <> 0
        ELSE NOT COALESCE((NEW.listing_fields->'variants') @>
            jsonb_build_array(jsonb_build_object('variant_id', i.variant_key)), false)
        END
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', CONSTRAINT = 'cbt_publication_variant_identity',
            MESSAGE = 'publication variant identity is reserved';
    END IF;
    RETURN NEW;
END;
$$;
CREATE OR REPLACE TRIGGER cbt_keep_publication_variants
BEFORE UPDATE OF listing_fields ON app.own_products
FOR EACH ROW WHEN (OLD.listing_fields IS DISTINCT FROM NEW.listing_fields)
EXECUTE FUNCTION app.cbt_keep_publication_variants();
