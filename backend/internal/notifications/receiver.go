// Package notifications owns the inbound Mercado Libre CBT event inbox.
package notifications

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const Path = "/api/v1/mercadolibre/notifications"
const ingressHeader = "X-Una-Webhook-Ingress"
const maxBody = 16 << 10

var (
	ErrScope            = errors.New("notification scope not uniquely authorized")
	ErrScopeUnavailable = errors.New("known notification scope temporarily unavailable")
	idPattern           = regexp.MustCompile(`^[1-9][0-9]{0,29}$`)
	messageIDPattern    = regexp.MustCompile(`^[A-Za-z0-9]{16,64}$`)
)

// Keep the receiver and activation registry on the same fixed allowlist.
var topicPaths = map[string]string{
	"marketplace_orders":    "/marketplace/orders/",
	"marketplace_shipments": "/marketplace/shipments/",
	"marketplace_questions": "/marketplace/questions/",
	"marketplace_messages":  "/marketplace/messages/",
}

func ValidResourceID(topic, resource string) bool {
	if _, ok := topicPaths[topic]; !ok {
		return false
	}
	if topic == "marketplace_messages" {
		return messageIDPattern.MatchString(resource)
	}
	return idPattern.MatchString(resource)
}

// Event is an allowlisted hint, never a platform URL or a business update.
type Event struct {
	ApplicationID string
	UserID        string
	Topic         string
	ResourceID    string
	Key           string
	ReceivedAt    time.Time
}

type Sink interface {
	Enqueue(context.Context, Event) error
}

// IDs are strings or integer JSON literals; never float64.
func id(raw json.RawMessage) string {
	s := string(raw)
	if len(s) > 0 && s[0] == '"' {
		if json.Unmarshal(raw, &s) != nil {
			return ""
		}
	}
	if !idPattern.MatchString(s) {
		return ""
	}
	return s
}

func parse(r io.Reader, appID string) (Event, error) {
	var in struct {
		ID            string          `json:"_id"`
		ApplicationID json.RawMessage `json:"application_id"`
		UserID        json.RawMessage `json:"user_id"`
		Topic         string          `json:"topic"`
		Resource      string          `json:"resource"`
		Received      string          `json:"received"`
	}
	bad := errors.New("invalid notification")
	d := json.NewDecoder(r)
	if d.Decode(&in) != nil || d.Decode(new(any)) != io.EOF {
		return Event{}, bad
	}
	a, u := id(in.ApplicationID), id(in.UserID)
	if a == "" || a != appID || u == "" {
		return Event{}, bad
	}
	prefix, supported := topicPaths[in.Topic]
	resource, ok := strings.CutPrefix(in.Resource, prefix)
	if !supported || !ok || !ValidResourceID(in.Topic, resource) {
		return Event{}, bad
	}
	received, err := time.Parse(time.RFC3339Nano, in.Received)
	if err != nil || received.IsZero() {
		return Event{}, bad
	}
	// Retry counters and sent timestamps can vary per delivery. An official ID
	// is preferred; only when absent use the original received timestamp.
	identity := "received:" + received.UTC().Format(time.RFC3339Nano)
	if in.ID != "" {
		// Opaque official ID: bounded by the body limit, hashed and never stored raw.
		identity = "id:" + in.ID
	}
	hash := sha256.Sum256([]byte(identity))
	return Event{ApplicationID: a, UserID: u, Topic: in.Topic, ResourceID: resource,
		Key: hex.EncodeToString(hash[:]), ReceivedAt: received.UTC()}, nil
}

// NewReceiver requires a private ingress token injected by a trusted proxy
// AFTER its verified source gate. It is NOT a Mercado Libre signature.
// Never expose this header to untrusted clients or enable before source review.
func NewReceiver(appID, ingressToken string, sink Sink) (http.Handler, error) {
	if !idPattern.MatchString(appID) || len(ingressToken) < 32 || len(ingressToken) > 256 || sink == nil {
		return nil, errors.New("notification receiver configuration invalid")
	}
	expected := sha256.Sum256([]byte(ingressToken))
	slots := make(chan struct{}, 8)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodPost {
			w.Header().Set("Connection", "close")
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		provided := sha256.Sum256([]byte(r.Header.Get(ingressHeader)))
		if subtle.ConstantTimeCompare(expected[:], provided[:]) != 1 {
			w.Header().Set("Connection", "close")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			w.Header().Set("Connection", "close")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Budget body reads and persistence together, with headroom for the proxy.
		ctx, cancel := context.WithTimeout(r.Context(), 350*time.Millisecond)
		defer cancel()
		controller := http.NewResponseController(w)
		deadline, _ := ctx.Deadline()
		if controller.SetReadDeadline(deadline) == nil {
			defer controller.SetReadDeadline(time.Time{})
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxBody)
		event, err := parse(r.Body, appID)
		if err != nil {
			w.Header().Set("Connection", "close")
			if ctx.Err() != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := sink.Enqueue(ctx, event); err != nil {
			if errors.Is(err, ErrScope) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}), nil
}
