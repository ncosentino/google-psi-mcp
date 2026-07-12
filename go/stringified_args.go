package main

import (
	"context"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var toolArrayFields = map[string][]string{
	"analyze_page":     {"categories"},
	"analyze_pages":    {"urls", "categories"},
	"get_crux_data":    {"metrics"},
	"get_crux_history": {"metrics"},
}

func coerceStringifiedArrayArgs(arrayFieldsByTool map[string][]string) mcp.Middleware {
	return func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(
			ctx context.Context,
			method string,
			request mcp.Request,
		) (mcp.Result, error) {
			call, ok := request.(*mcp.CallToolRequest)
			if !ok || method != "tools/call" {
				return next(ctx, method, request)
			}
			fields := arrayFieldsByTool[call.Params.Name]
			if len(fields) == 0 || len(call.Params.Arguments) == 0 {
				return next(ctx, method, request)
			}

			var arguments map[string]json.RawMessage
			if err := json.Unmarshal(call.Params.Arguments, &arguments); err != nil {
				return next(ctx, method, request)
			}

			changed := false
			for _, field := range fields {
				if coerced, ok := coerceStringifiedArray(arguments[field]); ok {
					arguments[field] = coerced
					changed = true
				}
			}
			if changed {
				if rewritten, err := json.Marshal(arguments); err == nil {
					call.Params.Arguments = rewritten
				}
			}
			return next(ctx, method, request)
		}
	}
}

func coerceStringifiedArray(raw json.RawMessage) (json.RawMessage, bool) {
	if len(raw) == 0 {
		return nil, false
	}
	var encodedArray string
	if err := json.Unmarshal(raw, &encodedArray); err != nil {
		return nil, false
	}
	var array []json.RawMessage
	if err := json.Unmarshal([]byte(encodedArray), &array); err != nil {
		return nil, false
	}
	return json.RawMessage(encodedArray), true
}
