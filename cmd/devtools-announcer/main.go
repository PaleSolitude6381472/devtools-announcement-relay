package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	announcement "example.com/devtools-announcement-relay"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	client := &announcement.Client{APIKey: key}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /announcements", func(w http.ResponseWriter, r *http.Request) {
		var input announcement.Announcement
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid announcement"})
			return
		}
		publication, err := announcement.Prepare(input)
		if err != nil {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
			return
		}
		if err := client.Publish(r.Context(), publication); err != nil {
			var apiErr *announcement.APIError
			if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
				writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Error()})
				return
			}
			log.Printf("publish failed: %v", err)
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "announcement delivery failed"})
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]string{
			"status":  "published",
			"channel": publication.Channel,
			"event":   publication.Event,
		})
	})

	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("devtools announcer listening on %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}
