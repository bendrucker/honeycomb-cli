package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCall(t *testing.T) {
	srv, opts, ts := setupMCPTest(t)
	srv.AddTool(
		mcp.NewTool("echo",
			mcp.WithDescription("Echo arguments"),
			mcp.WithString("message"),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			msg, _ := req.GetArguments()["message"].(string)
			return mcp.NewToolResultText(msg), nil
		},
	)

	cmd := newCallCmd(opts, testFactory(srv))
	cmd.SetArgs([]string{"echo", "-f", "message=hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, ts.OutBuf.String())
	}
	if len(result.Content) == 0 {
		t.Fatal("no content in result")
	}
	if result.Content[0].Text != "hello" {
		t.Errorf("text = %q, want %q", result.Content[0].Text, "hello")
	}
}

func TestCall_ToolError(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)
	srv.AddTool(
		mcp.NewTool("fail"),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultError("something went wrong"), nil
		},
	)

	cmd := newCallCmd(opts, testFactory(srv))
	cmd.SetArgs([]string{"fail"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "something went wrong") {
		t.Errorf("error = %q, want tool error message", err.Error())
	}
}

func TestCall_JQ(t *testing.T) {
	srv, opts, ts := setupMCPTest(t)
	srv.AddTool(
		mcp.NewTool("data"),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText(`{"key":"value"}`), nil
		},
	)

	cmd := newCallCmd(opts, testFactory(srv))
	cmd.SetArgs([]string{"data", "--jq", ".content[0].text"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(ts.OutBuf.String())
	if got != `{"key":"value"}` {
		t.Errorf("output = %q, want the text content", got)
	}
}

func TestCall_MutuallyExclusiveFlags(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)
	srv.AddTool(mcp.NewTool("test"), nil)

	cmd := newCallCmd(opts, testFactory(srv))
	cmd.SetArgs([]string{"test", "-f", "key=val", "--input", "file.json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "cannot use") {
		t.Errorf("error = %q, want mutual exclusion message", err.Error())
	}
}
