package methods

import (
	"context"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	"github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
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

	req := &proto.SubscribeRequest{
		Accounts: map[string]*proto.SubscribeRequestFilterAccounts{
			"sliced": {
				Owner: []string{"TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
				Filters: []*proto.SubscribeRequestFilterAccountsFilter{
					{
						Filter: &proto.SubscribeRequestFilterAccountsFilter_Datasize{
							Datasize: 165,
						},
					},
				},
			},
		},
		AccountsDataSlice: []*proto.SubscribeRequestAccountsDataSlice{
			{Offset: 0, Length: 32},  // First 32 bytes (mint)
			{Offset: 32, Length: 32}, // Next 32 bytes (owner)
		},
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
			if len(account.Data) == 64 {
				fmt.Printf("   Mint: %s\n", solana.PublicKeyFromBytes(account.Data[:32]))
				fmt.Printf("   Owner: %s\n", solana.PublicKeyFromBytes(account.Data[32:]))
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
