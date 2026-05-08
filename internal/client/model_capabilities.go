package client

import "strings"

// ModelCapabilities describes per-model API constraints relevant to ShanClaw's
// context-window management and request building.
type ModelCapabilities struct {
	// ContextWindow is the model's maximum input+output token capacity.
	// For models with auto-1M (Opus 4.6+, Sonnet 4.6) this is 1_000_000.
	// For 200K-capped models this is 200_000.
	ContextWindow int
}

// ResolveModelCapabilities returns the API capabilities for a model.
//
// Precedence:
//  1. specificModel matched against known prefixes → that model's window
//  2. specificModel set but unrecognized → conservative 200K (NOT a tier
//     fallback, because mismatching tier on an unknown specific model
//     would silently widen the cap and risk a 200K-cap model hitting the
//     1M assumption — that's the failure mode this resolver exists to prevent)
//  3. specificModel empty + modelTier matches a known tier → tier window
//  4. both empty / unknown tier → conservative 200K default
//
// Source of truth for these caps:
//   - 2026-03-13 release: 1M context GA for Opus 4.6 / Sonnet 4.6 (no header).
//   - 2026-04-16 release: Opus 4.7 launched, same 1M behavior.
//   - 2026-04-30 release: context-1m-2025-08-07 beta retired for Sonnet 4.5/4
//     (header now no-op).
func ResolveModelCapabilities(specificModel, modelTier string) ModelCapabilities {
	if specificModel != "" {
		if caps, ok := lookupSpecificModel(specificModel); ok {
			return caps
		}
		// Specific model named but not in our table — never speculate
		// upward via tier fallback. Operator pinned a model we don't
		// recognize; assume the conservative 200K cap so the agent loop
		// triggers compaction at a safe boundary.
		return ModelCapabilities{ContextWindow: 200_000}
	}
	if caps, ok := lookupModelTier(modelTier); ok {
		return caps
	}
	return ModelCapabilities{ContextWindow: 200_000}
}

// Prefixes for models that auto-support 1M context (no beta header required).
// Match by prefix so dated variants ("claude-opus-4-7-20260416") are covered.
var prefixes1M = []string{
	"claude-opus-4-7",
	"claude-opus-4-6",
	"claude-sonnet-4-6",
}

// Prefixes for models confirmed at 200K. Non-exhaustive — anything else
// falls through to the 200K default, which is the safe direction.
var prefixes200K = []string{
	"claude-sonnet-4-5",
	"claude-sonnet-4-",
	"claude-haiku-4-5",
}

func lookupSpecificModel(model string) (ModelCapabilities, bool) {
	if model == "" {
		return ModelCapabilities{}, false
	}
	for _, p := range prefixes1M {
		if strings.HasPrefix(model, p) {
			return ModelCapabilities{ContextWindow: 1_000_000}, true
		}
	}
	for _, p := range prefixes200K {
		if strings.HasPrefix(model, p) {
			return ModelCapabilities{ContextWindow: 200_000}, true
		}
	}
	return ModelCapabilities{}, false
}

func lookupModelTier(tier string) (ModelCapabilities, bool) {
	switch tier {
	case "big", "medium":
		return ModelCapabilities{ContextWindow: 1_000_000}, true
	case "small":
		return ModelCapabilities{ContextWindow: 200_000}, true
	}
	return ModelCapabilities{}, false
}
