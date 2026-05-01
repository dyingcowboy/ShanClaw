package tools

import (
	"context"
	"testing"
)

// TestThinkTool_AckNotEcho verifies the tool returns a short ack instead of
// echoing the thought back. The thought lives in the assistant message's
// tool_use.input.thought field — echoing into tool_result was double-counting
// it against cache. Build A of plan #8.
func TestThinkTool_AckNotEcho(t *testing.T) {
	tool := &ThinkTool{}

	longThought := "I should read the file first. Then check the imports. " +
		"Then look for the bug in the parser. Specifically, the case where " +
		"input is empty might trigger the panic we saw earlier."
	result, err := tool.Run(context.Background(), `{"thought":"`+longThought+`"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", result.Content)
	}
	if result.Content == longThought {
		t.Errorf("tool_result must NOT echo the thought (would double-count vs assistant tool_use.input)")
	}
	if result.Content != "thought logged" {
		t.Errorf("expected ack 'thought logged', got %q", result.Content)
	}
	if len(result.Content) > 50 {
		t.Errorf("ack must be short (~15B), got %d bytes", len(result.Content))
	}
}

func TestThinkTool_EmptyThought(t *testing.T) {
	tool := &ThinkTool{}

	result, err := tool.Run(context.Background(), `{"thought":""}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for empty thought")
	}
}

func TestThinkTool_InvalidJSON(t *testing.T) {
	tool := &ThinkTool{}

	result, err := tool.Run(context.Background(), `not json`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid JSON")
	}
}

func TestThinkTool_Info(t *testing.T) {
	tool := &ThinkTool{}
	info := tool.Info()

	if info.Name != "think" {
		t.Errorf("expected name 'think', got %q", info.Name)
	}
	if tool.RequiresApproval() {
		t.Error("think tool should not require approval")
	}
}
