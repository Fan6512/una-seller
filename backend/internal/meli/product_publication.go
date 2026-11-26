package meli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/publication"
	"github.com/jackc/pgx/v5"
)

var (
	ErrPublicationConflict = errors.New("saved product or publication target changed")
	ErrPublicationNotReady = errors.New("publication target, capacity or verified pictures are unavailable")
	ErrPublicationReadback = errors.New("publication outcome is unknown; check the existing attempt")
)

type PublicationInputError struct{ error }

type PublicationConfirmation struct {
	StoreID   string                `json:"store_id"`
	Site      string                `json:"site_id"`
	Revision  int64                 `json:"expected_revision"`
	TitleMode publication.TitleMode `json:"title_mode"`
}

type PublicationTarget struct {
	StoreID     string `json:"store_id"`
	StoreName   string `json:"store_name"`
	Site        string `json:"site_id"`
	ListingType string `json:"listing_type_id,omitempty"`
	State       string `json:"state,omitempty"`
}

type PublicationResult struct {
	RequestID string                       `json:"request_id"`
	Site      string                       `json:"site_id"`
	State     string                       `json:"state"`
	Variants  []publication.MatchedVariant `json:"variants"`
	Pending   []string                     `json:"pending"`
}

func (s *Service) PublicationTargets(ctx context.Context, owner string, p collections.OwnProduct) ([]PublicationTarget, error) {
	rows, err := s.db.Query(ctx, `
 SELECT i.store_id::text,c.nickname,i.site_id,i.listing_type_id,'' FROM app.cbt_publication_inputs i
 JOIN app.meli_store_connections c ON c.id=i.store_id AND c.owner_user_id=i.owner_user_id
 WHERE i.owner_user_id=$1::uuid AND i.own_product_id=$2::uuid AND i.product_revision=$3
 AND i.approved_until>now() AND c.seller_id=i.parent_seller_id AND c.status='connected' AND c.marketplaces_status='verified'
 AND NOT EXISTS (SELECT 1 FROM app.cbt_publication_attempts a
 WHERE a.owner_user_id=i.owner_user_id AND a.own_product_id=i.own_product_id AND a.store_id=i.store_id AND a.released_at IS NULL)
 UNION ALL
 SELECT a.store_id::text,c.nickname,a.site_id,'',CASE WHEN a.state='submitting'
 AND a.active_request_deadline<=now() THEN 'pending_verify' ELSE a.state END
 FROM app.cbt_publication_attempts a JOIN app.meli_store_connections c ON c.id=a.store_id AND c.owner_user_id=a.owner_user_id
 WHERE a.owner_user_id=$1::uuid AND a.own_product_id=$2::uuid AND a.released_at IS NULL ORDER BY 2,3,1`, owner, p.ID, p.Revision)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	targets := []PublicationTarget{}
	for rows.Next() {
		var t PublicationTarget
		if err := rows.Scan(&t.StoreID, &t.StoreName, &t.Site, &t.ListingType, &t.State); err != nil {
			return nil, err
		}
		targets = append(targets, t)
	}
	return targets, rows.Err()
}

type publicationAttempt struct {
	id, state string
	scope     publication.ReadbackScope
	prepared  publication.Prepared
	hints     publication.CreateHints
}

func (s *Service) publicationAttempt(ctx context.Context, owner, product, store string) (publicationAttempt, error) {
	var a publicationAttempt
	var hints []byte
	err := s.db.QueryRow(ctx, `SELECT request_id::text,
 CASE WHEN state='submitting' AND active_request_deadline<=now() THEN 'pending_verify' ELSE state END,
 parent_seller_id,site_seller_id,site_id,logistic_type,title_mode,site_family_name,candidates
 FROM app.cbt_publication_attempts WHERE owner_user_id=$1::uuid AND own_product_id=$2::uuid AND store_id=$3::uuid AND released_at IS NULL`,
		owner, product, store).Scan(&a.id, &a.state, &a.scope.ParentSellerID, &a.scope.SiteSellerID, &a.scope.Site,
		&a.scope.Logistic, &a.prepared.Copy.Mode, &a.prepared.Copy.SiteFamilyName, &hints)
	if err != nil {
		return a, err
	}
	if hints != nil {
		if err = json.Unmarshal(hints, &a.hints); err != nil {
			return a, err
		}
	}
	a.prepared.Target = publication.Target{Site: a.scope.Site, Logistic: a.scope.Logistic}
	return a, nil
}

