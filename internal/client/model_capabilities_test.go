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

		// Tier-only resolution: medium/big/large default to 1M (priority 1
		// happy-path on Cloud is sonnet-4-6 1M auto for medium; opus-4-6 1M
		// auto for large). small stays 200K because haiku-4-5 has no 1M
		// variant and failover chain would 400 against haiku's actual cap.
		// Failover from medium/large priority 1 is handled by the reactive
		// recovery layer (single 400 + summary cap + retry).
		// "large" is Shannon Cloud nomenclature, "big" is ShanClaw's — both
		// accepted as aliases.
		{"big tier (1M happy-path)", "", "big", 1_000_000},
		{"large tier (Cloud nomenclature)", "", "large", 1_000_000},
		{"medium tier (1M happy-path)", "", "medium", 1_000_000},
		{"small tier (haiku 200K cap)", "", "small", 200_000},

		// Specific takes precedence over tier — a known 200K-cap model still
		// resolves to 200K even when paired with a 1M tier.
		{"specific wins over tier (haiku + big)", "claude-haiku-4-5", "big", 200_000},

		// Unknown / empty → conservative default.
		{"empty both", "", "", 200_000},
		{"unknown tier", "", "huge", 200_000},
		{"future model", "claude-opus-5-0", "", 200_000},

		// Regression: specific model that lookupSpecificModel does not
		// recognize must NOT fall back to tier. Operator pinned a model
		// we don't know; conservative 200K is the safe default. (See
		// 2026-05-08 review Finding 2 — earlier impl returned 1M here
		// via the tier fallback, defeating the resolver on operator
		// pins of newly-released models that hadn't been added to the
		// prefix table.)
		{"unknown specific + medium tier", "claude-flux-9000", "medium", 200_000},
		{"unknown specific + big tier", "claude-future-1-0", "big", 200_000},
		{"unknown specific + small tier", "claude-future-1-0", "small", 200_000},
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
