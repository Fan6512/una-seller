package collections

import (
	"encoding/json"
	"testing"
)

func TestMissingAndNullFields(t *testing.T) {
	var a, b Input
	if err := json.Unmarshal([]byte(`{"source_platform":"amazon","source_site":"US","source_url":"https://amazon.com/dp/B123456789","collection_status":"partial"}`), &a); err != nil {
		t.Fatal(err)
	}
	if a.Price.Present {
		t.Fatal("missing price treated as present")
	}
	if err := json.Unmarshal([]byte(`{"price":null,"bullet_points":[]}`), &b); err != nil {
		t.Fatal(err)
	}
	if !b.Price.Present || !b.Price.Null || !b.BulletPoints.Present {
		t.Fatal("explicit null/empty lost")
	}
}

func TestNormalizeAndValidate(t *testing.T) {
	in := Input{Platform: "amazon", Site: "US", SourceURL: "https://www.amazon.com/gp/product/B123456789?ref_=abc", Status: "partial"}
	got, err := in.Validate()
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://www.amazon.com/dp/B123456789" {
		t.Fatalf("got %s", got)
	}
	in.SourceURL = "https://amazon.com.evil.test/dp/B123456789"
	if _, err = in.Validate(); err == nil {
		t.Fatal("accepted unsafe host")
	}
	in.SourceURL = "https://www.amazon.com/dp/B123456789"
	in.Price = Field[string]{Present: true, Value: "12.34567"}
	if _, err = in.Validate(); err == nil {
		t.Fatal("accepted imprecise money")
	}
}