func (s *Service) publicationResult(ctx context.Context, owner string, a publicationAttempt) (PublicationResult, error) {
	result := PublicationResult{RequestID: a.id, Site: a.scope.Site, State: a.state,
		Variants: []publication.MatchedVariant{}, Pending: []string{}}
	rows, err := s.db.Query(ctx, `SELECT i.variant_key,i.verified_listing,i.title_verified
 FROM app.cbt_publication_identities i JOIN app.cbt_publication_attempts a USING (request_id,store_id)
 WHERE a.request_id=$1::uuid AND a.owner_user_id=$2::uuid ORDER BY i.variant_key`, a.id, owner)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var v publication.MatchedVariant
		var raw []byte
		if err = rows.Scan(&v.Key, &raw, &v.TitleVerified); err != nil {
			return result, err
		}
		if raw == nil {
			result.Pending = append(result.Pending, v.Key)
			continue
		}
		if err = json.Unmarshal(raw, &v.Listing); err != nil {
			return result, err
		}
		result.Variants = append(result.Variants, v)
	}
	if a.state != "submitting" && len(result.Pending) == 0 {
		result.State = "success"
	}
	return result, rows.Err()
}

type publicationPictures map[string][]string

type publicationApproval struct {
	parent, child, listingType string
	pictures                   publicationPictures
}

func (s *Service) SubmitPublication(ctx context.Context, owner string, p collections.OwnProduct, in PublicationConfirmation) (PublicationResult, error) {
	if !validUUID(in.StoreID) || in.Revision < 1 || (in.TitleMode != publication.PlatformAuto && in.TitleMode != publication.Localized) {
		return PublicationResult{}, PublicationInputError{errors.New("invalid publication confirmation")}
	}
	a, err := s.publicationAttempt(ctx, owner, p.ID, in.StoreID)
	if err == nil {
		if a.scope.Site != in.Site {
			return PublicationResult{}, ErrPublicationConflict
		}
		return s.publicationResult(ctx, owner, a)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return PublicationResult{}, err
	}
	if p.Revision != in.Revision || (p.ListingFields.Variants == nil && len(p.Variants) != 0) {
		return PublicationResult{}, ErrPublicationConflict
	}

	preflight, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var approval publicationApproval
	var pictures []byte
	err = s.db.QueryRow(preflight, `SELECT parent_seller_id,site_seller_id,listing_type_id,pictures_by_variant
 FROM app.cbt_publication_inputs WHERE owner_user_id=$1::uuid AND own_product_id=$2::uuid AND store_id=$3::uuid
 AND product_revision=$4 AND site_id=$5 AND approved_until>now()`, owner, p.ID, in.StoreID, in.Revision, in.Site).
		Scan(&approval.parent, &approval.child, &approval.listingType, &pictures)
	if errors.Is(err, pgx.ErrNoRows) {
		return PublicationResult{}, ErrPublicationNotReady
	}
	if err != nil {
		return PublicationResult{}, err
	}
	if err = json.Unmarshal(pictures, &approval.pictures); err != nil {
		return PublicationResult{}, err
	}

	const remote = "remote"
	auth, child, err := s.RequireSiteConnection(preflight, owner, in.StoreID, in.Site, remote)
	if err != nil {
		return PublicationResult{}, err
	}
	if auth.SellerID != approval.parent || child != approval.child {
		return PublicationResult{}, ErrPublicationNotReady
	}
	scope := publication.ReadbackScope{ParentSellerID: auth.SellerID, SiteSellerID: child, Site: in.Site, Logistic: remote}
	evidence, err := s.collectPublicationEvidence(preflight, auth)
	if err != nil {
		return PublicationResult{}, ErrPublicationNotReady
	}
	if evidence.Mode != "standard_up" || evidence.SellAllowed != "enabled" ||
		!slices.ContainsFunc(evidence.Sites, func(site PublicationSiteEvidence) bool {
			return site.SiteID == in.Site && site.LogisticType == remote && site.PricingModel == "net_proceeds" && site.UserProduct == "enabled"
		}) {
		return PublicationResult{}, ErrPublicationNotReady
	}

	prepared, err := s.prepareProductPublication(preflight, auth.AccessToken, p.ListingFields, in, approval)
	if err != nil {
		return PublicationResult{}, err
	}
	if !s.publicationCapacity(preflight, auth.AccessToken, scope, len(prepared.Variants)) {
		return PublicationResult{}, ErrPublicationNotReady
	}

	a, err = s.reservePublication(ctx, owner, p.ID, in, auth, scope, approval, pictures, prepared)
	if errors.Is(err, ErrPublicationConflict) {
		existing, readErr := s.publicationAttempt(ctx, owner, p.ID, in.StoreID)
		if readErr == nil && existing.scope.Site == in.Site {
			return s.publicationResult(ctx, owner, existing)
		}
	}
	if err != nil {
		return PublicationResult{}, err
	}

	// The only creation call. The durable reservation survives cancellation,
	// process exit and any response that could have created remote identities.
	status, platformCode, hints := s.createProductPublication(ctx, auth.AccessToken, a)
	finish, cancelFinish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancelFinish()
	encoded, err := json.Marshal(hints)
	if err != nil {
		return PublicationResult{}, ErrPublicationReadback
	}
	tag, err := s.db.Exec(finish, `UPDATE app.cbt_publication_attempts
 SET candidates=$3::jsonb, platform_http_status=NULLIF($4,0),platform_error_code=NULLIF($5,''),state='pending_verify',completed_at=now()
 WHERE request_id=$1::uuid AND owner_user_id=$2::uuid AND state IN ('submitting','pending_verify') AND candidates IS NULL AND released_at IS NULL`, a.id, owner, string(encoded), status, platformCode)
	if err != nil || tag.RowsAffected() != 1 {
		return PublicationResult{}, ErrPublicationReadback
	}
	a.state, a.hints = "pending_verify", hints
	return s.readProductPublication(ctx, owner, in.StoreID, a, auth)
}

