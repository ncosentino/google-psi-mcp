---
description: Build, test, lint, and publish the Go and C# implementations.
---

# Building from Source

## Go

Requires Go 1.26 or newer.

```bash
cd go
go test ./...
go vet ./...
golangci-lint run
go build -ldflags="-s -w -X main.version=dev" -trimpath -o psi-mcp-go .
```

Cross-compile by setting `GOOS`, `GOARCH`, and `CGO_ENABLED=0`.

## C# Native AOT

Requires the .NET 10 SDK and the platform's native linker toolchain.

```bash
cd csharp
dotnet test PageSpeedMcp.slnx -c Release
dotnet publish src/PageSpeedMcp/PageSpeedMcp.csproj \
  -r linux-x64 \
  -c Release \
  --self-contained true
```

| Platform | RID |
|---|---|
| Linux x64 | `linux-x64` |
| Linux ARM64 | `linux-arm64` |
| macOS Intel | `osx-x64` |
| macOS Apple Silicon | `osx-arm64` |
| Windows x64 | `win-x64` |
| Windows ARM64 | `win-arm64` |

## Running HTTP locally

```bash
cd go
go run . --transport http --listen-address 127.0.0.1 --port 8080 \
  --api-key YOUR_KEY
```

```bash
cd csharp
dotnet run --project src/PageSpeedMcp -- \
  --transport http --listen-address 127.0.0.1 --port 8080 \
  --api-key YOUR_KEY
```

Both commands expose MCP at `/mcp` and health metadata at `/health`.

## Documentation-site descriptions

MkDocs derives descriptions from each page's title and introductory prose, excluding
notes, navigation, images, code blocks and tables. Missing prose receives a truthful
site/page identity fallback; the homepage retains its project description.
Existing nonempty `description:` frontmatter takes precedence, but no new handwritten
descriptions are required. Maintain the ordinary documentation source instead.

Generated summaries target 240 characters at word boundaries; explicit descriptions
are not truncated. Invalid non-string descriptions fail the build. HTML and social
rendering escape values while source metadata remains plain text for JSON-LD and
`llms.txt`. Hosted documentation validation exercises the real templates and preserves
headings and content. No runtime AI or separate metadata authoring is involved.
