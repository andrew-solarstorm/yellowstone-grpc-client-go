package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	"github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"github.com/joho/godotenv"
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

	fmt.Println("Listening for updates...")

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Printf("⚡ Stream error: %v", err)
			break
		}

		switch update := msg.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Slot:
			fmt.Printf("📦 Slot: %d\n", update.Slot.Slot)

		case *pb.SubscribeUpdate_Account:
			fmt.Printf("🔹 Account update: %s\n", update.Account.Account.Pubkey)

		case *pb.SubscribeUpdate_Transaction:
			fmt.Printf("🔹 Transaction update\n")

		case *pb.SubscribeUpdate_Block:
			fmt.Printf("🔹 Block update: slot=%d\n", update.Block.Slot)

		case *pb.SubscribeUpdate_Ping:
			continue

		case *pb.SubscribeUpdate_Pong:
			continue

		case *pb.SubscribeUpdate_BlockMeta:
			fmt.Printf("🔹 BlockMeta update: slot=%d\n", update.BlockMeta.Slot)

		case *pb.SubscribeUpdate_Entry:
			fmt.Printf("🔹 Entry update: slot=%d\n", update.Entry.Slot)

		case nil:
			fmt.Println("⚠️  Empty update")

		default:
			fmt.Printf("🔹 Other: %T\n", update)
		}
	}
}

func main() {

	godotenv.Load()

	token := os.Getenv("TOKEN")
	endpoint := os.Getenv("ENDPOINT")
	if token == "" || endpoint == "" {
		log.Fatalf("TOKEN and ENDPOINT must be set")
	}

	SubscribeSlot(endpoint, token)
}