func (s *Service) reservePublication(ctx context.Context, owner, product string, in PublicationConfirmation, auth AuthorizedStore,
	scope publication.ReadbackScope, approval publicationApproval, pictures []byte, p publication.Prepared) (publicationAttempt, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return publicationAttempt{}, err
	}
	defer tx.Rollback(ctx)
	var revision int64
	err = tx.QueryRow(ctx, `SELECT revision FROM app.own_products
 WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`, product, owner).Scan(&revision)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && revision != in.Revision) {
		return publicationAttempt{}, ErrPublicationConflict
	}
	if err != nil {
		return publicationAttempt{}, err
	}

	digest := p.PayloadSHA256()
	a := publicationAttempt{state: "submitting", scope: scope, prepared: p}
	err = tx.QueryRow(ctx, `WITH approved AS (
 SELECT c.id FROM app.meli_store_connections c JOIN app.cbt_publication_inputs i ON i.store_id=c.id AND i.owner_user_id=c.owner_user_id
 WHERE c.id=$3::uuid AND c.owner_user_id=$1::uuid AND c.seller_id=$7 AND c.status='connected'
 AND c.access_token_ciphertext=$13 AND c.marketplaces_status='verified' AND c.marketplaces=$14::jsonb
 AND i.own_product_id=$2::uuid AND i.product_revision=$4 AND i.site_id=$5 AND i.parent_seller_id=$7 AND i.site_seller_id=$8
 AND i.listing_type_id=$15 AND i.pictures_by_variant=$16::jsonb AND i.approved_until>clock_timestamp() FOR UPDATE OF c,i
 )
 INSERT INTO app.cbt_publication_attempts (owner_user_id,own_product_id,store_id,product_revision,site_id,logistic_type,
 parent_seller_id,site_seller_id,title_mode,site_family_name,payload,payload_sha256)
 SELECT $1,$2,id,$4,$5,$6,$7,$8,$9,$10,$11,decode($12,'hex') FROM approved
 ON CONFLICT (owner_user_id,own_product_id,store_id) WHERE released_at IS NULL DO NOTHING RETURNING request_id::text`,
		owner, product, in.StoreID, in.Revision, in.Site, scope.Logistic, scope.ParentSellerID, scope.SiteSellerID,
		in.TitleMode, p.Copy.SiteFamilyName, p.Body, digest, auth.accessCipher, string(auth.marketsJSON), approval.listingType, string(pictures)).Scan(&a.id)
	if errors.Is(err, pgx.ErrNoRows) {
		return publicationAttempt{}, ErrPublicationConflict
	}
	if err != nil {
		return publicationAttempt{}, err
	}
	batch := &pgx.Batch{}
	for _, v := range p.Variants {
		batch.Queue(`INSERT INTO app.cbt_publication_identities(request_id,store_id,variant_key,seller_sku)
 VALUES($1::uuid,$2::uuid,$3,$4)`, a.id, in.StoreID, v.Key, v.SKU)
	}
	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return publicationAttempt{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return publicationAttempt{}, ErrPublicationReadback
	}
	return a, nil
}

