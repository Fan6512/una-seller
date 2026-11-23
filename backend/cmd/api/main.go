package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/httpapi"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/migrations"
)

func main() {
	// Bind to loopback in both local and cloud mode; public HTTPS is served by Nginx.
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
	if os.Getenv("UNA_SELLER_CLOUD") == "1" && (dsn == "" || supabaseURL == "" || publishableKey == "") {
		log.Fatal("cloud API requires DATABASE_URL, SUPABASE_URL and SUPABASE_PUBLISHABLE_KEY")
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
	var meliService *meli.Service
	id, secret, redirect, key := os.Getenv("MELI_CLIENT_ID"), os.Getenv("MELI_CLIENT_SECRET"), os.Getenv("MELI_REDIRECT_URI"), os.Getenv("MELI_TOKEN_ENCRYPTION_KEY")
	if strings.TrimSpace(id+secret+redirect+key) != "" {
		if store == nil {
			log.Fatal("Mercado Libre OAuth needs configured database and Supabase Auth")
		}
		meliService, err = meli.NewService(store.DB, meli.Config{
			ClientID:           id,
			ClientSecret:       secret,
			RedirectURI:        redirect,
			TokenEncryptionKey: key,
			PKCE:               os.Getenv("MELI_PKCE_ENABLED") == "1",
		})
		if err != nil {
			log.Fatalf("invalid Mercado Libre OAuth config: %v", err)
		}
	}
	log.Printf("Una Seller API at http://%s (collections configured: %t; Mercado Libre OAuth configured: %t)", addr, store != nil, meliService != nil)
	translator := httpapi.NewAzureTranslationService(os.Getenv("AZURE_TRANSLATOR_KEY"), os.Getenv("AZURE_TRANSLATOR_REGION"))
	aiCopy, err := httpapi.NewAICopyService(os.Getenv("AI_BASE_URL"), os.Getenv("AI_API_KEY"), os.Getenv("AI_MODEL"))
	if err != nil {
		log.Print(err)
	}
	server := &http.Server{Addr: addr, Handler: httpapi.HandlerWithServices(store, verifier, meliService, translator, aiCopy), ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
