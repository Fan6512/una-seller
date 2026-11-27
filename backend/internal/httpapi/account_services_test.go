package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func newAccountTestManager(t *testing.T) *AccountServices {
	t.Helper()
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{19}, 32))
	s, err := NewAccountServices(&pgxpool.Pool{}, key, "", "https://api.vendor.example.org/v1")
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAccountSecretsAreAuthenticatedAndOwnerBound(t *testing.T) {
	s := newAccountTestManager(t)
	secret := []byte(`{"api_key":"never-emit-this-token"}`)
	enc, err := s.encrypt("owner-a", "ai", secret)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(enc, secret) {
		t.Fatal("credential leaked in stored ciphertext")
	}
	got, err := s.decrypt("owner-a", "ai", enc)
	if err != nil || !bytes.Equal(got, secret) {
		t.Fatal("ciphertext failed to decrypt")
	}
	for _, mismatch := range [][2]string{{"owner-b", "ai"}, {"owner-a", "translation"}} {
		if _, err := s.decrypt(mismatch[0], mismatch[1], enc); err == nil {
			t.Fatalf("cross-owner/kind decrypt accepted: %v", mismatch)
		}
	}
	enc[len(enc)-1] ^= 1
	if _, err := s.decrypt("owner-a", "ai", enc); err == nil {
		t.Fatal("modified ciphertext accepted")
	}
}

func TestAccountEndpointAndProviderValidation(t *testing.T) {
	s := newAccountTestManager(t)
	good := accountServiceConfig{Provider: "openai-compatible",
		BaseURL: "https://api.vendor.example.org/v1", Model: "gpt-6-mini", APIKey: "test-only"}
	if err := s.validateAccountConfig("ai", good); err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{
		"http://api.vendor.example.org/v1", "https://127.0.0.1/v1",
		"https://169.254.169.254/v1", "https://localhost/v1",
		"https://api.vendor.example.org@localhost/v1",
		"https://api.vendor.example.org/v1?redirect=1",
		"https://api.vendor.example.org/v1/",
		"https://api.vendor.example.org/v1//other",
	} {
		if safeAccountAIBaseURL(bad) {
			t.Errorf("unsafe endpoint allowed: %s", bad)
		}
		c := good
		c.BaseURL = bad
		if err := s.validateAccountConfig("ai", c); err == nil {
			t.Errorf("unlisted AI endpoint accepted: %s", bad)
		}
	}
	for _, bad := range []string{"sk-\ntoken", "", strings.Repeat("x", 4097)} {
		c := good
		c.APIKey = bad
		if err := s.validateAccountConfig("ai", c); err == nil {
			t.Error("unsafe credential allowed")
		}
	}
	azure := accountServiceConfig{Provider: "azure", Region: "eastus", APIKey: "a"}
	if err := s.validateAccountConfig("translation", azure); err != nil {
		t.Fatal(err)
	}
	baidu := accountServiceConfig{Provider: "baidu", AppID: "1234567", APIKey: "secret"}
	if err := s.validateAccountConfig("translation", baidu); err != nil {
		t.Fatal(err)
	}
	doubao := accountServiceConfig{Provider: "doubao", APIKey: "mock-doubao"}
	if err := s.validateAccountConfig("translation", doubao); err != nil {
		t.Fatal(err)
	}
	invalidDoubao := doubao
	invalidDoubao.AppID = "unexpected"
	if err := s.validateAccountConfig("translation", invalidDoubao); err == nil {
		t.Fatal("Doubao accepted legacy APP ID")
	}
	bad := azure
	bad.BaseURL = "https://example.org/v1"
	if err := s.validateAccountConfig("translation", bad); err == nil {
		t.Fatal("translation endpoint override accepted")
	}
}

func TestAccountServicesRejectWeakEncryptionKey(t *testing.T) {
	if _, err := NewAccountServices(&pgxpool.Pool{}, "not-base64", "", ""); err == nil {
		t.Fatal("weak key accepted")
	}
	if _, err := NewAccountServices(nil, base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32)), "", ""); err == nil {
		t.Fatal("nil DB accepted")
	}
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, 32))
	if _, err := NewAccountServices(&pgxpool.Pool{}, key, strings.Repeat("v", 65), ""); err == nil {
		t.Fatal("65-character encryption key ID accepted despite DB 64-character limit")
	}
	if _, err := NewAccountServices(&pgxpool.Pool{}, key, strings.Repeat("v", 64), ""); err != nil {
		t.Fatalf("64-character encryption key ID rejected: %v", err)
	}
}