func (s *Service) publicationCapacity(ctx context.Context, token string, scope publication.ReadbackScope, count int) bool {
	var caps []struct {
		Seller   json.Number `json:"user_id"`
		Site     string      `json:"site_id"`
		Logistic string      `json:"logistic_type"`
		Quota    *int64      `json:"quota"`
		Total    *int64      `json:"total_items"`
	}
	if status, err := s.listingGET(ctx, token, "/marketplace/users/cap", &caps, nil); err != nil || status != http.StatusOK {
		return false
	}
	matches, allowed := 0, false
	for _, c := range caps {
		if c.Site == scope.Site && c.Logistic == scope.Logistic && c.Seller.String() == scope.SiteSellerID {
			matches++
			allowed = c.Quota != nil && c.Total != nil && *c.Total >= 0 && (*c.Quota == 0 || (*c.Quota > 0 && *c.Quota-*c.Total >= int64(count)))
		}
	}
	return matches == 1 && allowed
}

func (s *Service) createProductPublication(ctx context.Context, token string, a publicationAttempt) (int, string, publication.CreateHints) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	incomplete := publication.CreateHints{Incomplete: true}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+a.prepared.Path, bytes.NewReader(a.prepared.Body))
	if err != nil {
		return 0, "", incomplete
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Request-ID", a.id)
	// No redirects or transport replay of the create body.
	req.GetBody = nil
	client := *s.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", incomplete
	}
	defer resp.Body.Close()
	const maxBody = 2 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(body) > maxBody {
		return resp.StatusCode, "", incomplete
	}
	var rejected struct {
		Code string `json:"error"`
	}
	if resp.StatusCode >= 400 && json.Unmarshal(body, &rejected) == nil {
		switch rejected.Code {
		case "validation_error", "bad_request", "invalid_token", "forbidden", "user.quota_exceeded", "item.max_limit_publish":
		default:
			rejected.Code = ""
		}
	}
	hints, err := publication.ExtractCreateHints(a.prepared, a.scope, body)
	if err != nil {
		return resp.StatusCode, rejected.Code, incomplete
	}
	return resp.StatusCode, rejected.Code, hints
}

type publicationAttributeRule struct {
	ID        string          `json:"id"`
	Hierarchy string          `json:"hierarchy"`
	Type      string          `json:"value_type"`
	MaxLength int             `json:"value_max_length"`
	Tags      map[string]bool `json:"tags"`
	Values    []struct {
		ID   json.Number `json:"id"`
		Name string      `json:"name"`
	} `json:"values"`
	Units []struct {
		ID string `json:"id"`
	} `json:"allowed_units"`
	DefaultUnit string `json:"default_unit"`
}

var publicationDecimal = regexp.MustCompile(`^-?[0-9]+(?:\.[0-9]+)?$`)
var publicationNumericID = regexp.MustCompile(`^[1-9][0-9]{0,29}$`)

