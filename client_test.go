package yellowstone

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/metadata"
)

func TestChannelHTTPSSuccess(t *testing.T) {
	endpoint := "https://ams17.rpcpool.com:443"
	xToken := "1000000000000000000000000007"

	builder, err := BuildFromShared(endpoint)
	if err != nil {
		t.Fatalf("BuildFromShared failed: %v", err)
	}

	builder = builder.XToken(xToken)

	client, err := builder.ConnectLazy()
	if err != nil {
		t.Fatalf("ConnectLazy failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

func TestChannelHTTPSuccess(t *testing.T) {
	endpoint := "http://127.0.0.1:10000"
	xToken := "1234567891012141618202224268"

	builder, err := BuildFromShared(endpoint)
	if err != nil {
		t.Fatalf("BuildFromShared failed: %v", err)
	}

	builder = builder.XToken(xToken)

	client, err := builder.ConnectLazy()
	if err != nil {
		t.Fatalf("ConnectLazy failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

func TestChannelEmptyTokenSome(t *testing.T) {
	endpoint := "http://127.0.0.1:10000"
	xToken := ""

	builder, err := BuildFromShared(endpoint)
	if err != nil {
		t.Fatalf("BuildFromShared failed: %v", err)
	}

	builder = builder.XToken(xToken)

	client, err := builder.ConnectLazy()
	if err != nil {
		t.Fatalf("ConnectLazy failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

func TestChannelNoToken(t *testing.T) {
	endpoint := "http://127.0.0.1:10000"

	builder, err := BuildFromShared(endpoint)
	if err != nil {
		t.Fatalf("BuildFromShared failed: %v", err)
	}

	client, err := builder.ConnectLazy()
	if err != nil {
		t.Fatalf("ConnectLazy failed: %v", err)
	}

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	defer client.Close()
}

func TestChannelInvalidURI(t *testing.T) {
	endpoint := "sites/files/images/picture.png"

	_, err := BuildFromShared(endpoint)
	if err == nil {
		t.Fatal("Expected error for invalid URI, got nil")
	}

	builderErr, ok := err.(*GeyserGrpcBuilderError)
	if !ok {
		t.Fatalf("Expected GeyserGrpcBuilderError, got %T", err)
	}

	if builderErr.Type != "InvalidUri" {
		t.Fatalf("Expected InvalidUri error type, got %s", builderErr.Type)
	}
}

func TestBuilderOptions(t *testing.T) {
	endpoint := "http://127.0.0.1:10000"

	builder, err := BuildFromShared(endpoint)
	if err != nil {
		t.Fatalf("BuildFromShared failed: %v", err)
	}

	builder = builder.
		XToken("test-token").
		SetXRequestSnapshot(true).
		ConnectTimeout(5 * time.Second).
		HTTP2KeepAliveInterval(60 * time.Second).
		KeepAliveTimeout(10 * time.Second).
		KeepAliveWhileIdle(true).
		TCPNodelay(true).
		SendCompressed(true).
		AcceptCompressed(true).
		MaxDecodingMessageSize(4 * 1024 * 1024).
		MaxEncodingMessageSize(4 * 1024 * 1024)

	if builder.xToken != "test-token" {
		t.Errorf("Expected xToken to be 'test-token', got %s", builder.xToken)
	}

	if !builder.xRequestSnapshot {
		t.Error("Expected xRequestSnapshot to be true")
	}

	if builder.connectTimeout != 5*time.Second {
		t.Errorf("Expected connectTimeout to be 5s, got %v", builder.connectTimeout)
	}

	if !builder.sendCompressed {
		t.Error("Expected sendCompressed to be true")
	}

	if !builder.acceptCompressed {
		t.Error("Expected acceptCompressed to be true")
	}
}

func TestBuildFromStatic(t *testing.T) {
	endpoint := "http://127.0.0.1:10000"

	builder := BuildFromStatic(endpoint)
	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}

	if builder.endpoint != endpoint {
		t.Errorf("Expected endpoint to be %s, got %s", endpoint, builder.endpoint)
	}
}

func TestInterceptorXToken(t *testing.T) {
	interceptor := &InterceptorXToken{
		XToken:           "test-token",
		XRequestSnapshot: true,
	}

	ctx := context.Background()
	ctx = interceptor.addMetadata(ctx)

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("Expected metadata in context")
	}

	tokens := md.Get("x-token")
	if len(tokens) != 1 || tokens[0] != "test-token" {
		t.Errorf("Expected x-token to be 'test-token', got %v", tokens)
	}

	snapshots := md.Get("x-request-snapshot")
	if len(snapshots) != 1 || snapshots[0] != "true" {
		t.Errorf("Expected x-request-snapshot to be 'true', got %v", snapshots)
	}
}

func TestInterceptorXTokenNoToken(t *testing.T) {
	interceptor := &InterceptorXToken{
		XToken:           "",
		XRequestSnapshot: false,
	}

	ctx := context.Background()
	ctx = interceptor.addMetadata(ctx)

	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		t.Fatalf("Expected no metadata in context, got %v", md)
	}
}
