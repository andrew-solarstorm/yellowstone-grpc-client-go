package methods

import (
	"context"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"github.com/gagliardetto/solana-go"
)

func SubscribeAccounts(endpoint string, token string) {
	builder, err := yellowstone.BuildFromShared(endpoint)
	if err != nil {
		log.Fatalf("Error building client: %v", err)
	}

	clientBuilder := builder.XToken(token).KeepAliveWhileIdle(true)

	if tlsConfig := getTLSConfig(endpoint); tlsConfig != nil {
		clientBuilder = clientBuilder.TLSConfig(tlsConfig)
	}

	grpcClient, err := clientBuilder.Connect(context.Background())
	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer grpcClient.Close()

	req := &pb.SubscribeRequest{
		Accounts: map[string]*pb.SubscribeRequestFilterAccounts{
			"account_filter": {
				Account: []string{},
			},
		},
		Slots: map[string]*pb.SubscribeRequestFilterSlots{
			"slot": {},
		},
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("👤 Listening for account updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Account:
			accountUpdate := update.GetAccount()
			account := accountUpdate.Account

			fmt.Printf("\n👤 Account Update:\n")
			fmt.Printf("   Pubkey: %s\n", solana.PublicKeyFromBytes(account.Pubkey).String())
			fmt.Printf("   Owner: %s\n", solana.PublicKeyFromBytes(account.Owner).String())
			fmt.Printf("   Lamports: %d\n", account.Lamports)
			fmt.Printf("   Executable: %v\n", account.Executable)
			fmt.Printf("   Rent Epoch: %d\n", account.RentEpoch)
			fmt.Printf("   Write Version: %d\n", account.WriteVersion)
			fmt.Printf("   Data Length: %d bytes\n", len(account.Data))
			fmt.Printf("   Slot: %d\n", accountUpdate.Slot)
			fmt.Printf("   Is Startup: %v\n", accountUpdate.IsStartup)

		case *pb.SubscribeUpdate_Slot:
			slot := update.GetSlot()
			fmt.Printf("📦 Slot: %d (Status: %s)\n", slot.Slot, slot.Status)

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
	fmt.Println("✅ Account subscription example completed")
}
