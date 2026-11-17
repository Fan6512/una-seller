package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/httpapi"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func main() {
	// Authenticated local development API. Public deployment requires separate
	// HTTPS, proxy and client-origin review; never expose this listener directly.
	const addr = "127.0.0.1:8787"
	local, err := loadLocalAPIConfig()
	if err != nil {
		log.Fatalf("local API configuration: %v", err)
	}
	dsn := configValue("DATABASE_URL", local.DatabaseURL)
	supabaseURL := configValue("SUPABASE_URL", local.SupabaseURL)
	publishableKey := configValue("SUPABASE_PUBLISHABLE_KEY", local.SupabasePublishableKey)
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if dsn == "" {
			log.Fatal("DATABASE_URL required for migrations")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		store, err := collections.Open(ctx, dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer store.DB.Close()
		if err = migrations.Apply(ctx, store.DB); err != nil {
			log.Fatal(err)
		}
		log.Print("source collections migration applied")
		return
	}
	if len(os.Args) > 1 {
		log.Fatal("usage: api [migrate]")
	}
	var store *collections.Store
	var verifier httpapi.IdentityVerifier
	if dsn != "" {
		var err error
		verifier, err = httpapi.NewSupabaseVerifier(supabaseURL, publishableKey)
		if err != nil {
			log.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		store, err = collections.Open(ctx, dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer store.DB.Close()
	}
	log.Printf("Una Seller local development API at http://%s (collections configured: %t)", addr, store != nil)
	log.Fatal(http.ListenAndServe(addr, httpapi.HandlerWithCollections(store, verifier)))
}
