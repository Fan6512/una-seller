package meli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var ErrReauthorizationRequired = errors.New("Mercado Libre reauthorization required")
var ErrRefreshUnknown = errors.New("Mercado Libre token refresh outcome unknown")
var ErrRefreshBusy = errors.New("Mercado Libre token refresh in progress")
var ErrSiteNotAuthorized = errors.New("marketplace site not verified for this seller")

const refreshAhead = 60 * time.Second

// AuthorizedStore is server-only. Do not JSON-encode it or send its token to clients.
type AuthorizedStore struct {
	SellerID	string
	AccessToken	string
	Markets	[]Marketplace
	Verified	bool
}

type tokenSnapshot struct {
	sellerID	string
	status	string
	accessCipher	[]byte
	refreshCipher	[]byte
	expires	*time.Time
	lock	*string
	leaseUntil	*time.Time
	marketsJSON	[]byte
	marketsStatus	string
}

func (s *Service) snapshot(ctx context.Context, owner, storeID string) (tokenSnapshot, error) {
	if !validUUID(storeID) {
		return tokenSnapshot{}, ErrStoreNotFound
	}
	var row tokenSnapshot
	err := s.db.QueryRow(ctx, `SELECT seller_id,status,access_token_ciphertext,refresh_token_ciphertext,
 access_token_expires_at,refresh_lock::text,refresh_lease_until,marketplaces,marketplaces_status
 FROM app.meli_store_connections WHERE id=$1::uuid AND owner_user_id=$2::uuid`,
		storeID, owner).Scan(&row.sellerID, &row.status, &row.accessCipher, &row.refreshCipher,
		&row.expires, &row.lock, &row.leaseUntil, &row.marketsJSON, &row.marketsStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return tokenSnapshot{}, ErrStoreNotFound
	}
	return row, err
}

func tokenStatusError(status string) error {
	switch status {
	case "connected":
		return nil
	case "reauth_required", "disconnected":
		return ErrReauthorizationRequired
	case "refresh_unknown":
		return ErrRefreshUnknown
	default:
		return ErrReauthorizationRequired
	}
}

func (s *Service) accessFromSnapshot(owner string, row tokenSnapshot) (AuthorizedStore, error) {
	if err := tokenStatusError(row.status); err != nil {
		return AuthorizedStore{}, err
	}
	if len(row.accessCipher) == 0 || row.expires == nil {
		return AuthorizedStore{}, ErrReauthorizationRequired
	}
	token, err := s.cipher.Open(row.accessCipher, tokenAAD(owner, row.sellerID, "access"))
	if err != nil {
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	var markets []Marketplace
	if err = json.Unmarshal(row.marketsJSON, &markets); err != nil {
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	return AuthorizedStore{
		SellerID: row.sellerID, AccessToken: token,
		Markets: markets, Verified: row.marketsStatus == "verified",
	}, nil
}

// RequireStoreConnection is the shared backend auth gate for future order and
// publication handlers. Supabase identity MUST be verified before passing owner.
func (s *Service) RequireStoreConnection(ctx context.Context, owner, storeID string) (AuthorizedStore, error) {
	row, err := s.snapshot(ctx, owner, storeID)
	if err != nil {
		return AuthorizedStore{}, err
	}
	if err = tokenStatusError(row.status); err != nil {
		return AuthorizedStore{}, err
	}
	if row.expires != nil && row.expires.After(time.Now().Add(refreshAhead)) {
		return s.accessFromSnapshot(owner, row)
	}
	// A single Go process will perform one refresh for this store at a time;
	// the DB lease below protects multiple Go processes sharing the database.
	value, err, _ := s.refreshGroup.Do(storeID, func() (any, error) {
		return s.rotate(ctx, owner, storeID, false)
	})
	if errors.Is(err, ErrRefreshBusy) {
		return s.waitForRefresh(ctx, owner, storeID)
	}
	if err != nil {
		return AuthorizedStore{}, err
	}
	return value.(AuthorizedStore), nil
}

// RequireSiteConnection also checks that this exact site-child relationship
// was confirmed by the platform; a hard-coded list of CBT countries is unsafe.
func (s *Service) RequireSiteConnection(ctx context.Context, owner, storeID, siteID string) (AuthorizedStore, string, error) {
	auth, err := s.RequireStoreConnection(ctx, owner, storeID)
	if err != nil {
		return AuthorizedStore{}, "", err
	}
	if !auth.Verified {
		return AuthorizedStore{}, "", ErrSiteNotAuthorized
	}
	for _, child := range auth.Markets {
		if child.SiteID == siteID {
			return auth, child.UserID.String(), nil
		}
	}
	return AuthorizedStore{}, "", ErrSiteNotAuthorized
}

// ForceRefresh exists for controlled integration testing, not a public endpoint.
func (s *Service) ForceRefresh(ctx context.Context, owner, storeID string) error {
	_, err, _ := s.refreshGroup.Do(storeID, func() (any, error) {
		return s.rotate(ctx, owner, storeID, true)
	})
	return err
}

func (s *Service) waitForRefresh(ctx context.Context, owner, storeID string) (AuthorizedStore, error) {
	timer := time.NewTimer(16 * time.Second)
	defer timer.Stop()
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return AuthorizedStore{}, ctx.Err()
		case <-timer.C:
			return AuthorizedStore{}, ErrRefreshBusy
		case <-ticker.C:
			row, err := s.snapshot(ctx, owner, storeID)
			if err != nil {
				return AuthorizedStore{}, err
			}
			if err = tokenStatusError(row.status); err != nil {
				return AuthorizedStore{}, err
			}
			if row.lock == nil {
				if row.expires != nil && row.expires.After(time.Now().Add(refreshAhead)) {
					return s.accessFromSnapshot(owner, row)
				}
				return AuthorizedStore{}, ErrRefreshUnknown
			}
			if row.leaseUntil == nil || !row.leaseUntil.After(time.Now()) {
				return AuthorizedStore{}, ErrRefreshUnknown
			}
		}
	}
}