func (s *Service) prepareProductPublication(ctx context.Context, token string, f collections.ListingFields,
	in PublicationConfirmation, approval publicationApproval) (publication.Prepared, error) {
	mode := publication.Target{Site: in.Site, Logistic: "remote", Mode: "standard_up", Pricing: "net_proceeds", ListingType: approval.listingType}
	var desc collections.LocalizedDescriptions
	if f.Descriptions != nil {
		desc = *f.Descriptions
	}
	copy, err := publication.SelectCopy(in.Site, in.TitleMode, publication.LocalizedText(f.Titles), publication.LocalizedText(desc))
	if err != nil {
		return publication.Prepared{}, PublicationInputError{err}
	}
	if !cbtCategoryID.MatchString(f.CategoryID) {
		return publication.Prepared{}, PublicationInputError{ErrCategoryInvalid}
	}
	var metadata []publicationAttributeRule
	status, err := s.listingGET(ctx, token, "/categories/"+f.CategoryID+"/attributes", &metadata, nil)
	if err != nil || status != http.StatusOK || len(metadata) == 0 {
		return publication.Prepared{}, ErrPublicationNotReady
	}
	category, status := s.readCBTTechnicalRequirements(ctx, token, CategoryRules{CategoryID: f.CategoryID})
	if status != http.StatusOK || !category.TechnicalSpecsVerified || category.SpecialRequirements {
		return publication.Prepared{}, ErrPublicationNotReady
	}
	input := publication.Input{Copy: copy, Target: mode, CategoryID: f.CategoryID}
	fields := make(map[string]publicationAttributeRule, len(metadata))
	technical := make(map[string]bool, len(category.TechnicalRequirements))
	for _, key := range category.TechnicalRequirements {
		technical[key] = true
	}
	for _, rule := range metadata {
		required := technical[rule.ID] || rule.Tags["required"] || rule.Tags["new_required"]
		delete(technical, rule.ID)
		if rule.Tags["conditional_required"] || rule.Tags["size_grid_required"] {
			return publication.Prepared{}, ErrPublicationNotReady
		}
		if _, duplicate := fields[rule.ID]; duplicate {
			return publication.Prepared{}, ErrPublicationNotReady
		}
		fields[rule.ID] = rule
		input.Rules = append(input.Rules, publication.Rule{ID: rule.ID, Hierarchy: rule.Hierarchy, Required: required,
			Enumerated: rule.Type == "list" || rule.Type == "boolean"})
	}
	if len(technical) != 0 {
		return publication.Prepared{}, ErrPublicationNotReady
	}
	makeVariant := func(key, sku string, quantity *int, net *string, child *[]collections.ListingCategoryAttribute) (publication.Variant, error) {
		var v publication.Variant
		if net == nil || len(approval.pictures[key]) == 0 {
			return v, ErrPublicationNotReady
		}
		v = publication.Variant{Key: key, SKU: sku, Quantity: quantity, NetProceeds: *net, PictureIDs: approval.pictures[key]}
		for index, attributes := range []*[]collections.ListingCategoryAttribute{f.CategoryAttributes, child} {
			if attributes == nil {
				continue
			}
			for _, attribute := range *attributes {
				if attribute.CategoryID != f.CategoryID {
					continue
				}
				rule, ok := fields[attribute.ID]
				if ok && index == 0 && f.Variants != nil && len(*f.Variants) != 0 && rule.Hierarchy == "CHILD_PK" {
					continue
				}
				if !ok || (index == 1 && rule.Hierarchy != "CHILD_PK") {
					return v, PublicationInputError{fmt.Errorf("attribute %s is not valid for the saved category or variant", attribute.ID)}
				}
				value, err := rule.publicationValue(attribute)
				if err != nil {
					return v, PublicationInputError{err}
				}
				v.Attributes = append(v.Attributes, publication.Attribute{ID: rule.ID, Values: []publication.Value{value}})
			}
		}
		return v, nil
	}
	if f.Variants == nil || len(*f.Variants) == 0 {
		v, err := makeVariant(publication.SingleKey, f.SellerSKU, f.AvailableQuantity, f.TargetNetProceedsUSD, nil)
		if err != nil {
			return publication.Prepared{}, err
		}
		input.Variants = append(input.Variants, v)
	} else {
		if f.VariantIdentityVersion != 1 {
			return publication.Prepared{}, ErrPublicationConflict
		}
		for _, entry := range *f.Variants {
			v, err := makeVariant(entry.VariantID, entry.SellerSKU, entry.AvailableQuantity, entry.TargetNetProceedsUSD, entry.CategoryAttributes)
			if err != nil {
				return publication.Prepared{}, err
			}
			input.Variants = append(input.Variants, v)
		}
	}
	if len(approval.pictures) != len(input.Variants) {
		return publication.Prepared{}, ErrPublicationNotReady
	}
	prepared, err := publication.Prepare(input)
	if err != nil {
		return publication.Prepared{}, PublicationInputError{err}
	}
	return prepared, nil
}

