package publication

import (
	"encoding/json"
	"strings"
	"testing"
)

func attr(id, valueID, name string) Attribute {
	return Attribute{ID: id, Values: []Value{{ID: valueID, Name: name}}}
}

func sample() Input {
	q1, q2 := 10, 0
	common := func(color string) []Attribute {
		return []Attribute{attr("BRAND", "1", "Example"), attr("COLOR", color, ""),
			attr("ITEM_CONDITION", "2230284", "New")}
	}
	return Input{
		FamilyName: "Example Tool", CategoryID: "CBT123", Description: "Example plain text.",
		Target: Target{Site: "MLM", Logistic: "remote", Mode: "standard_up", Pricing: "net_proceeds", ListingType: "gold_pro"},
		Rules:  []Rule{{ID: "BRAND", Hierarchy: "PARENT_PK", Enumerated: true}, {ID: "COLOR", Hierarchy: "CHILD_PK", Enumerated: true}, {ID: "ITEM_CONDITION", Required: true, Enumerated: true}},
		Variants: []Variant{
			{Key: "variant-a", SKU: "TOOL-A", Quantity: &q1, NetProceeds: "12.35", PictureIDs: []string{"123-CBT123_112026"}, Attributes: common("10")},
			{Key: "variant-b", SKU: "TOOL-B", Quantity: &q2, NetProceeds: "13.125", PictureIDs: []string{"456-CBT456_112026"}, Attributes: common("11")},
		},
	}
}

func TestPrepareFamilyAndSingle(t *testing.T) {
	in := sample()
	got, err := Prepare(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Path != FamilyPath || len(got.VariantKeys) != 2 || got.VariantKeys[1] != "variant-b" {
		t.Fatalf("bad route or local mapping: %+v", got)
	}
	var items []up
	if err := json.Unmarshal(got.Body, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].FamilyName != items[1].FamilyName ||
		items[0].Quantity != 10 || items[1].Quantity != 0 ||
		items[0].NetProceeds != "12.35" || items[1].NetProceeds != "13.125" ||
		items[0].Pictures[0].ID == items[1].Pictures[0].ID ||
		items[0].Currency != "USD" || len(items[0].Sites) != 1 || items[0].Sites[0].ID != "MLM" {
		t.Fatalf("incorrect family body %s", got.Body)
	}
	if !strings.Contains(string(got.Body), `"global_net_proceeds":13.125`) ||
		!strings.Contains(string(got.Body), `"value_name":"TOOL-B"`) {
		t.Fatal("amount or SKU not encoded")
	}
	for _, forbidden := range []string{`"title":`, `"variations":`, `"price":`, "variant-a", "purchase_cost"} {
		if strings.Contains(string(got.Body), forbidden) {
			t.Fatalf("unexpected wire field %s", forbidden)
		}
	}
	in.Variants = in.Variants[:1]
	in.Variants[0].Key = SingleKey
	single, err := Prepare(in)
	if err != nil || single.Path != SinglePath || len(single.Body) == 0 || single.Body[0] != '{' {
		t.Fatalf("bad single: %s %v", single.Body, err)
	}
}

func TestPrepareRejectsInvalidFamily(t *testing.T) {
	cases := map[string]func(*Input){
		"parent mismatch": func(in *Input) { in.Variants[1].Attributes[0] = attr("BRAND", "2", "") },
		"duplicate child": func(in *Input) { in.Variants[1].Attributes[1] = attr("COLOR", "10", "Different label") },
		"missing child": func(in *Input) {
			in.Variants[1].Attributes = append(in.Variants[1].Attributes[:1], in.Variants[1].Attributes[2:]...)
		},
		"duplicate SKU": func(in *Input) { in.Variants[1].SKU = in.Variants[0].SKU },
		"duplicate key": func(in *Input) { in.Variants[1].Key = in.Variants[0].Key },
		"missing stock": func(in *Input) { in.Variants[1].Quantity = nil },
		"picture URL":   func(in *Input) { in.Variants[1].PictureIDs = []string{"https://example.com/a.jpg"} },
		"zero proceeds": func(in *Input) { in.Variants[1].NetProceeds = "0.00" },
		"wrong mode":    func(in *Input) { in.Target.Mode = "legacy_candidate" },
		"wrong site":    func(in *Input) { in.Target.Site = "MLZ" },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			in := sample()
			change(&in)
			got, err := Prepare(in)
			if err == nil || len(got.Body) != 0 {
				t.Fatalf("invalid input prepared: %s", got.Body)
			}
		})
	}
}
