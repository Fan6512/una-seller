package cbtcategory

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	DumpURL       = "https://api.mercadolibre.com/sites/CBT/categories/all"
	maxCompressed = 16 << 20
	maxJSON       = 64 << 20
	maxCategories = 100_000
)

var idPattern = regexp.MustCompile("^CBT[0-9]{1,15}$")

type PathNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type version struct {
	MD5     string `json:"md5"`
	Created string `json:"created"`
}

type category struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	ParentID       string     `json:"parent_id,omitempty"`
	Path           []PathNode `json:"path"`
	Depth          int        `json:"depth"`
	HasChildren    bool       `json:"has_children"`
	ListingAllowed bool       `json:"listing_allowed"`
	Status         string     `json:"status"`
	Zh             string     `json:"zh,omitempty"`
}

type snapshot struct {
	Version    version               `json:"version"`
	Categories map[string]category   `json:"categories"`
	Retired    map[string][]category `json:"retired,omitempty"`
	Skipped    map[string]string     `json:"skipped,omitempty"`
}

type dumpNode struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Path     []PathNode `json:"path_from_root"`
	Children []PathNode `json:"children_categories"`
	Settings struct {
		ListingAllowed *bool  `json:"listing_allowed"`
		Status         string `json:"status"`
	} `json:"settings"`
}

func fetchDump(ctx context.Context, client *http.Client, endpoint string) (*http.Response, version, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, version{}, errors.New("invalid CBT dump request")
	}
	// An explicit encoding keeps Go from decompressing before checksum validation.
	req.Header.Set("Accept-Encoding", "gzip")
	resp, err := client.Do(req)
	if err != nil {
		return nil, version{}, errors.New("failed to fetch official CBT category dump")
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, version{}, fmt.Errorf("official CBT category dump returned HTTP %d", resp.StatusCode)
	}
	v := version{
		MD5:     strings.ToLower(resp.Header.Get("X-Content-MD5")),
		Created: resp.Header.Get("X-Content-Created"),
	}
	if !v.valid() || !strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		resp.Body.Close()
		return nil, version{}, errors.New("official CBT category dump is missing valid gzip/version headers")
	}
	return resp, v, nil
}

func (v version) valid() bool {
	decoded, err := hex.DecodeString(v.MD5)
	if err != nil || len(decoded) != md5.Size {
		return false
	}
	_, err = time.Parse(time.RFC3339Nano, v.Created)
	return err == nil
}

func readDump(body io.Reader, v version) (snapshot, error) {
	compressed, err := io.ReadAll(io.LimitReader(body, maxCompressed+1))
	if err != nil || len(compressed) > maxCompressed {
		return snapshot{}, errors.New("official CBT category gzip exceeds size limit or could not be read")
	}
	// X-Content-MD5 identifies the exact gzip bytes published by Mercado Libre.
	sum := md5.Sum(compressed)
	if hex.EncodeToString(sum[:]) != v.MD5 {
		return snapshot{}, errors.New("official CBT category gzip checksum mismatch")
	}
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return snapshot{}, errors.New("official CBT category dump is not gzip")
	}
	defer gz.Close()
	plain, err := io.ReadAll(io.LimitReader(gz, maxJSON+1))
	if err != nil || len(plain) > maxJSON || !utf8.Valid(plain) {
		return snapshot{}, errors.New("official CBT category JSON exceeds size limit or could not be decoded")
	}
	return parseDump(plain, v)
}

func parseDump(plain []byte, v version) (snapshot, error) {
	dec := json.NewDecoder(bytes.NewReader(plain))
	if token, err := dec.Token(); err != nil || token != json.Delim('{') {
		return snapshot{}, errors.New("official CBT dump must be a category-ID object")
	}
	nodes := make(map[string]dumpNode)
	for dec.More() {
		token, err := dec.Token()
		id, ok := token.(string)
		if err != nil || !ok || !idPattern.MatchString(id) || len(nodes) >= maxCategories {
			return snapshot{}, errors.New("invalid category ID or count in official CBT dump")
		}
		if _, found := nodes[id]; found {
			return snapshot{}, fmt.Errorf("duplicate category ID %s", id)
		}
		var node dumpNode
		if err := dec.Decode(&node); err != nil || node.ID != id || !validName(node.Name) ||
			len(node.Path) == 0 || len(node.Path) > 32 || node.Children == nil ||
			node.Settings.ListingAllowed == nil || (node.Settings.Status != "enabled" && node.Settings.Status != "disabled") ||
			node.Path[len(node.Path)-1] != (PathNode{ID: id, Name: node.Name}) {
			return snapshot{}, fmt.Errorf("invalid official category %s", id)
		}
		nodes[id] = node
	}
	if _, err := dec.Token(); err != nil || dec.Decode(new(any)) != io.EOF || len(nodes) == 0 {
		return snapshot{}, errors.New("empty or incomplete official CBT dump")
	}

	out := snapshot{Version: v, Categories: make(map[string]category, len(nodes))}
	for id, node := range nodes {
		parent := ""
		if len(node.Path) > 1 {
			parent = node.Path[len(node.Path)-2].ID
			p, found := nodes[parent]
			if !found || !slices.Equal(node.Path[:len(node.Path)-1], p.Path) ||
				!slices.Contains(p.Children, PathNode{ID: id, Name: node.Name}) {
				return snapshot{}, fmt.Errorf("invalid parent or path for %s", id)
			}
		}
		children := make(map[string]bool, len(node.Children))
		for _, child := range node.Children {
			c, found := nodes[child.ID]
			if !found || c.Name != child.Name || len(c.Path) != len(node.Path)+1 ||
				!slices.Equal(node.Path, c.Path[:len(c.Path)-1]) || children[child.ID] {
				return snapshot{}, fmt.Errorf("invalid child of %s", id)
			}
			children[child.ID] = true
		}
		out.Categories[id] = category{
			ID: id, Name: node.Name, ParentID: parent, Path: node.Path, Depth: len(node.Path),
			HasChildren: len(node.Children) > 0, Status: node.Settings.Status, ListingAllowed: *node.Settings.ListingAllowed,
		}
	}
	return out, nil
}

func validName(value string) bool {
	return len(value) <= 2000 && utf8.ValidString(value) && strings.TrimSpace(value) != "" && !strings.ContainsAny(value, "\r\n")
}
