package main

import (
	"context"
	"io"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestStdioTransport_ServesRealSession(t *testing.T) {
	t.Parallel()

	srv := newServer(&trackingAnalyzer{}, fakeCruxQuerier{})
	serverRead, clientWrite := io.Pipe()
	clientRead, serverWrite := io.Pipe()
	ctx := context.Background()

	serverSession, err := srv.Connect(
		ctx,
		&mcp.IOTransport{Reader: serverRead, Writer: serverWrite},
		nil,
	)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	defer serverSession.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "test"}, nil)
	clientSession, err := client.Connect(
		ctx,
		&mcp.IOTransport{Reader: clientRead, Writer: clientWrite},
		nil,
	)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	defer clientSession.Close()

	tools, err := clientSession.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	if len(tools.Tools) != 4 {
		t.Errorf("tools = %d, want 4", len(tools.Tools))
	}
}