// claimRefresh holds SELECT FOR UPDATE only during the short lease transaction.
// Network calls and token exchange must happen AFTER commit.
func (s *Service) claimRefresh(ctx context.Context, owner, storeID string, force bool) (tokenSnapshot, string, bool, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return tokenSnapshot{}, "", false, err
	}
	defer tx.Rollback(ctx)
	var row tokenSnapshot
	err = tx.QueryRow(ctx, `SELECT seller_id,status,access_token_ciphertext,refresh_token_ciphertext,
 access_token_expires_at,refresh_lock::text,refresh_lease_until,marketplaces,marketplaces_status
 FROM app.meli_store_connections WHERE id=$1::uuid AND owner_user_id=$2::uuid FOR UPDATE`,
		storeID, owner).Scan(&row.sellerID, &row.status, &row.accessCipher, &row.refreshCipher,
		&row.expires, &row.lock, &row.leaseUntil, &row.marketsJSON, &row.marketsStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return tokenSnapshot{}, "", false, ErrStoreNotFound
	}
	if err != nil {
		return tokenSnapshot{}, "", false, err
	}
	if err = tokenStatusError(row.status); err != nil {
		return tokenSnapshot{}, "", false, err
	}
	if row.lock != nil {
		// Once a lease expires, the remote refresh may ALREADY have consumed the
		// one-use token. Never automatically claim it again.
		if row.leaseUntil == nil || !row.leaseUntil.After(time.Now()) {
			_, err = tx.Exec(ctx, `UPDATE app.meli_store_connections SET status='refresh_unknown',
 refresh_lock=NULL,refresh_lease_until=NULL,updated_at=now() WHERE id=$1::uuid`, storeID)
			if err != nil {
				return tokenSnapshot{}, "", false, err
			}
			if err = tx.Commit(ctx); err != nil {
				return tokenSnapshot{}, "", false, err
			}
			return tokenSnapshot{}, "", false, ErrRefreshUnknown
		}
		return tokenSnapshot{}, "", false, ErrRefreshBusy
	}
	if !force && row.expires != nil && row.expires.After(time.Now().Add(refreshAhead)) {
		return row, "", false, nil
	}
	if len(row.refreshCipher) == 0 {
		return tokenSnapshot{}, "", false, ErrReauthorizationRequired
	}
	// Random UUID from PostgreSQL is a short-lived fencing token.
	var leaseID string
	if err = tx.QueryRow(ctx, "SELECT gen_random_uuid()::text").Scan(&leaseID); err != nil {
		return tokenSnapshot{}, "", false, err
	}
	if _, err = tx.Exec(ctx, `UPDATE app.meli_store_connections SET refresh_lock=$1::uuid,
 refresh_lease_until=now()+interval '30 seconds' WHERE id=$2::uuid`, leaseID, storeID); err != nil {
		return tokenSnapshot{}, "", false, err
	}
	if err = tx.Commit(ctx); err != nil {
		return tokenSnapshot{}, "", false, err
	}
	return row, leaseID, true, nil
}

