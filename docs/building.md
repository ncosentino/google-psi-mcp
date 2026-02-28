# Building from Source

Build the Go or C# implementation locally from source.

---

## Go

Requires **Go 1.26+**. Install from [go.dev/dl](https://go.dev/dl).

```bash
cd go
go mod tidy
go build -ldflags="-s -w" -trimpath -o psi-mcp-go .
```

### Cross-compile

```bash
# Linux amd64 (from any platform)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o psi-mcp-go-linux-amd64 .

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o psi-mcp-go-darwin-arm64 .

# Windows amd64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -o psi-mcp-go-windows-amd64.exe .
```

### Run tests

```bash
go test ./...
```

---

## C# (Native AOT)

Requires the [.NET 10 SDK](https://dotnet.microsoft.com/download/dotnet/10.0).

```bash
cd csharp

# Development build
dotnet build PageSpeedMcp.slnx

# Publish Native AOT (no runtime needed -- matches release binaries)
dotnet publish src/PageSpeedMcp/PageSpeedMcp.csproj -r linux-x64 -c Release --self-contained true
```

### Runtime identifier (RID) reference

| Platform | RID |
|----------|-----|
| Linux x64 | `linux-x64` |
| Linux ARM64 | `linux-arm64` |
| macOS x64 | `osx-x64` |
| macOS ARM64 | `osx-arm64` |
| Windows x64 | `win-x64` |
| Windows ARM64 | `win-arm64` |

### Run tests

```bash
cd csharp
dotnet test PageSpeedMcp.slnx
```

!!! note "Native AOT toolchain requirements"
    Publishing Native AOT requires platform-specific native tooling: `clang` on Linux/macOS, MSVC build tools on Windows. The standard `dotnet build` works for development without those tools.

