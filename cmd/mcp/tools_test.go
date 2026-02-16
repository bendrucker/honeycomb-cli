package mcp

import (
	"context"
	"encoding/json"
	"testing"

	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/bendrucker/honeycomb-cli/cmd/options"
	"github.com/bendrucker/honeycomb-cli/internal/config"
	"github.com/bendrucker/honeycomb-cli/internal/iostreams"
	"github.com/zalando/go-keyring"
)

func init() {
	keyring.MockInit()
}

func testFactory(srv *server.MCPServer) clientFactory {
	return func(ctx context.Context, _, _ string) (*mcpclient.Client, error) {
		c, err := mcpclient.NewInProcessClient(srv)
		if err != nil {
			return nil, err
		}
		if err := c.Start(ctx); err != nil {
			return nil, err
		}
		_, err = c.Initialize(ctx, mcp.InitializeRequest{
			Params: mcp.InitializeParams{
				ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
				ClientInfo:      mcp.Implementation{Name: "test", Version: "0.0.0"},
			},
		})
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

func setupMCPTest(t *testing.T) (*server.MCPServer, *options.RootOptions, *iostreams.TestStreams) {
	t.Helper()

	srv := server.NewMCPServer("test-server", "1.0.0")

	ts := iostreams.Test(t)
	opts := &options.RootOptions{
		IOStreams: ts.IOStreams,
		Config:    &config.Config{},
		Format:    "json",
	}

	if err := config.SetKey("default", config.KeyConfig, "test-key"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.DeleteKey("default", config.KeyConfig) })

	return srv, opts, ts
}

func TestTools(t *testing.T) {
	srv, opts, ts := setupMCPTest(t)
	srv.AddTool(mcp.NewTool("run_query", mcp.WithDescription("Run a query")), nil)
	srv.AddTool(mcp.NewTool("get_columns", mcp.WithDescription("Get dataset columns")), nil)

	cmd := newToolsCmd(opts, testFactory(srv))
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	var tools []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(ts.OutBuf.Bytes(), &tools); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, ts.OutBuf.String())
	}
	if len(tools) != 2 {
		t.Fatalf("got %d tools, want 2", len(tools))
	}
	if tools[0].Name != "get_columns" && tools[1].Name != "get_columns" {
		t.Errorf("expected get_columns tool in output")
	}
}
