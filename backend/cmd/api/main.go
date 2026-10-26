package main

import (
	"log"
	"net/http"

	"github.com/Fan6512/una-seller/backend/internal/httpapi"
)

func main() {
	// M0-01 only: strictly local API, not a public or authenticated service.
	const addr = "127.0.0.1:8787"
	log.Printf("Una Seller local development API at http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, httpapi.Handler()))
}
