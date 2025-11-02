# Word of Wisdom TCP Server

A TCP server protected from DDoS attacks using Proof of Work (PoW) challenge-response protocol. After successful PoW verification, the server responds with a random wisdom quote.

## Features

- **DDoS Protection**: SHA-256 Hashcash Proof of Work algorithm
- **Rate Limiting**: Sliding window rate limiter per IP address
- **Challenge-Response Protocol**: Custom binary protocol with message types
- **Quote Collection**: 20 wisdom quotes from philosophers and thinkers
- **Docker Support**: Containerized server and client applications
- **Graceful Shutdown**: Proper signal handling and connection cleanup

## Architecture

### Proof of Work Algorithm

This implementation uses **SHA-256 Hashcash**, chosen for:

1. **Asymmetric Cost**: Solving requires significant CPU work (O(2^difficulty)), but verification is instant (single hash)
2. **Adjustable Difficulty**: Can tune difficulty level to balance security vs. legitimate user experience
3. **Stateless**: No server-side state needed between challenge and verification
4. **Well-Tested**: Industry standard, used in Bitcoin and email spam prevention
5. **Simple to Implement**: Clear algorithm without complex cryptographic primitives

**Algorithm**: Find a nonce where `SHA256(difficulty || timestamp || random_data || client_ip || nonce)` has at least N leading zero bits.

### Protocol Flow

```
Client                          Server
  |                               |
  |----(1) ChallengeRequest------>|
  |                               |
  |<---(2) ChallengeResponse------|
  |      (difficulty, data, IP)   |
  |                               |
  |  (3) Solve PoW Challenge      |
  |                               |
  |----(4) Solution (nonce)------>|
  |                               |
  |                          (5) Verify
  |                               |
  |<---(6) Quote or Error---------|
```

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Build and start both server and client
docker-compose up --build

# The client will automatically connect and request 5 quotes
```

### Manual Docker Build

```bash
# Build server image
docker build -f Dockerfile.server -t pow-server .

# Build client image
docker build -f Dockerfile.client -t pow-client .

# Run server
docker run -p 8080:8080 pow-server

# Run client (in another terminal)
docker run --network host pow-client --server localhost:8080
```

### Local Development

```bash
# Build binaries
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client

# Run server
./bin/server

# Run client (in another terminal)
./bin/client
./bin/client --count=5  # Request 5 quotes
./bin/client --server=example.com:8080  # Connect to remote server
```

## Configuration

Edit `configs/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  max_connections: 1000
  read_timeout: 30s
  write_timeout: 10s
  connection_timeout: 120s

pow:
  base_difficulty: 20      # Number of leading zero bits required
  max_difficulty: 24
  challenge_max_age: 300s

rate_limit:
  max_requests: 10        # Max requests per window
  window: 60s             # Time window
  cleanup_interval: 60s

quotes:
  file_path: "./data/quotes.txt"
```

**Difficulty Guidelines**:
- `difficulty: 16` - Very easy, ~10ms solve time
- `difficulty: 20` - Easy, ~100-500ms solve time (current default)
- `difficulty: 24` - Medium, ~2-10s solve time
- `difficulty: 28` - Hard, ~30-120s solve time

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/pow
go test ./internal/protocol
go test ./internal/server

# Benchmark PoW performance
go test -bench=. ./internal/pow
```

## Project Structure

```
.
├── cmd/
│   ├── server/         # Server entry point
│   └── client/         # Client entry point
├── internal/
│   ├── protocol/       # Binary protocol implementation
│   ├── pow/            # Proof of Work (Hashcash)
│   ├── quotes/         # Quote manager
│   ├── server/         # TCP server and handlers
│   └── client/         # TCP client
├── configs/
│   └── config.yaml     # Server configuration
├── data/
│   └── quotes.txt      # Wisdom quotes collection
├── Dockerfile.server   # Server Docker image
├── Dockerfile.client   # Client Docker image
└── docker-compose.yml  # Orchestration
```

## Protocol Specification

### Message Format

All messages follow this structure:
```
[1 byte: MessageType][4 bytes: PayloadLength][N bytes: Payload]
```

### Message Types

- `0x01` - ChallengeRequest (client → server)
- `0x02` - ChallengeResponse (server → client)
- `0x03` - Solution (client → server)
- `0x04` - Quote (server → client)
- `0x05` - Error (server → client)

### Payload Formats

**ChallengeResponse**: `[1: difficulty][8: timestamp][32: random_data][N: client_ip]`
**Solution**: `[8: nonce]`
**Quote**: `[N: utf8_text]`
**Error**: `[2: error_code][N: error_message]`

## Security Considerations

1. **IP Binding**: PoW solutions include client IP, preventing solution reuse from different IPs
2. **Timestamp Validation**: Challenges expire after `challenge_max_age` to prevent replay attacks
3. **Rate Limiting**: Per-IP sliding window prevents excessive connection attempts
4. **Connection Limits**: Maximum concurrent connections to prevent resource exhaustion
5. **Payload Size Limits**: 1MB maximum payload size to prevent memory attacks

## Performance

With `difficulty: 20`:
- Solve time: 50-500ms (average ~250ms)
- Verification time: <1ms
- Server throughput: 1000+ verifications/second (single core)
- Memory per connection: <10KB

## License

MIT

## Requirements

- Go 1.25+
- Docker (optional, for containerized deployment)