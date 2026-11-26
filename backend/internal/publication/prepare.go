// Package publication builds Standard UP publication requests without I/O.
// Callers must load the saved product, authorized target, current category
// rules and owned uploaded picture IDs on the server. A prepared body is not
// permission to publish; the submission service must persist intent first.
package publication

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	SinglePath = "/global/items"
	FamilyPath = "/global/user-products/families"
	SingleKey  = "single"
)

type Target struct {
	Site, Logistic, Mode, Pricing, ListingType string
}

// Rule is the normalized subset of the current official category schema.
// Conditional requirements and size charts must be resolved by the caller.
type Rule struct {
	ID, Hierarchy string
	Required      bool
	// Enumerated values must have a resolved official value ID.
	Enumerated bool
}

type Attribute struct {
	ID        string  `json:"id"`
	Values    []Value `json:"values,omitempty"`
	ValueName string  `json:"value_name,omitempty"`
}

type Value struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Variant struct {
	Key, SKU    string
	Quantity    *int
	NetProceeds string
	PictureIDs  []string
	Attributes  []Attribute
}

type Input struct {
	FamilyName, CategoryID, Description string
	Target                              Target
	Rules                               []Rule
	Variants                            []Variant
}

// Prepared keeps local variant keys separate from the official wire body.
// The keys describe request order, not a guarantee of response ordering.
type Prepared struct {
	Path        string
	Body        []byte
	VariantKeys []string
}

type site struct {
	ID          string `json:"site_id"`
	Logistic    string `json:"logistic_type"`
	ListingType string `json:"listing_type_id"`
}

type picture struct {
	ID string `json:"id"`
}

type description struct {
	PlainText string `json:"plain_text"`
}

type up struct {
	FamilyName  string      `json:"family_name"`
	CategoryID  string      `json:"category_id"`
	Currency    string      `json:"currency_id"`
	NetProceeds json.Number `json:"global_net_proceeds"`
	Quantity    int         `json:"available_quantity"`
	Description description `json:"description"`
	Pictures    []picture   `json:"pictures"`
	Attributes  []Attribute `json:"attributes"`
	Sites       []site      `json:"sites_to_sell"`
}

