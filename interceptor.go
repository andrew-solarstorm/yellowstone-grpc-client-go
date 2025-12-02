package yellowstone

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type InterceptorXToken struct {
	XToken           string
	XRequestSnapshot bool
}

func (i *InterceptorXToken) UnaryInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	ctx = i.addMetadata(ctx)
	return invoker(ctx, method, req, reply, cc, opts...)
}

func (i *InterceptorXToken) StreamInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	ctx = i.addMetadata(ctx)
	return streamer(ctx, desc, cc, method, opts...)
}

func (i *InterceptorXToken) addMetadata(ctx context.Context) context.Context {
	md := metadata.MD{}

	if i.XToken != "" {
		md.Set("x-token", i.XToken)
	}

	if i.XRequestSnapshot {
		md.Set("x-request-snapshot", "true")
	}

	if len(md) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return ctx
}
