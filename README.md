## PriceFetcher: gRPC microservice for stock prices

This service exposes:
- JSON HTTP API on port 3000
- gRPC API on port 4000 (unary, server streaming, bidirectional streaming)

It fetches real-time/near real-time U.S. stock quote data from Twelve Data via RapidAPI and returns a normalized response:

```
{
  "symbol": "GOOGL",
  "name": "Alphabet Inc.",
  "datetime": "2025-08-25",
  "close": "208.49001",
  "percent_change": "1.16454"
}
```

### Prerequisites
- Go 1.21+
- Make
- Protobuf compiler (`protoc`)
- gRPC tooling for Go (codegen plugins)

### Environment configuration
Create a `.env` file in the project root with the following variables. These are required for the app to run:

```
XRapidAPIKey=<your-rapidapi-key>
API_URL=https://twelve-data1.p.rapidapi.com/quote?
XRapidAPIHost2=twelve-data1.p.rapidapi.com
```

Notes:
- `API_URL` must include the trailing `?` because the service appends the query string parameters.
- The service internally adds: `symbol=<TICKER>&interval=1day&outputsize=30&format=json`.

---

## Setup

### Install Protocol Buffers (protoc)
Linux:
```bash
sudo apt update && sudo apt install -y protobuf-compiler
```

macOS:
```bash
brew install protobuf
```

### Install Go protobuf/gRPC plugins
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure `GOBIN` is on your PATH (typically `${HOME}/go/bin`):
```bash
export PATH="${PATH}:${HOME}/go/bin"
```

### Generate protobuf stubs
This project provides a Make target to regenerate Go stubs from `proto/service.proto` into the `pb/` folder.
```bash
make proto
```

---

## Build and run
Build the binary:
```bash
make build
```

Run the service (builds if necessary):
```bash
make run
```

This will start:
- JSON server on `:3000`
- gRPC server on `:4000`

Docker (optional):
```bash
docker build -t pricefetcher:latest .
docker run --env-file .env -p 3000:3000 -p 4000:4000 pricefetcher:latest
```

---

## Using the APIs

### JSON HTTP API (port 3000)
Endpoint: `GET /?symbol=<TICKER>`

Examples:
```bash
curl "http://localhost:3000/?symbol=GOOGL"
```

Successful response (example):
```json
{
  "close": "208.49001",
  "datetime": "2025-08-25",
  "name": "Alphabet Inc.",
  "percent_change": "1.16454",
  "symbol": "GOOGL"
}
```

Error response:
```json
{"error":"invalid ticker"}
```

### gRPC API (port 4000)
Service definition: see `proto/service.proto`.

RPCs:
- `FetchPrice(PriceRequest) returns (PriceResponse)` — unary
- `FetchPriceServerStreaming(SymbolsList) returns (stream PriceResponse)` — server streaming
- `FetchPriceBidirectionalStreaming(stream PriceRequest) returns (stream PriceResponse)` — bidirectional streaming

#### With grpcurl
Install grpcurl:
```bash
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

Package managers:
```bash
# Ubuntu/Debian (via apt repository)
sudo apt update && sudo apt install -y grpcurl || echo "If not found, use the Go install above"

# macOS (Homebrew)
brew install grpcurl
```

Unary:
```bash
grpcurl -plaintext -d '{"symbol":"GOOGL"}' localhost:4000 PriceFetcher/FetchPrice
```

Server streaming:
```bash
grpcurl -plaintext -d '{"symbols":[{"symbol":"AAPL"},{"symbol":"GOOGL"},{"symbol":"MSFT"}]}' \
  localhost:4000 PriceFetcher/FetchPriceServerStreaming
```

Bidirectional streaming (interactive stdin streaming):
```bash
grpcurl -plaintext -d @ localhost:4000 PriceFetcher/FetchPriceBidirectionalStreaming <<'EOF'
{"symbol":"AAPL"}
{"symbol":"GOOGL"}
{"symbol":"MSFT"}
EOF
```

#### From Go (example client)
See `client/client.go` for examples of:
- Unary call: `CallFetchPrice`
- Server streaming: `CallFetchPriceServerStreaming`
- Bidirectional streaming: `CallFetchPriceBidirectionalStreaming`

You can wire these up in `main.go` or your own program to test the gRPC flows.

Minimal unary example:
```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    pb "github.com/bruce-mig/pricefetcher-grpc-microservices/pb"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    conn, err := grpc.NewClient(remoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil { log.Fatal(err) }
    defer conn.Close()

    client := pb.NewPriceFetcherClient(conn)
    resp, err := client.FetchPrice(ctx, &pb.PriceRequest{Symbol: "GOOGL"})
    if err != nil { log.Fatal(err) }

    fmt.Printf("%+v\n", resp)
}
```

---

## Development notes
- `make proto` uses `protoc` with the Go and gRPC plugins to generate files into `pb/`.
- `make run` loads `.env` via `github.com/joho/godotenv`.
- The service logs request IDs and latency via Logrus.

## Troubleshooting
- 404/connection refused: ensure the service is running and ports 3000/4000 are free.
- 400 with `{ "error": "invalid ticker" }`: ensure the `symbol` query param is present.
- Empty or unexpected data: verify `.env` values, especially `XRapidAPIKey`, `API_URL`, `XRapidAPIHost2`.
- Proto errors: verify `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` are installed and on PATH.

---

## License
MIT