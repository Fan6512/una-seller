package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
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
	replayCommand := len(os.Args) > 1 && os.Args[1] == "missed-feeds-replay"
	billingProbeCommand := len(os.Args) > 1 && os.Args[1] == "billing-probe"
	if len(os.Args) > 1 && !replayCommand && !billingProbeCommand {
		log.Fatal("usage: api [migrate | cbt-categories update [-max-batches N] | cbt-categories review | cbt-categories incremental {plan|build|assemble|publish|rollback} | missed-feeds-replay OWNER_UUID STORE_UUID TOPIC | billing-probe OWNER_UUID STORE_UUID SITE LOGISTIC PACK_ID]")
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
	if billingProbeCommand {
		if len(os.Args) != 7 || store == nil || meliService == nil ||
			os.Getenv("UNA_SELLER_BILLING_PROBE_ENABLED") != "1" {
			log.Fatal("billing probe requires explicit operator enablement, configured OAuth and OWNER_UUID STORE_UUID SITE LOGISTIC PACK_ID")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		report, err := meliService.ProbeBillingByPack(ctx, os.Args[2], os.Args[3],
			meli.OrderScope{Site: os.Args[4], Logistic: os.Args[5]}, os.Args[6])
		if err != nil {
			log.Fatal("billing probe failed; review scoped order, authorization, provider and database")
		}
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			log.Fatal("billing probe report unavailable")
		}
		return
	}
	if replayCommand {
		if len(os.Args) != 5 || store == nil || meliService == nil ||
			os.Getenv("MELI_WEBHOOK_ENABLED") != "1" || os.Getenv("MELI_MISSED_FEEDS_REPLAY_ENABLED") != "1" {
			log.Fatal("missed feed replay requires explicit operator enablement, configured worker and OWNER_UUID STORE_UUID TOPIC")
		}
		topics, err := notifications.ParseEnabledTopics(os.Getenv("MELI_WEBHOOK_TOPICS"))
		if err != nil || !slices.Contains(topics, os.Args[4]) {
			log.Fatal("missed feed replay topic is not enabled")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		inbox := notifications.PostgresInbox{DB: store.DB}
		if err := inbox.CheckActivationSchema(ctx); err != nil {
			log.Fatal("notification activation schema preflight failed")
		}
		if err := meliService.EnableSharedOrderCoordination(ctx); err != nil {
			log.Fatal("shared order coordination unavailable")
		}
		result, err := inbox.ReplayMissedFeedPage(ctx, id, os.Args[4], os.Args[3], func(ctx context.Context, topic string, offset int) ([]byte, error) {
			return meliService.FetchMissedFeedPage(ctx, os.Args[2], os.Args[3], topic, offset)
		})
		if err != nil {
			// No provider body, token, resource ID or buyer data in operator logs.
			log.Fatal("missed feed replay failed; reread durable checkpoint before retry, review authorization, contract, capacity and database")
		}
		log.Printf("missed feed replay offset=%d next_offset=%d records=%d cycle_complete=%t", result.Offset, result.NextOffset, result.Records, result.CycleComplete)
		return
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
		topics, err := notifications.ParseEnabledTopics(os.Getenv("MELI_WEBHOOK_TOPICS"))
		if err != nil {
			log.Fatal(err)
		}
		all := meliService.NotificationHandlers()
		enabled := make(map[string]notifications.Prepare)
		for _, topic := range topics {
			enabled[topic] = all[topic]
		}
		if os.Getenv("MELI_COMPENSATION_ENABLED") == "1" &&
			(enabled["marketplace_orders"] == nil || enabled["marketplace_shipments"] == nil) {
			log.Fatal("compensation requires order and shipment topics")
		}
		checkCtx, checkCancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = inbox.CheckActivationSchema(checkCtx)
		checkCancel()
		if err != nil {
			log.Fatal("notification activation schema preflight failed")
		}
		worker := notifications.NewWorker(inbox, id, enabled)
		if err := worker.ValidateTopicsActivation(topics); err != nil {
			log.Fatal(err)
		}
		receiver, err := notifications.NewReceiverForTopics(id, os.Getenv("MELI_WEBHOOK_INGRESS_TOKEN"), inbox, topics)
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
	// Independent public reference data; never on the notification ACK path.
	if db != nil && os.Getenv("UNA_SELLER_DAILY_FX_ENABLED") == "1" {
		go meli.RunDailyFX(context.Background(), db, log.Default())
	}
	// Billing reconciliation is independent and disabled unless explicitly enabled.
	if os.Getenv("UNA_SELLER_BILLING_COLLECTOR_ENABLED") == "1" {
		if meliService != nil {
			go meliService.RunBillingCollector(context.Background(), log.Default())
		} else {
			log.Print("billing collector disabled: Mercado Libre OAuth unavailable")
		}
	}
	server := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
