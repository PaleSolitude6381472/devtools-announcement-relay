package announcement

import "testing"

func TestPrepareChoosesDeveloperEvent(t *testing.T) {
	tests := []struct {
		name string
		kind Kind
		want string
	}{
		{name: "build result", kind: BuildCompleted, want: "build.completed"},
		{name: "release operation", kind: ReleaseStarted, want: "release.started"},
		{name: "developer diagnostic", kind: Diagnostic, want: "diagnostic.reported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Prepare(Announcement{Kind: tt.kind, Summary: "artifact is ready", Details: "revision a1b2c3"})
			if err != nil {
				t.Fatal(err)
			}
			if got.Channel != DeveloperToolsChannel || got.Event != tt.want {
				t.Fatalf("got channel %q event %q", got.Channel, got.Event)
			}
			again, _ := Prepare(Announcement{Kind: tt.kind, Summary: "artifact is ready", Details: "revision a1b2c3"})
			if got.IdempotencyKey == "" || got.IdempotencyKey != again.IdempotencyKey {
				t.Fatal("expected a stable idempotency key")
			}
		})
	}
}