func (rule publicationAttributeRule) publicationValue(field collections.ListingCategoryAttribute) (publication.Value, error) {
	invalid := fmt.Errorf("attribute %s requires a supported official value and unit", rule.ID)
	if rule.Hierarchy == "CHILD_DEPENDENT" || rule.Tags["read_only"] || rule.Tags["hidden"] || rule.Tags["fixed"] || rule.Tags["multivalued"] {
		return publication.Value{}, invalid
	}
	value := strings.Join(strings.Fields(field.Value), " ")
	switch rule.Type {
	case "string", "list", "boolean", "number", "number_unit":
	default:
		return publication.Value{}, invalid
	}
	if field.ValueID != "" {
		if field.UnitID != "" || rule.Type == "number" || rule.Type == "number_unit" {
			return publication.Value{}, invalid
		}
		for _, official := range rule.Values {
			if official.ID.String() == field.ValueID && strings.TrimSpace(official.Name) != "" {
				return publication.Value{ID: field.ValueID, Name: official.Name}, nil
			}
		}
		return publication.Value{}, invalid
	}
	if value == "" || (rule.MaxLength > 0 && utf8.RuneCountInString(value) > rule.MaxLength) ||
		rule.Type == "list" || rule.Type == "boolean" || (rule.Type != "number_unit" && field.UnitID != "") {
		return publication.Value{}, invalid
	}
	if rule.Type == "number" || rule.Type == "number_unit" {
		if !publicationDecimal.MatchString(value) {
			return publication.Value{}, invalid
		}
		integer, fraction, _ := strings.Cut(strings.TrimPrefix(value, "-"), ".")
		integer = strings.TrimLeft(integer, "0")
		if integer == "" {
			integer = "0"
		}
		valueSign := strings.HasPrefix(value, "-")
		value = integer
		if fraction = strings.TrimRight(fraction, "0"); fraction != "" {
			value += "." + fraction
		}
		if valueSign && value != "0" {
			value = "-" + value
		}
	}
	if rule.Type == "number_unit" {
		if !slices.ContainsFunc(rule.Units, func(unit struct {
			ID string `json:"id"`
		}) bool {
			return unit.ID == field.UnitID
		}) || field.UnitID == "" {
			return publication.Value{}, invalid
		}
		// Use one official unit for group keys where several are available.
		if len(rule.Units) > 1 && (rule.Hierarchy == "PARENT_PK" || rule.Hierarchy == "CHILD_PK") && field.UnitID != rule.DefaultUnit {
			return publication.Value{}, invalid
		}
		value += " " + field.UnitID
	}
	return publication.Value{Name: value}, nil
}

func (s *Service) VerifyPublication(ctx context.Context, owner, product, store, site string) (PublicationResult, error) {
	if !validUUID(store) {
		return PublicationResult{}, ErrPublicationConflict
	}
	a, err := s.publicationAttempt(ctx, owner, product, store)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && a.scope.Site != site) {
		return PublicationResult{}, ErrPublicationConflict
	}
	if err != nil {
		return PublicationResult{}, err
	}
	if a.state == "pending_verify" && len(a.hints.Candidates) == 0 {
		if _, err := s.db.Exec(ctx, `UPDATE app.cbt_publication_attempts SET state='pending_verify',completed_at=now()
 WHERE request_id=$1::uuid AND owner_user_id=$2::uuid AND state='submitting' AND active_request_deadline<=now()`,
			a.id, owner); err != nil {
			return PublicationResult{}, err
		}
	}
	if a.state == "submitting" || len(a.hints.Candidates) == 0 {
		return s.publicationResult(ctx, owner, a)
	}
	if a.state == "success" {
		result, err := s.publicationResult(ctx, owner, a)
		if err != nil {
			return result, err
		}
		if !slices.ContainsFunc(result.Variants, func(v publication.MatchedVariant) bool { return !v.TitleVerified }) {
			return result, nil
		}
	}
	auth, child, err := s.RequireSiteConnection(ctx, owner, store, site, a.scope.Logistic)
	if err != nil {
		return PublicationResult{}, err
	}
	if auth.SellerID != a.scope.ParentSellerID || child != a.scope.SiteSellerID {
		return PublicationResult{}, ErrPublicationConflict
	}
	return s.readProductPublication(ctx, owner, store, a, auth)
}

