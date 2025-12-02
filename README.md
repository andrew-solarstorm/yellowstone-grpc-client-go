# yellowstone-grpc-client-go

A Go client library for [Yellowstone gRPC](https://github.com/rpcpool/yellowstone-grpc), providing real-time access to Solana blockchain data via Geyser plugin.

## Features

- 🚀 **Real-time Streaming** - Subscribe to slots, accounts, transactions, blocks, and more
- 🔧 **Builder Pattern** - Fluent API for configuring gRPC connections
- 🔐 **Authentication Support** - Built-in support for x-token authentication
- 🔒 **TLS/SSL Support** - Secure connections with configurable TLS
- ⚡ **Keep-Alive** - Configurable HTTP/2 keep-alive for stable connections
- 📦 **Type-Safe** - Full protobuf definitions with generated Go types
- 🏥 **Health Checks** - Built-in health check and watch capabilities

## Installation

```bash
go get github.com/andrew-solarstorm/yellowstone-grpc-client-go
```

## Quick Start

```go
package main

import (
    "context"
    "crypto/x509"
    "log"

    yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
    pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
    "google.golang.org/grpc/credentials"
)

func main() {
    // Create a builder
    builder, err := yellowstone.BuildFromShared("https://your-endpoint.com")
    if err != nil {
        log.Fatalf("Error building client: %v", err)
    }

    // Configure TLS
    pool, _ := x509.SystemCertPool()
    tlsConfig := credentials.NewClientTLSFromCert(pool, "")

    // Connect with authentication
    client, err := builder.
        XToken("your-auth-token").
        TLSConfig(tlsConfig).
        KeepAliveWhileIdle(true).
        Connect(context.Background())
    
    if err != nil {
        log.Fatalf("Error connecting: %v", err)
    }
    defer client.Close()

    // Subscribe to slot updates
    req := &pb.SubscribeRequest{
        Slots: map[string]*pb.SubscribeRequestFilterSlots{
            "slot": {},
        },
    }

    stream, err := client.SubscribeWithRequest(context.Background(), req)
    if err != nil {
        log.Fatalf("Error subscribing: %v", err)
    }

    // Process updates
    for {
        msg, err := stream.Recv()
        if err != nil {
            log.Printf("Stream error: %v", err)
            break
        }

        if update := msg.GetSlot(); update != nil {
            log.Printf("Slot: %d", update.Slot)
        }
    }
}
```

## Usage

### Building a Client

There are two ways to create a client builder:

#### 1. BuildFromShared (Recommended)
Validates the endpoint URL format:

```go
builder, err := yellowstone.BuildFromShared("https://your-endpoint.com")
if err != nil {
    log.Fatal(err)
}
```

#### 2. BuildFromStatic
Skips URL validation (use when you're certain the endpoint is valid):

```go
builder := yellowstone.BuildFromStatic("https://your-endpoint.com")
```

### Configuring the Connection

The builder provides a fluent API for configuration:

```go
client, err := builder.
    XToken("your-token").                          // Authentication token
    TLSConfig(tlsConfig).                          // Custom TLS configuration
    KeepAliveWhileIdle(true).                     // Keep connection alive
    HTTP2KeepAliveInterval(30 * time.Second).     // Keep-alive interval
    KeepAliveTimeout(10 * time.Second).           // Keep-alive timeout
    ConnectTimeout(5 * time.Second).              // Connection timeout
    MaxDecodingMessageSize(64 * 1024 * 1024).     // Max message size (64MB)
    TCPNodelay(true).                             // Disable Nagle's algorithm
    Connect(context.Background())
```

### Subscription Types

#### Slot Updates

```go
req := &pb.SubscribeRequest{
    Slots: map[string]*pb.SubscribeRequestFilterSlots{
        "slot": {},
    },
}
stream, err := client.SubscribeWithRequest(ctx, req)
```

#### Account Updates

```go
req := &pb.SubscribeRequest{
    Accounts: map[string]*pb.SubscribeRequestFilterAccounts{
        "account_filter": {
            Account: []string{"your-account-pubkey"},
        },
    },
}
stream, err := client.SubscribeWithRequest(ctx, req)
```

#### Transaction Updates

```go
req := &pb.SubscribeRequest{
    Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
        "tx_filter": {
            AccountInclude: []string{"your-account-pubkey"},
        },
    },
}
stream, err := client.SubscribeWithRequest(ctx, req)
```

#### Block Updates

```go
req := &pb.SubscribeRequest{
    Blocks: map[string]*pb.SubscribeRequestFilterBlocks{
        "block_filter": {},
    },
}
stream, err := client.SubscribeWithRequest(ctx, req)
```

### Processing Updates

```go
for {
    msg, err := stream.Recv()
    if err != nil {
        log.Printf("Stream error: %v", err)
        break
    }

    switch update := msg.GetUpdateOneof().(type) {
    case *pb.SubscribeUpdate_Slot:
        log.Printf("Slot: %d", update.Slot.Slot)
    
    case *pb.SubscribeUpdate_Account:
        log.Printf("Account: %s", update.Account.Account.Pubkey)
    
    case *pb.SubscribeUpdate_Transaction:
        log.Printf("Transaction signature: %v", update.Transaction.Transaction.Signature)
    
    case *pb.SubscribeUpdate_Block:
        log.Printf("Block slot: %d", update.Block.Slot)
    
    case *pb.SubscribeUpdate_Ping:
        // Handle ping
    
    case *pb.SubscribeUpdate_Pong:
        // Handle pong
    
    default:
        log.Printf("Unknown update type: %T", update)
    }
}
```

## API Reference

### GeyserGrpcClient Methods

#### Subscription
- `Subscribe(ctx) (stream, error)` - Create a subscription stream
- `SubscribeWithRequest(ctx, *SubscribeRequest) (stream, error)` - Subscribe with initial request
- `SubscribeOnce(ctx, *SubscribeRequest) (stream, error)` - Alias for SubscribeWithRequest

#### Health
- `HealthCheck(ctx) (*HealthCheckResponse, error)` - Check service health
- `HealthWatch(ctx) (HealthWatchClient, error)` - Watch health status changes

#### Blockchain Queries
- `GetLatestBlockhash(ctx, *CommitmentLevel) (*GetLatestBlockhashResponse, error)`
- `GetBlockHeight(ctx, *CommitmentLevel) (*GetBlockHeightResponse, error)`
- `GetSlot(ctx, *CommitmentLevel) (*GetSlotResponse, error)`
- `IsBlockhashValid(ctx, blockhash, *CommitmentLevel) (*IsBlockhashValidResponse, error)`
- `GetVersion(ctx) (*GetVersionResponse, error)`

#### Utility
- `Ping(ctx, count) (*PongResponse, error)` - Send ping request
- `Close() error` - Close the connection

### Builder Configuration Methods

| Method | Description |
|--------|-------------|
| `XToken(string)` | Set authentication token |
| `SetXRequestSnapshot(bool)` | Enable/disable snapshot requests |
| `TLSConfig(TransportCredentials)` | Set custom TLS configuration |
| `KeepAliveWhileIdle(bool)` | Keep connection alive when idle |
| `HTTP2KeepAliveInterval(duration)` | Set keep-alive interval |
| `KeepAliveTimeout(duration)` | Set keep-alive timeout |
| `ConnectTimeout(duration)` | Set connection timeout |
| `TCPKeepalive(*duration)` | Set TCP keep-alive duration |
| `TCPNodelay(bool)` | Enable/disable TCP Nodelay (Nagle's algorithm) |
| `HTTP2AdaptiveWindow(bool)` | Enable HTTP/2 adaptive window |
| `InitialConnectionWindowSize(int)` | Set initial connection window size |
| `InitialStreamWindowSize(int)` | Set initial stream window size |
| `MaxDecodingMessageSize(int)` | Set max message receive size |
| `MaxEncodingMessageSize(int)` | Set max message send size |
| `SendCompressed(bool)` | Enable compression for sent messages |
| `AcceptCompressed(bool)` | Accept compressed messages |

## Environment Variables

The example uses the following environment variables:

```bash
TOKEN=your-yellowstone-token
ENDPOINT=https://your-geyser-endpoint.com
```

Create a `.env` file in your project:

```env
TOKEN=your-actual-token
ENDPOINT=https://your-actual-endpoint.com
```

## Running the Example

```bash
cd example
cp .env.example .env  # Edit with your credentials
go run main.go
```

## Error Handling

The library provides custom error types for better error handling:

```go
client, err := builder.Connect(ctx)
if err != nil {
    // Handle connection errors
    log.Fatalf("Connection error: %v", err)
}

stream, err := client.SubscribeWithRequest(ctx, req)
if err != nil {
    // Handle subscription errors
    log.Fatalf("Subscription error: %v", err)
}
```

## Advanced Configuration

### Custom TLS Configuration

```go
import (
    "crypto/tls"
    "crypto/x509"
    "google.golang.org/grpc/credentials"
)

// Load custom CA certificates
caCert, err := os.ReadFile("ca-cert.pem")
if err != nil {
    log.Fatal(err)
}

certPool := x509.NewCertPool()
certPool.AppendCertsFromPEM(caCert)

tlsConfig := credentials.NewTLS(&tls.Config{
    RootCAs: certPool,
})

client, err := builder.TLSConfig(tlsConfig).Connect(ctx)
```

### Connection Pooling and Performance

```go
client, err := builder.
    HTTP2AdaptiveWindow(true).                    // Enable adaptive flow control
    InitialConnectionWindowSize(1 << 20).         // 1MB
    InitialStreamWindowSize(1 << 20).             // 1MB
    MaxDecodingMessageSize(64 * 1024 * 1024).    // 64MB
    HTTP2KeepAliveInterval(30 * time.Second).
    KeepAliveTimeout(10 * time.Second).
    Connect(ctx)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here]

## Acknowledgments

- [Yellowstone gRPC](https://github.com/rpcpool/yellowstone-grpc) - The Geyser gRPC plugin
- [Solana](https://solana.com) - The blockchain this client interfaces with

## Support

For issues and questions:
- Open an issue on GitHub
- Check the [example](./example) directory for usage examples

## Related Projects

- [yellowstone-grpc](https://github.com/rpcpool/yellowstone-grpc) - The gRPC plugin for Solana
- [Solana Go SDK](https://github.com/gagliardetto/solana-go) - Complete Solana SDK for Go
