package presentation

import "testing"

func TestStateForUsesFixedWidthTextIndicators(t *testing.T) {
	tests := []struct {
		name      string
		done      bool
		important bool
		want      SparkState
	}{
		{"active", false, false, SparkState{"[ ]", "active", Muted}},
		{"important", false, true, SparkState{"[!]", "important", Important}},
		{"done", true, false, SparkState{"[x]", "done", Completed}},
		{"done takes precedence", true, true, SparkState{"[x]", "done", Completed}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StateFor(tt.done, tt.important)
			if got != tt.want {
				t.Fatalf("StateFor(%t, %t) = %#v, want %#v", tt.done, tt.important, got, tt.want)
			}
			if len(got.Indicator) != 3 {
				t.Fatalf("indicator %q has width %d, want 3", got.Indicator, len(got.Indicator))
			}
		})
	}
}