func (s *Service) readProductPublication(ctx context.Context, owner, store string, a publicationAttempt, auth AuthorizedStore) (PublicationResult, error) {
	if len(a.hints.Candidates) == 0 {
		finish, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		return s.publicationResult(finish, owner, a)
	}
	read, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	var verified []publication.VerifiedListing
	for _, hint := range a.hints.Candidates {
		if read.Err() != nil {
			break
		}
		if listing, ok := s.verifyProductCandidate(read, auth.AccessToken, a.scope, hint); ok {
			verified = append(verified, listing)
		}
	}
	finish, cancelFinish := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancelFinish()
	if len(verified) > 0 {
		if err := s.savePublicationReadback(finish, owner, store, a, auth, verified); err != nil {
			return PublicationResult{}, ErrPublicationReadback
		}
	}
	return s.publicationResult(finish, owner, a)
}

type publicationUP struct {
	ID         string                  `json:"id"`
	Seller     json.Number             `json:"user_id"`
	Family     json.Number             `json:"family_id"`
	FamilyName string                  `json:"family_name"`
	Attributes []publication.Attribute `json:"attributes"`
}

func (s *Service) verifyProductCandidate(ctx context.Context, token string, scope publication.ReadbackScope, h publication.CreateCandidate) (publication.VerifiedListing, bool) {
	var zero publication.VerifiedListing
	get := func(path string, result any) bool {
		status, err := s.listingGET(ctx, token, path, result, nil)
		return err == nil && status == http.StatusOK
	}
	var up publicationUP
	if !get("/user-products/"+url.PathEscape(h.SitelessUPID), &up) || up.ID != h.SitelessUPID ||
		up.Seller.String() != scope.ParentSellerID || !publicationNumericID.MatchString(up.Family.String()) ||
		(h.SitelessFamilyID != "" && up.Family.String() != h.SitelessFamilyID) {
		return zero, false
	}
	sku := ""
	for _, attr := range up.Attributes {
		if attr.ID != "SELLER_SKU" {
			continue
		}
		if sku != "" || len(attr.Values) > 1 {
			return zero, false
		}
		sku = attr.ValueName
		if len(attr.Values) == 1 {
			if sku != "" && sku != attr.Values[0].Name {
				return zero, false
			}
			sku = attr.Values[0].Name
		}
		if sku == "" {
			return zero, false
		}
	}
	if sku == "" {
		return zero, false
	}

	var mapping []struct {
		Owner      json.Number `json:"owner_id"`
		Site       string      `json:"site_id"`
		SitelessUP string      `json:"siteless_user_product_id"`
		SiteItems  []struct {
			Item     string `json:"item_id"`
			Site     string `json:"site_id"`
			Logistic string `json:"logistic_type"`
		} `json:"site_items"`
	}
	if !get("/marketplace/user-products/"+url.PathEscape(up.ID)+"/mapping", &mapping) {
		return zero, false
	}
	itemID := ""
	for _, link := range mapping {
		if link.Owner.String() != scope.ParentSellerID || link.Site != "CBT" || link.SitelessUP != up.ID {
			return zero, false
		}
		for _, item := range link.SiteItems {
			if item.Site == scope.Site && item.Logistic == scope.Logistic {
				if itemID != "" || !strings.HasPrefix(item.Item, scope.Site) ||
					!publicationNumericID.MatchString(strings.TrimPrefix(item.Item, scope.Site)) {
					return zero, false
				}
				itemID = item.Item
			}
		}
	}
	if itemID == "" || (h.ItemID != "" && itemID != h.ItemID) {
		return zero, false
	}
	var item listingDetail
	if !get("/marketplace/items/"+itemID+"?include_attributes=all", &item) || item.ID != itemID ||
		item.Seller.String() != scope.SiteSellerID || item.Site != scope.Site || item.Logistic != scope.Logistic ||
		item.UserProduct == nil || (item.SitelessID != nil && *item.SitelessID != up.ID) ||
		(item.SitelessFamily != nil && item.SitelessFamily.String() != up.Family.String()) {
		return zero, false
	}
	sitePrefix := scope.Site + "U"
	siteUPID := *item.UserProduct
	if !strings.HasPrefix(siteUPID, sitePrefix) || !publicationNumericID.MatchString(strings.TrimPrefix(siteUPID, sitePrefix)) ||
		(h.SiteUPID != "" && siteUPID != h.SiteUPID) {
		return zero, false
	}
	var siteUP publicationUP
	if !get("/user-products/"+siteUPID, &siteUP) || siteUP.ID != siteUPID || siteUP.Seller.String() != scope.SiteSellerID ||
		!publicationNumericID.MatchString(siteUP.Family.String()) || (item.Family != nil && item.Family.String() != siteUP.Family.String()) {
		return zero, false
	}
	for _, member := range []struct{ site, id, seller, up string }{
		{"CBT", up.Family.String(), scope.ParentSellerID, up.ID},
		{scope.Site, siteUP.Family.String(), scope.SiteSellerID, siteUP.ID},
	} {
		var family struct {
			ID      json.Number `json:"family_id"`
			Seller  json.Number `json:"user_id"`
			Members []string    `json:"user_products_ids"`
		}
		if !get("/sites/"+member.site+"/user-products-families/"+member.id, &family) || family.ID.String() != member.id ||
			family.Seller.String() != member.seller || !slices.Contains(family.Members, member.up) {
			return zero, false
		}
	}
	return publication.VerifiedListing{Scope: scope, SKU: sku, SitelessUPID: up.ID, SitelessFamilyID: up.Family.String(), SiteUPID: siteUPID,
		SiteFamilyID: siteUP.Family.String(), ItemID: itemID, SiteFamilyName: siteUP.FamilyName}, true
}