func TestAccountCredentialHTTPOwnerIsolationAndEncryptedPersistence(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if os.Getenv("CI") != "" {
			t.Fatal("TEST_DATABASE_URL missing")
		}
		t.Skip("requires disposable PostgreSQL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	store, err := collections.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer store.DB.Close()
	if err := migrations.Apply(ctx, store.DB); err != nil {
		t.Fatal(err)
	}
	owner := "6a880000-0000-4000-8000-000000000075"
	other := "6a880000-0000-4000-8000-000000000076"
	// A first-login user has no app.users FK row until their first write.
	for _, id := range []string{owner, other} {
		store.DB.Exec(ctx, "DELETE FROM app.user_service_credentials WHERE owner_user_id=$1", id)
		store.DB.Exec(ctx, "DELETE FROM app.users WHERE id=$1", id)
	}
	var ownerExists bool
	if err := store.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app.users WHERE id=$1)", owner).Scan(&ownerExists); err != nil || ownerExists {
		t.Fatalf("expected new account without FK owner: exists=%v err=%v", ownerExists, err)
	}
	defer func() {
		for _, id := range []string{owner, other} {
			store.DB.Exec(context.Background(), "DELETE FROM app.user_service_credentials WHERE owner_user_id=$1", id)
			store.DB.Exec(context.Background(), "DELETE FROM app.users WHERE id=$1", id)
		}
	}()
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{99}, 32))
	services, err := NewAccountServices(store.DB, key, "", "")
	if err != nil {
		t.Fatal(err)
	}
	call := func(user, method, path, body, origin string) *httptest.ResponseRecorder {
		handler := HandlerWithAccountServices(store, stubVerifier{id: user}, nil, nil, nil, nil, services)
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if user != "" {
			req.Header.Set("Authorization", "Bearer synthetic-session")
		}
		if body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res
	}
	path := "/api/v1/account/service-configs/ai"
	secret := "integration-should-never-be-returned"
	body := `{"provider":"openai-compatible","base_url":"https://api.openai.com/v1","model":"gpt-4o-mini","api_key":"` + secret + `"}`
	if result := call("", http.MethodPut, path, body, "http://tauri.localhost"); result.Code != 401 {
		t.Fatalf("unauthenticated write returned %d", result.Code)
	}
	if result := call(owner, http.MethodPut, path, body, "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); result.Code != 403 {
		t.Fatalf("browser-extension write returned %d", result.Code)
	}
	if result := call(owner, http.MethodPut, path, body, "http://tauri.localhost"); result.Code != 200 {
		t.Fatalf("first-use PUT returned %d %s", result.Code, result.Body.String())
	}
	if err := store.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM app.users WHERE id=$1)", owner).Scan(&ownerExists); err != nil || !ownerExists {
		t.Fatalf("PUT did not create app.users row: exists=%v err=%v", ownerExists, err)
	}
	var sealed []byte
	if err := store.DB.QueryRow(ctx, "SELECT encrypted_data FROM app.user_service_credentials WHERE owner_user_id=$1 AND kind='ai'", owner).Scan(&sealed); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(sealed, []byte(secret)) {
		t.Fatal("secret leaked in raw PostgreSQL column")
	}
	// Changing the server key identifier cannot silently use old ciphertext.
	rotated, err := NewAccountServices(store.DB, key, "v2", "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rotated.AI(ctx, owner); !errors.Is(err, errAccountCredentialRebind) {
		t.Fatalf("rotated key should require rebind, got: %v", err)
	}
	rotatedHandler := HandlerWithAccountServices(store, stubVerifier{id: owner}, nil, nil, nil, nil, rotated)
	request := httptest.NewRequest("GET", "/api/v1/account/service-configs", nil)
	request.Header.Set("Authorization", "Bearer synthetic-session")
	request.Header.Set("Origin", "http://tauri.localhost")
	response := httptest.NewRecorder()
	rotatedHandler.ServeHTTP(response, request)
	if response.Code != 200 || !strings.Contains(response.Body.String(), "\"needs_rebind\":true") {
		t.Fatalf("rotation metadata did not require user action: %d %s", response.Code, response.Body.String())
	}
	// Baidu APP ID is public configuration metadata, never the secret Key.
	baiduBody := `{"provider":"baidu","app_id":"demo-app-009","api_key":"test-baidu-secret"}`
	if result := call(owner, http.MethodPut, "/api/v1/account/service-configs/translation", baiduBody, "http://tauri.localhost"); result.Code != 200 {
		t.Fatalf("Baidu save returned %d: %s", result.Code, result.Body.String())
	}
	if result := call(owner, http.MethodGet, "/api/v1/account/service-configs", "", "http://tauri.localhost"); !strings.Contains(result.Body.String(), "\"app_id\":\"demo-app-009\"") || strings.Contains(result.Body.String(), "test-baidu-secret") {
		t.Fatalf("Baidu metadata missing or secret exposed: %s", result.Body.String())
	}
	for _, tc := range []struct {
		user string
		want string
	}{
		{owner, "openai-compatible"},
		{other, "[]"},
	} {
		result := call(tc.user, http.MethodGet, "/api/v1/account/service-configs", "", "http://tauri.localhost")
		if result.Code != 200 || !strings.Contains(result.Body.String(), tc.want) {
			t.Fatalf("owner-scoped read: %s status=%d body=%s", tc.user, result.Code, result.Body.String())
		}
		if strings.Contains(result.Body.String(), secret) {
			t.Fatal("plain credential returned from GET")
		}
	}
	if result := call(other, http.MethodDelete, path, "", "http://tauri.localhost"); result.Code != 204 {
		t.Fatalf("foreign delete returned %d", result.Code)
	}
	if result := call(owner, http.MethodGet, "/api/v1/account/service-configs", "", "http://tauri.localhost"); !strings.Contains(result.Body.String(), "openai-compatible") {
		t.Fatal("foreign delete mutated owner credential")
	}
	if result := call(owner, http.MethodDelete, path, "", "http://tauri.localhost"); result.Code != 204 {
		t.Fatalf("owner delete returned %d", result.Code)
	}
	if result := call(owner, http.MethodGet, "/api/v1/account/service-configs", "", "http://tauri.localhost"); strings.Contains(result.Body.String(), "\"kind\":\"ai\"") || !strings.Contains(result.Body.String(), "\"provider\":\"baidu\"") {
		t.Fatalf("deleted AI credential still returned or Baidu credential lost: %s", result.Body.String())
	}
	if result := call(owner, http.MethodDelete, "/api/v1/account/service-configs/translation", "", "http://tauri.localhost"); result.Code != 204 {
		t.Fatalf("translation credential deletion returned %d", result.Code)
	}
	if result := call(owner, http.MethodGet, "/api/v1/account/service-configs", "", "http://tauri.localhost"); strings.TrimSpace(result.Body.String()) != "[]" {
		t.Fatalf("deleted credentials still returned: %s", result.Body.String())
	}
}
