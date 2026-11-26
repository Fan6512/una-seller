ALTER TABLE app.cbt_publication_attempts
    ADD COLUMN IF NOT EXISTS platform_error_code text,
    ADD COLUMN IF NOT EXISTS recovery_audit jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS released_at timestamptz;

CREATE UNIQUE INDEX IF NOT EXISTS cbt_publication_active_product_store
    ON app.cbt_publication_attempts(owner_user_id, own_product_id, store_id)
    WHERE released_at IS NULL;
DO $$
DECLARE constraint_name text;
BEGIN
    SELECT conname INTO constraint_name FROM pg_constraint
    WHERE conrelid='app.cbt_publication_attempts'::regclass AND contype='u'
      AND pg_get_constraintdef(oid)='UNIQUE (owner_user_id, own_product_id, store_id)';
    IF FOUND THEN
        EXECUTE format('ALTER TABLE app.cbt_publication_attempts DROP CONSTRAINT %I', constraint_name);
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION app.cbt_keep_publication_variants() RETURNS trigger
LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM app.cbt_publication_attempts a
        JOIN app.cbt_publication_identities i USING (request_id, store_id)
        WHERE a.owner_user_id = OLD.owner_user_id AND a.own_product_id = OLD.id AND a.released_at IS NULL
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

-- Invoke only from the private database role under the documented evidence policy.
-- A NULL UP means release after definitive proof of no creation.
CREATE OR REPLACE FUNCTION app.cbt_recover_publication(
    target_request uuid, target_owner uuid, target_product uuid, target_store uuid,
    evidence_ref text, candidate_up text
) RETURNS void
LANGUAGE plpgsql VOLATILE SET search_path = pg_catalog AS $$
DECLARE attempt app.cbt_publication_attempts%ROWTYPE;
        recorded jsonb;
BEGIN
    IF evidence_ref IS NULL OR evidence_ref !~ '^[a-zA-Z0-9][a-zA-Z0-9._/-]{2,119}$'
       OR (candidate_up IS NOT NULL AND candidate_up !~ '^U[1-9][0-9]{0,29}$') THEN
        RAISE EXCEPTION 'invalid publication recovery evidence reference or candidate';
    END IF;
    PERFORM 1 FROM app.own_products WHERE id=target_product AND owner_user_id=target_owner FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'publication recovery product unavailable'; END IF;
    SELECT * INTO attempt FROM app.cbt_publication_attempts
    WHERE request_id=target_request AND owner_user_id=target_owner AND own_product_id=target_product
      AND store_id=target_store AND released_at IS NULL AND state IN ('pending_verify','submitting')
      AND active_request_deadline<=clock_timestamp()
    FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'publication recovery attempt unavailable'; END IF;

    recorded := COALESCE(NULLIF(attempt.candidates->'Candidates','null'::jsonb),'[]'::jsonb);
    IF candidate_up IS NULL THEN
        IF jsonb_array_length(recorded)<>0 OR EXISTS (
            SELECT 1 FROM app.cbt_publication_identities
            WHERE request_id=target_request AND verified_listing IS NOT NULL
        ) THEN
            RAISE EXCEPTION 'publication attempt has evidence of creation';
        END IF;
        -- Invalidate the old approval and every confirmation using the previous revision.
        UPDATE app.own_products SET revision=revision+1,updated_at=now()
        WHERE id=target_product AND owner_user_id=target_owner;
    ELSE
        IF recorded @> jsonb_build_array(jsonb_build_object('SitelessUPID',candidate_up)) OR
            jsonb_array_length(recorded)>=(SELECT count(*) FROM app.cbt_publication_identities WHERE request_id=target_request) THEN
            RAISE EXCEPTION 'publication candidate is already known or exceeds the reserved variant count';
        END IF;
        recorded := recorded || jsonb_build_array(jsonb_build_object('SitelessUPID',candidate_up));
    END IF;
    UPDATE app.cbt_publication_attempts
    SET state='pending_verify', completed_at=COALESCE(completed_at,now()),
        released_at=CASE WHEN candidate_up IS NULL THEN now() END,
        candidates=CASE WHEN candidate_up IS NULL THEN candidates
            ELSE jsonb_build_object('Candidates',recorded,'Incomplete',true) END,
        recovery_audit=recovery_audit || jsonb_build_array(jsonb_build_object(
            'at',now(),'operator',session_user,'evidence_ref',evidence_ref,
            'action',CASE WHEN candidate_up IS NULL THEN 'confirmed_not_created' ELSE 'add_candidate' END,
            'siteless_up_id',candidate_up))
    WHERE request_id=target_request;
END;
$$;
REVOKE ALL ON FUNCTION app.cbt_recover_publication(uuid,uuid,uuid,uuid,text,text) FROM PUBLIC;