func (s *Service) savePublicationReadback(ctx context.Context, owner, store string, a publicationAttempt, auth AuthorizedStore,
	verified []publication.VerifiedListing) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var locked string
	err = tx.QueryRow(ctx, `SELECT a.request_id::text FROM app.cbt_publication_attempts a
 JOIN app.meli_store_connections c ON c.id=a.store_id AND c.owner_user_id=a.owner_user_id
 WHERE a.request_id=$1::uuid AND a.owner_user_id=$2::uuid AND a.store_id=$3::uuid AND a.state IN ('pending_verify','success')
 AND a.released_at IS NULL AND a.candidates IS NOT NULL AND c.seller_id=a.parent_seller_id AND c.status='connected'
 AND c.marketplaces_status='verified' AND c.marketplaces=$4::jsonb AND c.access_token_ciphertext=$5 FOR UPDATE OF a,c`,
		a.id, owner, store, string(auth.marketsJSON), auth.accessCipher).Scan(&locked)
	if err != nil {
		return err
	}
	rows, err := tx.Query(ctx, `SELECT variant_key,seller_sku,verified_listing FROM app.cbt_publication_identities
 WHERE request_id=$1::uuid ORDER BY variant_key`, a.id)
	if err != nil {
		return err
	}
	a.prepared.Variants = nil
	var prior []publication.VerifiedListing
	for rows.Next() {
		var v publication.ExpectedVariant
		var raw []byte
		if err = rows.Scan(&v.Key, &v.SKU, &raw); err != nil {
			rows.Close()
			return err
		}
		a.prepared.Variants = append(a.prepared.Variants, v)
		if raw != nil {
			var saved publication.VerifiedListing
			if err = json.Unmarshal(raw, &saved); err != nil {
				rows.Close()
				return err
			}
			prior = append(prior, saved)
		}
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return err
	}
	for _, saved := range prior {
		found := false
		for _, fresh := range verified {
			if fresh.SKU != saved.SKU {
				continue
			}
			fresh.SiteFamilyName = saved.SiteFamilyName
			if fresh != saved {
				return nil
			}
			found = true
		}
		if !found {
			verified = append(verified, saved)
		}
	}
	result, err := publication.MatchReadback(a.prepared, a.scope, verified)
	if err != nil {
		return nil
	} // Contradictory evidence leaves durable associations unchanged.
	for _, v := range result.Matched {
		raw, err := json.Marshal(v.Listing)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE app.cbt_publication_identities
 SET verified_listing=CASE WHEN title_verified AND NOT $4 THEN verified_listing ELSE $3::jsonb END,
 title_verified=title_verified OR $4,last_verified_at=now()
 WHERE request_id=$1::uuid AND variant_key=$2`, a.id, v.Key, string(raw), v.TitleVerified); err != nil {
			return err
		}
	}
	if len(result.Pending) == 0 {
		if _, err = tx.Exec(ctx, `UPDATE app.cbt_publication_attempts SET state='success'
 WHERE request_id=$1::uuid AND owner_user_id=$2::uuid`, a.id, owner); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
