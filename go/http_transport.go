package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
	"slices"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func runHTTP(srv *mcp.Server, allowedHosts []string) {
	port := resolveHTTPPort()
	handler := buildHTTPHandler(srv, allowedHosts)

	slog.Info(
		"google-psi-mcp starting",
		"version",
		version,
		"transport",
		"http",
		"port",
		port,
		"allowed_hosts",
		allowedHosts,
	)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

func buildHTTPHandler(srv *mcp.Server, allowedHosts []string) http.Handler {
	handler := mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server {
			return srv
		},
		&mcp.StreamableHTTPOptions{Stateless: true},
	)
	protection := http.NewCrossOriginProtection()
	return allowedHostsMiddleware(protection.Handler(handler), allowedHosts)
}

func resolveHTTPPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return "8080"
}

func allowedHostsMiddleware(next http.Handler, allowedHosts []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		host := request.Host
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			host = parsedHost
		}
		if !slices.Contains(allowedHosts, host) {
			http.Error(w, "host not allowed", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, request)
	})
}
