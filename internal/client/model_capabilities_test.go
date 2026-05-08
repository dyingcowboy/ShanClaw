package client

import "testing"

func TestResolveModelCapabilities(t *testing.T) {
	cases := []struct {
		name          string
		specificModel string
		modelTier     string
		wantWindow    int
	}{
		// 1M-capable models (auto-1M as of 2026-03-13, no beta header).
		{"opus-4-7 specific", "claude-opus-4-7", "", 1_000_000},
		{"opus-4-7 dated", "claude-opus-4-7-20260416", "", 1_000_000},
		{"opus-4-6 specific", "claude-opus-4-6", "", 1_000_000},
		{"opus-4-6 dated", "claude-opus-4-6-20260205", "", 1_000_000},
		{"sonnet-4-6 specific", "claude-sonnet-4-6", "", 1_000_000},
		{"sonnet-4-6 dated", "claude-sonnet-4-6-20260217", "", 1_000_000},

		// 200K-capped models (beta header retired 2026-04-30).
		{"sonnet-4-5", "claude-sonnet-4-5", "", 200_000},
		{"sonnet-4-5 dated", "claude-sonnet-4-5-20250929", "", 200_000},
		{"sonnet-4 legacy", "claude-sonnet-4-20250514", "", 200_000},
		{"haiku-4-5", "claude-haiku-4-5", "", 200_000},
		{"haiku-4-5 dated", "claude-haiku-4-5-20251001", "", 200_000},

		// Tier fallback when specific model is empty.
		{"big tier", "", "big", 1_000_000},
		{"medium tier", "", "medium", 1_000_000},
		{"small tier", "", "small", 200_000},

		// Specific takes precedence over tier.
		{"specific wins over tier", "claude-haiku-4-5", "big", 200_000},

		// Unknown / empty → conservative default.
		{"empty both", "", "", 200_000},
		{"unknown tier", "", "huge", 200_000},
		{"future model", "claude-opus-5-0", "", 200_000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveModelCapabilities(tc.specificModel, tc.modelTier)
			if got.ContextWindow != tc.wantWindow {
				t.Errorf("ContextWindow = %d, want %d", got.ContextWindow, tc.wantWindow)
			}
		})
	}
}
