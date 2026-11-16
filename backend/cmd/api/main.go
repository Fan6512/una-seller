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
	// M0-03 is still local-only. Never expose these routes through public Nginx:
	// user login and browser extension pairing are not implemented in PR-A.
	const addr = "127.0.0.1:8787"
	dsn := os.Getenv("DATABASE_URL")
	owner := os.Getenv("UNA_SELLER_LOCAL_USER_ID")
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
	if dsn != "" {
		if !collections.ValidUserID(owner) {
			log.Fatal("UNA_SELLER_LOCAL_USER_ID must be a persistent UUID")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		var err error
		store, err = collections.Open(ctx, dsn)
		if err != nil {
			log.Fatal(err)
		}
		defer store.DB.Close()
	}
	log.Printf("Una Seller local development API at http://%s (collections configured: %t)", addr, store != nil)
	log.Fatal(http.ListenAndServe(addr, httpapi.HandlerWithCollections(store, owner)))
}
