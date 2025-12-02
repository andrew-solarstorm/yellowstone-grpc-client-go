package yellowstone

import (
	"context"
	"crypto/x509"
	"errors"
	"net/url"
	"time"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
)

type GeyserGrpcBuilder struct {
	endpoint                string
	xToken                  string
	xRequestSnapshot        bool
	sendCompressed          bool
	acceptCompressed        bool
	maxDecodingMessageSize  int
	maxEncodingMessageSize  int
	connectTimeout          time.Duration
	keepAliveInterval       time.Duration
	keepAliveTimeout        time.Duration
	keepAliveWithoutStream  bool
	tcpKeepalive            *time.Duration
	tcpNodelay              bool
	timeout                 time.Duration
	http2AdaptiveWindow     bool
	initialConnWindowSize   int
	initialStreamWindowSize int
	useTLS                  bool
	tlsConfig               *credentials.TransportCredentials
}

func BuildFromShared(endpoint string) (*GeyserGrpcBuilder, error) {
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, NewInvalidUriError(endpoint)
	}

	return newGeyserGrpcBuilder(endpoint), nil
}

func BuildFromStatic(endpoint string) *GeyserGrpcBuilder {
	return newGeyserGrpcBuilder(endpoint)
}

func newGeyserGrpcBuilder(endpoint string) *GeyserGrpcBuilder {
	return &GeyserGrpcBuilder{
		endpoint:               endpoint,
		tcpNodelay:             true,
		keepAliveWithoutStream: true,
	}
}

func (b *GeyserGrpcBuilder) XToken(token string) *GeyserGrpcBuilder {
	b.xToken = token
	return b
}

func (b *GeyserGrpcBuilder) SetXRequestSnapshot(value bool) *GeyserGrpcBuilder {
	b.xRequestSnapshot = value
	return b
}

func (b *GeyserGrpcBuilder) ConnectTimeout(dur time.Duration) *GeyserGrpcBuilder {
	b.connectTimeout = dur
	return b
}

func (b *GeyserGrpcBuilder) HTTP2AdaptiveWindow(enabled bool) *GeyserGrpcBuilder {
	b.http2AdaptiveWindow = enabled
	return b
}

func (b *GeyserGrpcBuilder) HTTP2KeepAliveInterval(interval time.Duration) *GeyserGrpcBuilder {
	b.keepAliveInterval = interval
	return b
}

func (b *GeyserGrpcBuilder) InitialConnectionWindowSize(size int) *GeyserGrpcBuilder {
	b.initialConnWindowSize = size
	return b
}

func (b *GeyserGrpcBuilder) InitialStreamWindowSize(size int) *GeyserGrpcBuilder {
	b.initialStreamWindowSize = size
	return b
}

func (b *GeyserGrpcBuilder) KeepAliveTimeout(duration time.Duration) *GeyserGrpcBuilder {
	b.keepAliveTimeout = duration
	return b
}

func (b *GeyserGrpcBuilder) KeepAliveWhileIdle(enabled bool) *GeyserGrpcBuilder {
	b.keepAliveWithoutStream = enabled
	return b
}

func (b *GeyserGrpcBuilder) TCPKeepalive(tcpKeepalive *time.Duration) *GeyserGrpcBuilder {
	b.tcpKeepalive = tcpKeepalive
	return b
}

func (b *GeyserGrpcBuilder) TCPNodelay(enabled bool) *GeyserGrpcBuilder {
	b.tcpNodelay = enabled
	return b
}

func (b *GeyserGrpcBuilder) Timeout(dur time.Duration) *GeyserGrpcBuilder {
	b.timeout = dur
	return b
}

func (b *GeyserGrpcBuilder) TLSConfig(config credentials.TransportCredentials) *GeyserGrpcBuilder {
	b.tlsConfig = &config
	b.useTLS = true
	return b
}

func (b *GeyserGrpcBuilder) SendCompressed(enable bool) *GeyserGrpcBuilder {
	b.sendCompressed = enable
	return b
}

func (b *GeyserGrpcBuilder) AcceptCompressed(enable bool) *GeyserGrpcBuilder {
	b.acceptCompressed = enable
	return b
}

func (b *GeyserGrpcBuilder) MaxDecodingMessageSize(limit int) *GeyserGrpcBuilder {
	b.maxDecodingMessageSize = limit
	return b
}

func (b *GeyserGrpcBuilder) MaxEncodingMessageSize(limit int) *GeyserGrpcBuilder {
	b.maxEncodingMessageSize = limit
	return b
}

func (b *GeyserGrpcBuilder) build(conn *grpc.ClientConn) *GeyserGrpcClient {
	geyser := pb.NewGeyserClient(conn)
	health := grpc_health_v1.NewHealthClient(conn)

	return NewGeyserGrpcClient(health, geyser, conn)
}

func (b *GeyserGrpcBuilder) Connect(ctx context.Context) (*GeyserGrpcClient, error) {
	conn, err := b.dial(ctx)
	if err != nil {
		return nil, NewTransportError(err)
	}
	return b.build(conn), nil
}

func (b *GeyserGrpcBuilder) ConnectLazy() (*GeyserGrpcClient, error) {
	conn, err := b.dial(context.Background())
	if err != nil {
		return nil, NewTransportError(err)
	}
	return b.build(conn), nil
}

func (b *GeyserGrpcBuilder) dial(ctx context.Context) (*grpc.ClientConn, error) {
	u, err := url.Parse(b.endpoint)
	if err != nil {
		return nil, err
	}

	httpMode := u.Scheme == "http"

	port := u.Port()
	if port == "" {
		if httpMode {
			port = "80"
		} else {
			port = "443"
		}
	}

	hostname := u.Hostname()
	if hostname == "" {
		return nil, errors.New("provide URL format endpoint e.g. http(s)://<endpoint>:<port>")
	}

	address := hostname + ":" + port

	var opts []grpc.DialOption

	if b.tlsConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(*b.tlsConfig))
	} else if httpMode {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, _ := x509.SystemCertPool()
		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	interceptor := &InterceptorXToken{
		XToken:           b.xToken,
		XRequestSnapshot: b.xRequestSnapshot,
	}
	opts = append(opts, grpc.WithUnaryInterceptor(interceptor.UnaryInterceptor))
	opts = append(opts, grpc.WithStreamInterceptor(interceptor.StreamInterceptor))

	if b.keepAliveInterval > 0 || b.keepAliveTimeout > 0 {
		keepAliveParams := keepalive.ClientParameters{
			PermitWithoutStream: b.keepAliveWithoutStream,
		}
		if b.keepAliveInterval > 0 {
			keepAliveParams.Time = b.keepAliveInterval
		}
		if b.keepAliveTimeout > 0 {
			keepAliveParams.Timeout = b.keepAliveTimeout
		}
		opts = append(opts, grpc.WithKeepaliveParams(keepAliveParams))
	}

	if b.maxDecodingMessageSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(b.maxDecodingMessageSize)))
	}

	if b.maxEncodingMessageSize > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(b.maxEncodingMessageSize)))
	}

	if b.initialConnWindowSize > 0 {
		opts = append(opts, grpc.WithInitialConnWindowSize(int32(b.initialConnWindowSize)))
	}

	if b.initialStreamWindowSize > 0 {
		opts = append(opts, grpc.WithInitialWindowSize(int32(b.initialStreamWindowSize)))
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
