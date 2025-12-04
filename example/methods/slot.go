package methods

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"google.golang.org/grpc/credentials"
)

func SubscribeSlot(endpoint string, token string) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	pool, err := x509.SystemCertPool()
	if err != nil {
		log.Fatalf("Error loading system cert pool: %v", err)
	}
	tlsConfig := credentials.NewClientTLSFromCert(pool, "")

	grpcClient, err := builder.
		XToken(token).
		TLSConfig(tlsConfig).
		KeepAliveWhileIdle(true).
		Connect(context.Background())

	if err != nil {
		log.Fatalf("Error connecting to geyser: %v", err)
	}
	defer grpcClient.Close()

	req := &pb.SubscribeRequest{
		Slots: map[string]*pb.SubscribeRequestFilterSlots{
			"slot": {},
		},
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("📦 Listening for slot updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Slot:
			slot := update.GetSlot()
			fmt.Printf("📦 Slot Update:\n")
			fmt.Printf("   Slot: %d\n", slot.Slot)
			fmt.Printf("   Parent: %d\n", slot.Parent)
			fmt.Printf("   Status: %s\n", slot.Status)

		case *pb.SubscribeUpdate_Ping:
			return nil

		case *pb.SubscribeUpdate_Pong:
			return nil

		default:
			fmt.Printf("⚠️  Unexpected update type: %T\n", update.GetUpdateOneof())
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting client: %v", err)
	}

	time.Sleep(30 * time.Second)
	grpcClient.Close()
	fmt.Println("✅ Slot subscription example completed")
}
