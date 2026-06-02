package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
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

	cmd := newCallCmd(opts, nil, testFactory(srv))
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

func TestCall_TypedJSONObjectArgument(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)

	var got map[string]any
	srv.AddTool(
		mcp.NewTool("run_query",
			mcp.WithString("dataset"),
			mcp.WithObject("query_json"),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			got = req.GetArguments()
			return mcp.NewToolResultText("ok"), nil
		},
	)

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"run_query",
		"-f", "dataset=api",
		"-F", `query_json={"calculations":[{"op":"SUM","column":"gen_ai.usage.cost"}],"time_range":604800,"granularity":86400}`,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if ds, ok := got["dataset"].(string); !ok || ds != "api" {
		t.Errorf("dataset = %#v, want string %q", got["dataset"], "api")
	}

	query, ok := got["query_json"].(map[string]any)
	if !ok {
		t.Fatalf("query_json = %#v, want map[string]any (a parsed object, not a string)", got["query_json"])
	}
	if tr, ok := query["time_range"].(float64); !ok || tr != 604800 {
		t.Errorf("query_json.time_range = %#v, want 604800", query["time_range"])
	}
	calcs, ok := query["calculations"].([]any)
	if !ok || len(calcs) != 1 {
		t.Fatalf("query_json.calculations = %#v, want one calculation", query["calculations"])
	}
	calc, ok := calcs[0].(map[string]any)
	if !ok || calc["op"] != "SUM" || calc["column"] != "gen_ai.usage.cost" {
		t.Errorf("query_json.calculations[0] = %#v, want SUM(gen_ai.usage.cost)", calcs[0])
	}
}

func TestCall_SchemaCoercesRawField(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)

	var got map[string]any
	srv.AddTool(
		mcp.NewTool("run_query",
			mcp.WithString("dataset"),
			mcp.WithObject("query_json"),
		),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			got = req.GetArguments()
			return mcp.NewToolResultText("ok"), nil
		},
	)

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"run_query",
		"-f", "dataset=api",
		"-f", `query_json={"calculations":[{"op":"SUM","column":"gen_ai.usage.cost"}],"time_range":604800}`,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	query, ok := got["query_json"].(map[string]any)
	if !ok {
		t.Fatalf("query_json = %#v, want a parsed object (-f coerced via schema, not left a string)", got["query_json"])
	}
	if tr, ok := query["time_range"].(float64); !ok || tr != 604800 {
		t.Errorf("query_json.time_range = %#v, want 604800", query["time_range"])
	}
	if ds, ok := got["dataset"].(string); !ok || ds != "api" {
		t.Errorf("dataset = %#v, want string %q", got["dataset"], "api")
	}
}

func TestCall_UnknownTool(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)
	srv.AddTool(mcp.NewTool("known"), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mcp.NewToolResultText("ok"), nil
	})

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"missing", "-f", "key=val"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
	if !strings.Contains(err.Error(), `unknown tool "missing"`) {
		t.Errorf("error = %q, want unknown tool message", err.Error())
	}
}

