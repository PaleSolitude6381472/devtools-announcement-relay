package announcement

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const DeveloperToolsChannel = "developer-tools"

type Kind string

const (
	BuildCompleted Kind = "build_completed"
	ReleaseStarted Kind = "release_started"
	Diagnostic     Kind = "diagnostic"
)

type Announcement struct {
	Kind      Kind   `json:"kind"`
	Summary   string `json:"summary"`
	Details   string `json:"details,omitempty"`
	AccountID string `json:"account_id,omitempty"`
}

type Publication struct {
	Channel        string
	Event          string
	Data           map[string]string
	AccountID      string
	IdempotencyKey string
}

func Prepare(a Announcement) (Publication, error) {
	events := map[Kind]string{
		BuildCompleted: "build.completed",
		ReleaseStarted: "release.started",
		Diagnostic:     "diagnostic.reported",
	}
	event, ok := events[a.Kind]
	if !ok {
		return Publication{}, errors.New("kind must be build_completed, release_started, or diagnostic")
	}
	if strings.TrimSpace(a.Summary) == "" {
		return Publication{}, errors.New("summary is required")
	}

	digest := sha256.Sum256([]byte(string(a.Kind) + "\x00" + a.Summary + "\x00" + a.Details + "\x00" + a.AccountID))
	return Publication{
		Channel: DeveloperToolsChannel,
		Event:   event,
		Data: map[string]string{
			"summary": a.Summary,
			"details": a.Details,
		},
		AccountID:      a.AccountID,
		IdempotencyKey: hex.EncodeToString(digest[:]),
	}, nil
}
