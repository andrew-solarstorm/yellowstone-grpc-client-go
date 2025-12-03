package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

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

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Slot:
			fmt.Printf("📦 Slot: %d\n", update.GetSlot().Slot)

		case *pb.SubscribeUpdate_Account:
			fmt.Printf("🔹 Account update: %s\n", update.GetAccount().Account.Pubkey)

		case *pb.SubscribeUpdate_Transaction:
			fmt.Printf("🔹 Transaction update: %s\n", update.GetTransaction().Transaction.Signature)

		case *pb.SubscribeUpdate_Block:
			fmt.Printf("🔹 Block update: slot=%d\n", update.GetBlock().Slot)

		case *pb.SubscribeUpdate_Ping:
			return nil

		case *pb.SubscribeUpdate_Pong:
			return nil

		case *pb.SubscribeUpdate_BlockMeta:
			fmt.Printf("🔹 BlockMeta update: slot=%d\n", update.GetBlockMeta().Slot)

		case *pb.SubscribeUpdate_Entry:
			fmt.Printf("🔹 Entry update: slot=%d\n", update.GetEntry().Slot)

		case nil:
			fmt.Println("⚠️  Empty update")

		default:
			fmt.Printf("🔹 Other: %T\n", update)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting client: %v", err)
	}

	fmt.Println("Listening for updates...")
	time.Sleep(10 * time.Second)
	grpcClient.Close()
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
