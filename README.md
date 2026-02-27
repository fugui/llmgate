# LLM Gate

A Go-based API Gateway for routing LLM (Large Language Model) API calls to different providers including OpenAI, Claude, and Azure OpenAI.

## Features

- **Multi-Provider Routing**: Distribute API calls across multiple LLM providers
- **Rate Limiting**: Configurable per-IP and global rate limiting
- **Request/Response Logging**: Structured logging with JSON or text format
- **Health Monitoring**: Health check endpoint for monitoring provider status
- **Automatic Failover**: Retry requests with alternate providers on failure
- **CORS Support**: Built-in CORS handling for web applications
- **Graceful Shutdown**: Clean shutdown with request draining

## Quick Start

### Prerequisites

- Go 1.21 or later
- API keys for desired LLM providers

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd llmgate

# Install dependencies
go mod download

# Set environment variables
export OPENAI_API_KEY="your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
# Optional: export AZURE_OPENAI_KEY="your-azure-key"

# Run the server
go run main.go -config config.yaml
```

### Configuration

Create a `config.yaml` file:

```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

providers:
  openai:
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"
    timeout: 60s
    priority: 1
  claude:
    base_url: "https://api.anthropic.com/v1"
    api_key: "${ANTHROPIC_API_KEY}"
    timeout: 60s
    priority: 2

rate_limit:
  enabled: true
  requests_per_second: 10
  burst_size: 20

logging:
  level: "info"
  format: "json"
  request_logging: true
```

## Usage

### Send Requests

Once running, send requests to the gateway:

```bash
# OpenAI-style chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Claude-style messages
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: your-anthropic-key" \
  -d '{
    "model": "claude-3-opus-20240229",
    "max_tokens": 1024,
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Health Check

```bash
curl http://localhost:8080/health
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `/v1/chat/completions` | OpenAI chat completions (routes to OpenAI/Azure/Claude) |
| `/v1/completions` | OpenAI completions (routes to OpenAI/Azure) |
| `/v1/messages` | Claude messages API (routes to Claude) |
| `/health` | Health check endpoint |

## Routing Logic

The gateway routes requests based on path matching:

1. `/v1/chat/completions` → OpenAI → Azure → Claude
2. `/v1/completions` → OpenAI → Azure
3. `/v1/messages` → Claude

If a provider returns a 5xx error, the gateway automatically tries the next provider in the chain.

## Configuration Options

### Server

| Option | Default | Description |
|--------|---------|-------------|
| `port` | 8080 | HTTP server port |
| `read_timeout` | 30s | Request read timeout |
| `write_timeout` | 30s | Response write timeout |
| `idle_timeout` | 120s | Keep-alive timeout |

### Rate Limiting

| Option | Default | Description |
|--------|---------|-------------|
| `enabled` | true | Enable rate limiting |
| `requests_per_second` | 10 | Rate limit per IP |
| `burst_size` | 20 | Burst capacity |
| `per_ip` | true | Apply limits per IP address |

### Logging

| Option | Default | Description |
|--------|---------|-------------|
| `level` | info | Log level (debug/info/warn/error) |
| `format` | json | Output format (json/text) |
| `output` | stdout | Output destination (stdout/file) |
| `request_logging` | true | Log HTTP requests |
| `response_logging` | true | Log response bodies (debug only) |

## Project Structure

```
llmgate/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── config.yaml                  # Default configuration
├── README.md                    # This file
└── internal/
    ├── config/                  # Configuration management
    │   └── config.go
    ├── gateway/                 # Request routing and proxying
    │   └── gateway.go
    ├── logging/                 # Request/response logging
    │   └── logger.go
    └── ratelimit/               # Rate limiting
        └── ratelimiter.go
```

## Development

### Run Tests

```bash
go test ./...
```

### Build Binary

```bash
go build -o llmgate .
```

### Docker

```bash
docker build -t llmgate .
docker run -p 8080:8080 -e OPENAI_API_KEY=$OPENAI_API_KEY llmgate
```

## License

MIT License

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.