func TestCall_InputNotCoerced(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)

	var got map[string]any
	srv.AddTool(
		mcp.NewTool("run_query", mcp.WithObject("query_json")),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			got = req.GetArguments()
			return mcp.NewToolResultText("ok"), nil
		},
	)

	dir := t.TempDir()
	path := filepath.Join(dir, "args.json")
	if err := os.WriteFile(path, []byte(`{"query_json":"raw-string"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"run_query", "--input", path})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if got["query_json"] != "raw-string" {
		t.Errorf("query_json = %#v, want the string sent as-is (--input is not schema-coerced)", got["query_json"])
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

	cmd := newCallCmd(opts, nil, testFactory(srv))
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

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"data", "--jq", ".content[0].text"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(ts.OutBuf.String())
	if got != `{"key":"value"}` {
		t.Errorf("output = %q, want the text content", got)
	}
}

func TestCall_Table(t *testing.T) {
	srv, opts, ts := setupMCPTest(t)
	opts.Format = "table"
	srv.AddTool(
		mcp.NewTool("echo", mcp.WithString("message")),
		func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			msg, _ := req.GetArguments()["message"].(string)
			return mcp.NewToolResultText(msg), nil
		},
	)

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"echo", "-f", "message=hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	out := ts.OutBuf.String()
	for _, want := range []string{"INDEX", "TYPE", "CONTENT", "text", "hello"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q\n%s", want, out)
		}
	}
}

func TestCall_NonTextContentJSON(t *testing.T) {
	srv, opts, ts := setupMCPTest(t)
	srv.AddTool(
		mcp.NewTool("image"),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultImage("here", "ZGF0YQ==", "image/png"), nil
		},
	)

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"image"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var result struct {
		Content []struct {
			Type     string `json:"type"`
			Text     string `json:"text"`
			Data     string `json:"data"`
			MIMEType string `json:"mimeType"`
		} `json:"content"`
		IsError bool `json:"is_error"`
	}
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, ts.OutBuf.String())
	}
	if len(result.Content) != 2 {
		t.Fatalf("got %d content items, want 2 (text + image)\n%s", len(result.Content), ts.OutBuf.String())
	}
	img := result.Content[1]
	if img.Type != "image" {
		t.Errorf("content[1].type = %q, want image", img.Type)
	}
	if img.Data != "ZGF0YQ==" {
		t.Errorf("content[1].data = %q, want base64 image data", img.Data)
	}
	if img.MIMEType != "image/png" {
		t.Errorf("content[1].mimeType = %q, want image/png", img.MIMEType)
	}
	if strings.Contains(ts.OutBuf.String(), "isError") {
		t.Errorf("envelope used camelCase isError, want is_error\n%s", ts.OutBuf.String())
	}
}

func TestCall_NonTextContentTable(t *testing.T) {
	srv, opts, ts := setupMCPTest(t)
	opts.Format = "table"
	srv.AddTool(
		mcp.NewTool("image"),
		func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultImage("here", "ZGF0YQ==", "image/png"), nil
		},
	)

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"image"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	out := ts.OutBuf.String()
	for _, want := range []string{"image", "[image image/png]", "here"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q\n%s", want, out)
		}
	}
}

func TestContentSummary(t *testing.T) {
	for _, tc := range []struct {
		name     string
		content  mcp.Content
		wantType string
		wantText string
	}{
		{
			name:     "text",
			content:  mcp.TextContent{Type: "text", Text: "hello"},
			wantType: "text",
			wantText: "hello",
		},
		{
			name:     "image",
			content:  mcp.ImageContent{Type: "image", MIMEType: "image/png"},
			wantType: "image",
			wantText: "[image image/png]",
		},
		{
			name:     "audio",
			content:  mcp.AudioContent{Type: "audio", MIMEType: "audio/wav"},
			wantType: "audio",
			wantText: "[audio audio/wav]",
		},
		{
			name:     "resource link",
			content:  mcp.ResourceLink{Type: "resource_link", URI: "hny://dataset/x"},
			wantType: "resource_link",
			wantText: "[resource_link hny://dataset/x]",
		},
		{
			name: "embedded resource",
			content: mcp.EmbeddedResource{
				Type:     "resource",
				Resource: mcp.TextResourceContents{URI: "hny://query/1"},
			},
			wantType: "resource",
			wantText: "[resource hny://query/1]",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := contentType(tc.content); got != tc.wantType {
				t.Errorf("contentType = %q, want %q", got, tc.wantType)
			}
			if got := contentSummary(tc.content); got != tc.wantText {
				t.Errorf("contentSummary = %q, want %q", got, tc.wantText)
			}
		})
	}
}

func TestCall_MutuallyExclusiveFlags(t *testing.T) {
	srv, opts, _ := setupMCPTest(t)
	srv.AddTool(mcp.NewTool("test"), nil)

	cmd := newCallCmd(opts, nil, testFactory(srv))
	cmd.SetArgs([]string{"test", "-f", "key=val", "--input", "file.json"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for mutually exclusive flags")
	}
	if !strings.Contains(err.Error(), "cannot use") {
		t.Errorf("error = %q, want mutual exclusion message", err.Error())
	}
}