var (
	categoryPattern = regexp.MustCompile(`^CBT[0-9]{1,15}$`)
	sitePattern     = regexp.MustCompile(`^(MLA|MLB|MLC|MLM|MCO)$`)
	attrPattern     = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,79}$`)
	picturePattern  = regexp.MustCompile(`^[A-Za-z0-9_-]{1,160}$`)
	moneyPattern    = regexp.MustCompile(`^(0|[1-9][0-9]{0,15})(\.[0-9]{1,6})?$`)
)

func textOK(s string, max int) bool {
	return utf8.ValidString(s) && strings.TrimSpace(s) != "" &&
		utf8.RuneCountInString(s) <= max &&
		strings.IndexFunc(s, func(r rune) bool { return unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' }) < 0
}

// Prepare uses one normalized input for both create endpoints. It does not
// read or mutate DB state, upload images, make HTTP calls, or invent defaults.
func Prepare(in Input) (Prepared, error) {
	t := in.Target
	if t.Mode != "standard_up" || t.Pricing != "net_proceeds" || t.Logistic != "remote" ||
		!sitePattern.MatchString(t.Site) || (t.ListingType != "gold_pro" && t.ListingType != "gold_special") {
		return Prepared{}, fmt.Errorf("unsupported publication target")
	}
	if !categoryPattern.MatchString(in.CategoryID) || !textOK(in.FamilyName, 500) ||
		!textOK(in.Description, 20000) || len(in.Variants) == 0 || len(in.Variants) > 100 {
		return Prepared{}, fmt.Errorf("invalid publication input")
	}
	rules := make(map[string]Rule, len(in.Rules))
	var parents, children []string
	for _, rule := range in.Rules {
		if !attrPattern.MatchString(rule.ID) {
			return Prepared{}, fmt.Errorf("invalid category rule")
		}
		if _, ok := rules[rule.ID]; ok {
			return Prepared{}, fmt.Errorf("duplicate category rule %s", rule.ID)
		}
		rules[rule.ID] = rule
		if rule.ID == "SELLER_SKU" && (rule.Hierarchy == "PARENT_PK" || rule.Hierarchy == "CHILD_PK") {
			return Prepared{}, fmt.Errorf("unsupported SKU hierarchy")
		}
		switch rule.Hierarchy {
		case "PARENT_PK":
			parents = append(parents, rule.ID)
		case "CHILD_PK":
			children = append(children, rule.ID)
		}
	}
	if len(rules) == 0 || (len(in.Variants) > 1 && len(children) == 0) {
		return Prepared{}, fmt.Errorf("category grouping rules unavailable")
	}
	sort.Strings(parents)
	sort.Strings(children)
	seenKeys, seenSKUs, seenChildren := map[string]bool{}, map[string]bool{}, map[string]bool{}
	var sharedParent string
	items := make([]up, 0, len(in.Variants))
	keys := make([]string, 0, len(in.Variants))
	for _, v := range in.Variants {
		if !textOK(v.Key, 120) || seenKeys[v.Key] || (v.Key == SingleKey && len(in.Variants) != 1) ||
			!textOK(v.SKU, 120) || strings.TrimSpace(v.SKU) != v.SKU || seenSKUs[v.SKU] ||
			v.Quantity == nil || *v.Quantity < 0 || !moneyPattern.MatchString(v.NetProceeds) ||
			strings.Trim(v.NetProceeds, "0.") == "" {
			return Prepared{}, fmt.Errorf("invalid or duplicate variant identity, SKU, quantity or net proceeds")
		}
		seenKeys[v.Key], seenSKUs[v.SKU] = true, true
		if len(v.PictureIDs) == 0 {
			return Prepared{}, fmt.Errorf("variant %s requires uploaded picture IDs", v.Key)
		}
		pics := make([]picture, 0, len(v.PictureIDs))
		seenPics := map[string]bool{}
		for _, id := range v.PictureIDs {
			if !picturePattern.MatchString(id) || seenPics[id] {
				return Prepared{}, fmt.Errorf("variant %s has invalid or duplicate picture ID", v.Key)
			}
			seenPics[id] = true
			pics = append(pics, picture{ID: id})
		}

		attrs := make(map[string]Attribute, len(v.Attributes)+1)
		for _, a := range v.Attributes {
			rule, known := rules[a.ID]
			if !known || a.ID == "SELLER_SKU" || a.ValueName != "" || len(a.Values) != 1 {
				return Prepared{}, fmt.Errorf("unsupported attribute %s", a.ID)
			}
			if _, exists := attrs[a.ID]; exists {
				return Prepared{}, fmt.Errorf("duplicate attribute %s", a.ID)
			}
			value := a.Values[0]
			if (value.ID == "" && !textOK(value.Name, 500)) ||
				(value.ID != "" && !textOK(value.ID, 120)) ||
				(value.Name != "" && !textOK(value.Name, 500)) ||
				(rule.Enumerated && value.ID == "") {
				return Prepared{}, fmt.Errorf("invalid attribute value %s", a.ID)
			}
			attrs[a.ID] = a
		}
		// SELLER_SKU is documented as an attribute in the UP create example.
		// Its actual account write/read contract remains a live acceptance gate.
		attrs["SELLER_SKU"] = Attribute{ID: "SELLER_SKU", ValueName: v.SKU}
		condition, ok := attrs["ITEM_CONDITION"]
		if !ok || condition.Values[0].ID != "2230284" {
			return Prepared{}, fmt.Errorf("new item condition must be explicitly confirmed")
		}
		for _, rule := range in.Rules {
			if rule.Required || rule.Hierarchy == "PARENT_PK" || (len(in.Variants) > 1 && rule.Hierarchy == "CHILD_PK") {
				if _, ok := attrs[rule.ID]; !ok {
					return Prepared{}, fmt.Errorf("missing required attribute %s", rule.ID)
				}
			}
		}
		parent := groupKey(attrs, parents)
		if len(items) == 0 {
			sharedParent = parent
		} else if parent != sharedParent {
			return Prepared{}, fmt.Errorf("PARENT_PK attributes differ")
		}
		child := groupKey(attrs, children)
		if len(in.Variants) > 1 && seenChildren[child] {
			return Prepared{}, fmt.Errorf("duplicate CHILD_PK combination")
		}
		seenChildren[child] = true
		ordered := make([]Attribute, 0, len(attrs))
		for _, a := range attrs {
			ordered = append(ordered, a)
		}
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
		items = append(items, up{
			FamilyName: in.FamilyName, CategoryID: in.CategoryID, Currency: "USD",
			NetProceeds: json.Number(v.NetProceeds), Quantity: *v.Quantity,
			Description: description{PlainText: in.Description}, Pictures: pics, Attributes: ordered,
			Sites: []site{{ID: t.Site, Logistic: t.Logistic, ListingType: t.ListingType}},
		})
		keys = append(keys, v.Key)
	}
	path := FamilyPath
	var body any = items
	if len(items) == 1 {
		path, body = SinglePath, items[0]
	}
	data, err := json.Marshal(body)
	if err != nil {
		return Prepared{}, err
	}
	return Prepared{Path: path, Body: data, VariantKeys: keys}, nil
}

// Use official value IDs where available; free text must already be normalized
// by the category adapter. JSON encoding keeps composite keys unambiguous.
func groupKey(attrs map[string]Attribute, ids []string) string {
	values := make([]Value, 0, len(ids))
	for _, id := range ids {
		var v Value
		if a, ok := attrs[id]; ok {
			v = a.Values[0]
			if v.ID != "" {
				v.Name = ""
			} else {
				v.Name = strings.ToLower(strings.Join(strings.Fields(v.Name), " "))
			}
		}
		values = append(values, v)
	}
	b, _ := json.Marshal(values)
	return string(b)
}
