package publication

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTitleModes(t *testing.T) {
	titles := LocalizedText{English: "Example Tool", Spanish: "Herramienta de ejemplo", Portuguese: "Ferramenta de exemplo"}
	desc := LocalizedText{English: "Example description", Spanish: "Descripción de ejemplo"}
	in := sample()
	for _, tc := range []struct {
		mode       TitleMode
		site, want string
	}{
		{PlatformAuto, "MLM", ""}, {Localized, "MLM", titles.Spanish}, {Localized, "MLB", titles.Portuguese},
	} {
		in.Target.Site = tc.site
		c, err := SelectCopy(tc.site, tc.mode, titles, desc)
		if err != nil {
			t.Fatal(err)
		}
		in.Copy = c
		p, err := Prepare(in)
		if err != nil {
			t.Fatal(err)
		}
		var items []up
		if err := json.Unmarshal(p.Body, &items); err != nil {
			t.Fatal(err)
		}
		for _, item := range items {
			if item.FamilyName != titles.English || item.Sites[0].FamilyName != tc.want {
				t.Fatalf("incorrect title mapping: %s", p.Body)
			}
		}
		if tc.mode == PlatformAuto && strings.Contains(string(p.Body), titles.Spanish) {
			t.Fatal("automatic mode must not send localized copy")
		}
		if len(p.PayloadSHA256()) != 64 {
			t.Fatal("missing payload hash")
		}
	}
	for _, title := range []string{"", strings.Repeat("á", 61)} {
		titles.Spanish = title
		if _, err := SelectCopy("MLM", Localized, titles, desc); err == nil {
			t.Fatal("invalid localized title must not fall back or truncate")
		}
	}
	if _, err := SelectCopy("MLM", "", titles, desc); err == nil {
		t.Fatal("mode must be explicit")
	}
}

func TestReadbackIdentityAndTitle(t *testing.T) {
	in := sample()
	in.Copy.Mode, in.Copy.SiteFamilyName = Localized, "Herramienta"
	p, err := Prepare(in)
	if err != nil {
		t.Fatal(err)
	}
	scope := ReadbackScope{ParentSellerID: "100", SiteSellerID: "200", Site: "MLM", Logistic: "remote"}
	a := VerifiedListing{Scope: scope, SKU: "TOOL-A", SitelessUPID: "U11", SitelessFamilyID: "9000000000000001", SiteUPID: "MLMU11", SiteFamilyID: "9000000000000002", ItemID: "MLM101", SiteFamilyName: "Herramienta"}
	b := VerifiedListing{Scope: scope, SKU: "TOOL-B", SitelessUPID: "U12", SitelessFamilyID: a.SitelessFamilyID, SiteUPID: "MLMU12", SiteFamilyID: a.SiteFamilyID, ItemID: "MLM102", SiteFamilyName: "Other title"}
	got, err := MatchReadback(p, scope, []VerifiedListing{b, a})
	if err != nil || len(got.Matched) != 2 || got.Matched[0].Key != "variant-a" || got.Matched[0].Listing.ItemID != a.ItemID || !got.Matched[0].TitleVerified || got.Matched[1].TitleVerified || len(got.Pending) != 0 {
		t.Fatalf("incorrect readback: %+v %v", got, err)
	}
	partial, err := MatchReadback(p, scope, []VerifiedListing{b})
	if err != nil || len(partial.Matched) != 1 || len(partial.Pending) != 1 || partial.Pending[0] != "variant-a" {
		t.Fatalf("missing evidence must remain pending: %+v %v", partial, err)
	}
	for _, change := range []func(*VerifiedListing){
		func(v *VerifiedListing) { v.SitelessFamilyID = "123" },
		func(v *VerifiedListing) { v.Scope.SiteSellerID = "999" },
		func(v *VerifiedListing) { v.ItemID = a.ItemID },
		func(v *VerifiedListing) { v.SKU = a.SKU },
	} {
		bad := b
		change(&bad)
		if _, err := MatchReadback(p, scope, []VerifiedListing{a, bad}); err == nil {
			t.Fatal("conflicting evidence must not confirm a family")
		}
	}
}
