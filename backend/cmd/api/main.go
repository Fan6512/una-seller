package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Fan6512/una-seller/backend/internal/cbtcategory"
	"github.com/Fan6512/una-seller/backend/internal/collections"
	"github.com/Fan6512/una-seller/backend/internal/httpapi"
	"github.com/Fan6512/una-seller/backend/internal/meli"
	"github.com/Fan6512/una-seller/backend/internal/notifications"
	"github.com/Fan6512/una-seller/backend/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	translator := configuredTranslator()
	dictionaryDir := os.Getenv("CBT_CATEGORY_DICTIONARY_DIR")
	if len(os.Args) > 1 && os.Args[1] == "cbt-categories" {
		if err := updateCBTCategories(os.Args[2:], dictionaryDir, translator); err != nil {
			log.Fatal(err)
		}
		return
	}
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
		log.Fatal("usage: api [migrate | cbt-categories update [-max-batches N]]")
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
	dictionary := cbtcategory.NewReader(dictionaryDir)
	status := dictionary.Status()
	log.Printf("CBT category dictionary: state=%s translated=%d version=%s (paths use official English if untranslated)", status.State, status.Translated, status.Version)
	aiCopy, err := httpapi.NewAICopyService(os.Getenv("AI_BASE_URL"), os.Getenv("AI_API_KEY"), os.Getenv("AI_MODEL"))
	if err != nil {
		log.Printf("AI listing copy disabled: %v", err)
	}
	var db *pgxpool.Pool
	if store != nil {
		db = store.DB
	}
	accounts, err := httpapi.NewAccountServices(db, os.Getenv("UNA_SELLER_SERVICE_CREDENTIALS_KEY"), os.Getenv("UNA_SELLER_SERVICE_CREDENTIALS_KEY_ID"), os.Getenv("AI_ALLOWED_BASE_URLS"))
	if err != nil {
		log.Fatalf("account credentials config: %v", err)
	}
	if accounts == nil && os.Getenv("UNA_SELLER_CLOUD") == "1" {
		log.Print("WARNING: account AI/translation key management is disabled: UNA_SELLER_SERVICE_CREDENTIALS_KEY is not configured")
	}
	handler := httpapi.HandlerWithAccountServices(store, verifier, meliService, translator, dictionary, aiCopy, accounts)
	if os.Getenv("MELI_COMPENSATION_ENABLED") == "1" && os.Getenv("MELI_WEBHOOK_ENABLED") != "1" {
		log.Fatal("compensation requires the enabled notification worker")
	}
	if os.Getenv("MELI_WEBHOOK_ENABLED") == "1" {
		if db == nil || meliService == nil {
			log.Fatal("notifications need configured database and Mercado Libre OAuth")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := meliService.EnableSharedOrderCoordination(ctx)
		cancel()
		if err != nil {
			log.Fatal("shared order coordination unavailable: ", err)
		}
		inbox := notifications.PostgresInbox{DB: db}
		// All accepted topics must register real read-only handlers before activation.
		worker := notifications.NewWorker(inbox, id, meliService.NotificationHandlers())
		if err := worker.ValidateActivation(); err != nil {
			log.Fatal(err)
		}
		receiver, err := notifications.NewReceiver(id, os.Getenv("MELI_WEBHOOK_INGRESS_TOKEN"), inbox)
		if err != nil {
			log.Fatal(err)
		}
		mux := http.NewServeMux()
		mux.Handle(notifications.Path, receiver)
		mux.Handle("/", handler)
		handler = mux
		go worker.Run(context.Background(), log.Default())
		if os.Getenv("MELI_COMPENSATION_ENABLED") == "1" {
			checkCtx, checkCancel := context.WithTimeout(context.Background(), 5*time.Second)
			var ready bool
			err := db.QueryRow(checkCtx, `SELECT to_regclass('app.meli_order_discovery') IS NOT NULL
				AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema='app'
					AND table_name='meli_notification_inbox' AND column_name='source')`).Scan(&ready)
			checkCancel()
			if err != nil || !ready {
				log.Fatal("order compensation migrations required")
			}
			go inbox.RunCompensation(context.Background(), id, log.Default(), meliService.SearchDiscoveryOrders)
		}
	}
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
