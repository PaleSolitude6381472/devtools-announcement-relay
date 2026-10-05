package announcement

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublishRetriesRateLimitWithoutChangingRequest(t *testing.T) {
	requests := 0
	var firstKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing bearer authorization")
		}
		if requests == 1 {
			firstKey = r.Header.Get("Idempotency-Key")
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": map[string]string{"message": "retry later"}})
			return
		}
		if r.Header.Get("Idempotency-Key") != firstKey {
			t.Error("idempotency key changed between attempts")
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["channel"] != DeveloperToolsChannel || body["event"] != "release.started" {
			t.Errorf("unexpected body: %#v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": map[string]string{"accepted": "yes"}, "metadata": map[string]string{}})
	}))
	defer server.Close()

	var delay time.Duration
	client := Client{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Sleep: func(_ context.Context, d time.Duration) error {
			delay = d
			return nil
		},
	}
	publication, _ := Prepare(Announcement{Kind: ReleaseStarted, Summary: "release 42 started"})
	if err := client.Publish(context.Background(), publication); err != nil {
		t.Fatal(err)
	}
	if requests != 2 || delay != 2*time.Second {
		t.Fatalf("requests = %d, delay = %s", requests, delay)
	}
}