func (s *Service) refreshToken(ctx context.Context, oldToken string) (Token, error) {
	form := url.Values{
		"grant_type": {"refresh_token"}, "client_id": {s.cfg.ClientID},
		"client_secret": {s.cfg.ClientSecret}, "refresh_token": {oldToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL+"/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, ErrRefreshUnknown
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return Token{}, ErrRefreshUnknown
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var failure struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(resp.Body, 1<<15)).Decode(&failure)
		if failure.Error == "invalid_grant" {
			return Token{}, ErrReauthorizationRequired
		}
		return Token{}, ErrRefreshUnknown
	}
	var token Token
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<16)).Decode(&token); err != nil {
		return Token{}, ErrRefreshUnknown
	}
	if token.AccessToken == "" || token.RefreshToken == "" || token.ExpiresIn <= 0 || token.ExpiresIn > 30*24*60*60 {
		return Token{}, ErrRefreshUnknown
	}
	return token, nil
}

func (s *Service) rotate(ctx context.Context, owner, storeID string, force bool) (AuthorizedStore, error) {
	row, leaseID, claimed, err := s.claimRefresh(ctx, owner, storeID, force)
	if err != nil {
		return AuthorizedStore{}, err
	}
	if !claimed {
		return s.accessFromSnapshot(owner, row)
	}
	oldToken, err := s.cipher.Open(row.refreshCipher, tokenAAD(owner, row.sellerID, "refresh"))
	if err != nil {
		s.setRefreshFailure(owner, storeID, leaseID, "refresh_unknown")
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	token, err := s.refreshToken(ctx, oldToken)
	if err != nil {
		if errors.Is(err, ErrReauthorizationRequired) {
			s.setRefreshFailure(owner, storeID, leaseID, "reauth_required")
		} else {
			s.setRefreshFailure(owner, storeID, leaseID, "refresh_unknown")
		}
		return AuthorizedStore{}, err
	}
	access, err := s.cipher.Seal(token.AccessToken, tokenAAD(owner, row.sellerID, "access"))
	if err != nil {
		s.setRefreshFailure(owner, storeID, leaseID, "refresh_unknown")
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	refresh, err := s.cipher.Seal(token.RefreshToken, tokenAAD(owner, row.sellerID, "refresh"))
	if err != nil {
		s.setRefreshFailure(owner, storeID, leaseID, "refresh_unknown")
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	// Persist even if the original HTTP caller disconnected. A transient database
	// failure after a successful remote rotation must NOT retry the old token.
	saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd, err := s.db.Exec(saveCtx, `UPDATE app.meli_store_connections SET access_token_ciphertext=$1,
 refresh_token_ciphertext=$2,access_token_expires_at=now()+($3::bigint*interval '1 second'),
 status='connected',refresh_lock=NULL,refresh_lease_until=NULL,updated_at=now()
 WHERE id=$4::uuid AND owner_user_id=$5::uuid AND refresh_lock=$6::uuid AND status='connected'`,
		access, refresh, token.ExpiresIn, storeID, owner, leaseID)
	if err != nil || cmd.RowsAffected() != 1 {
		s.setRefreshFailure(owner, storeID, leaseID, "refresh_unknown")
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	var markets []Marketplace
	if err = json.Unmarshal(row.marketsJSON, &markets); err != nil {
		return AuthorizedStore{}, ErrRefreshUnknown
	}
	return AuthorizedStore{
		SellerID: row.sellerID, AccessToken: token.AccessToken,
		Markets: markets, Verified: row.marketsStatus == "verified",
	}, nil
}

func (s *Service) setRefreshFailure(owner, storeID, leaseID, status string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Fence by lease ID, so a stale worker cannot overwrite a newer authorization.
	_, _ = s.db.Exec(ctx, `UPDATE app.meli_store_connections SET status=$1,
 refresh_lock=NULL,refresh_lease_until=NULL,updated_at=now()
 WHERE id=$2::uuid AND owner_user_id=$3::uuid AND refresh_lock=$4::uuid AND status='connected'`,
		status, storeID, owner, leaseID)
}
