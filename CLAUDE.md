# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run

```bash
# Build
go build -o main .

# Run
./main
# Requires /root directory to exist (auto-created if not present)

# Access
# Admin panel: http://localhost:8081/static/
# Web server:  http://localhost:8080/
```

## Architecture

### Dual Server Design
- **Admin Panel (port 8081)**: Uses Go standard library `net/http` - serves control panel UI and REST API
- **Web Server (port 8080)**: Raw TCP using `net.Listen` with manual HTTP parsing - serves static files

### Module Structure
```
├── main.go           # Entry point
├── config/           # Singleton config (sync.Once + RWMutex)
├── server/           # Web server core (TCP listener, graceful shutdown via context)
├── handler/          # HTTP parsing, file serving, response construction
├── logger/           # Async logging (channel producer-consumer, max 1000 entries)
├── router/           # Admin panel HTTP server (net/http mux)
├── static/           # Admin panel UI (HTML/CSS/JS)
└── www/              # Demo site (nested 3-level directory structure)
```

### Key Design Patterns

**Graceful Shutdown (server/server.go)**
Uses `context.Context` + `select` for race-free cancellation:
```go
func run() {
    for {
        conn, err := listener.Accept()
        if err != nil {
            select {
            case <-ctx.Done():
                return  // Clean exit
            default:
                continue
            }
        }
        // Handle connection...
    }
}
```

**Config Singleton (config/config.go)**
`sync.Once` for single initialization, `sync.RWMutex` for concurrent reads:
```go
var cfg *Config
once.Do(func() { cfg = &Config{Port: 8080, AdminPort: 8081, RootDir: "/root"} })
```

**Async Logger (logger/logger.go)**
Buffered channel (capacity 100) with background goroutine consumer. `select` with default to avoid blocking when channel is full.

### Security
- Path traversal prevention: checks `..` in URL and validates absolute path stays within rootDir
- Non-GET methods return 501 Not Implemented
- Connection read deadline: 30 seconds

### HTTP Implementation
- Manual parsing via `bufio.Reader.ReadString()`
- Content-Type mapping for 15+ file extensions
- Content-Disposition: `inline` for displayable types, `attachment` for binary
- Custom styled error pages (404, 501, 403, 500)
