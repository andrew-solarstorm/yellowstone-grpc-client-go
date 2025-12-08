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

func SubscribeAccountsByOwner(endpoint string, token string) {
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
			"token_program_accounts": {
				Owner: []string{
					"LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo",
				},
			},
		},
		Slots: map[string]*pb.SubscribeRequestFilterSlots{"slot": {}},
	}

	stream, err := grpcClient.SubscribeWithRequest(context.Background(), req)
	if err != nil {
		log.Fatalf("Error subscribing: %v", err)
	}

	fmt.Println("👥 Listening for account updates by owner...")
	fmt.Println("📡 Monitoring: Token Program owned accounts")
	fmt.Println()

	accountCount := 0

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Account:
			accountUpdate := update.GetAccount()
			account := accountUpdate.Account

			if !accountUpdate.IsStartup {
				accountCount++
				fmt.Printf("\n👥 Account Update #%d:\n", accountCount)
				fmt.Printf("   Pubkey: %s\n", solana.PublicKeyFromBytes(account.Pubkey).String())
				fmt.Printf("   Owner: %s\n", solana.PublicKeyFromBytes(account.Owner).String())
				fmt.Printf("   Lamports: %d\n", account.Lamports)
				fmt.Printf("   Data Length: %d bytes\n", len(account.Data))
				fmt.Printf("   Slot: %d\n", accountUpdate.Slot)
				fmt.Printf("   Write Version: %d\n", account.WriteVersion)
			}

		case *pb.SubscribeUpdate_Slot:
			slot := update.GetSlot()
			fmt.Printf("📦 Slot: %d (%s)\n", slot.Slot, slot.Status)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting client: %v", err)
	}

	time.Sleep(30 * time.Second)
	fmt.Printf("\n📊 Total account updates received: %d\n", accountCount)
	fmt.Println("✅ Completed")
}
