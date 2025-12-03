package yellowstone

import (
	"context"

	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GeyserGrpcClient struct {
	Health grpc_health_v1.HealthClient
	Geyser pb.GeyserClient
	conn   *grpc.ClientConn
	ctx    context.Context
	cancel context.CancelFunc
}

func NewGeyserGrpcClient(
	health grpc_health_v1.HealthClient,
	geyser pb.GeyserClient,
	conn *grpc.ClientConn,
) *GeyserGrpcClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &GeyserGrpcClient{
		Health: health,
		Geyser: geyser,
		conn:   conn,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *GeyserGrpcClient) Start(stream pb.Geyser_SubscribeClient, fn func(*pb.SubscribeUpdate) error) error {
	defer c.cancel()
	for {
		select {
		case <-c.ctx.Done():
			return nil
		default:
			msg, err := stream.Recv()
			if err != nil {
				return err
			}

			if err := fn(msg); err != nil {
				return err
			}
		}
	}
}

func (c *GeyserGrpcClient) Close() error {
	c.cancel()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *GeyserGrpcClient) HealthCheck(ctx context.Context) (*grpc_health_v1.HealthCheckResponse, error) {
	request := &grpc_health_v1.HealthCheckRequest{
		Service: "geyser.Geyser",
	}
	response, err := c.Health.Check(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}

func (c *GeyserGrpcClient) HealthWatch(ctx context.Context) (grpc_health_v1.Health_WatchClient, error) {
	request := &grpc_health_v1.HealthCheckRequest{
		Service: "geyser.Geyser",
	}
	stream, err := c.Health.Watch(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return stream, nil
}

func (c *GeyserGrpcClient) Subscribe(ctx context.Context) (pb.Geyser_SubscribeClient, error) {
	return c.SubscribeWithRequest(ctx, nil)
}

func (c *GeyserGrpcClient) SubscribeWithRequest(
	ctx context.Context,
	request *pb.SubscribeRequest,
) (pb.Geyser_SubscribeClient, error) {
	stream, err := c.Geyser.Subscribe(ctx)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}

	if request != nil {
		if err := stream.Send(request); err != nil {
			return nil, NewSubscribeSendError(err)
		}
	}

	return stream, nil
}

func (c *GeyserGrpcClient) SubscribeOnce(
	ctx context.Context,
	request *pb.SubscribeRequest,
) (pb.Geyser_SubscribeClient, error) {
	return c.SubscribeWithRequest(ctx, request)
}

func (c *GeyserGrpcClient) Ping(ctx context.Context, count int32) (*pb.PongResponse, error) {
	message := &pb.PingRequest{Count: count}
	response, err := c.Geyser.Ping(ctx, message)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}

func (c *GeyserGrpcClient) GetLatestBlockhash(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetLatestBlockhashResponse, error) {
	request := &pb.GetLatestBlockhashRequest{}
	if commitment != nil {
		request.Commitment = commitment
	}
	response, err := c.Geyser.GetLatestBlockhash(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}

func (c *GeyserGrpcClient) GetBlockHeight(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetBlockHeightResponse, error) {
	request := &pb.GetBlockHeightRequest{}
	if commitment != nil {
		request.Commitment = commitment
	}
	response, err := c.Geyser.GetBlockHeight(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}

func (c *GeyserGrpcClient) GetSlot(
	ctx context.Context,
	commitment *pb.CommitmentLevel,
) (*pb.GetSlotResponse, error) {
	request := &pb.GetSlotRequest{}
	if commitment != nil {
		request.Commitment = commitment
	}
	response, err := c.Geyser.GetSlot(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}

func (c *GeyserGrpcClient) IsBlockhashValid(
	ctx context.Context,
	blockhash string,
	commitment *pb.CommitmentLevel,
) (*pb.IsBlockhashValidResponse, error) {
	request := &pb.IsBlockhashValidRequest{
		Blockhash: blockhash,
	}
	if commitment != nil {
		request.Commitment = commitment
	}
	response, err := c.Geyser.IsBlockhashValid(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}

func (c *GeyserGrpcClient) GetVersion(ctx context.Context) (*pb.GetVersionResponse, error) {
	request := &pb.GetVersionRequest{}
	response, err := c.Geyser.GetVersion(ctx, request)
	if err != nil {
		return nil, NewGrpcStatusError(err)
	}
	return response, nil
}
