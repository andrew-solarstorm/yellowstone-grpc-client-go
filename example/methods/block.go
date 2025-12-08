package methods

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

func SubscribeBlocks(endpoint string, token string) {
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
		Blocks: map[string]*pb.SubscribeRequestFilterBlocks{
			"block_filter": {},
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

	fmt.Println("🧱 Listening for block updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Block:
			blockUpdate := update.GetBlock()

			fmt.Printf("\n🧱 Block Update:\n")
			fmt.Printf("   Slot: %d\n", blockUpdate.Slot)
			fmt.Printf("   Blockhash: %s\n", blockUpdate.Blockhash)
			fmt.Printf("   Parent Slot: %d\n", blockUpdate.ParentSlot)
			fmt.Printf("   Parent Blockhash: %s\n", blockUpdate.ParentBlockhash)
			if blockUpdate.BlockHeight != nil {
				fmt.Printf("   Block Height: %d\n", blockUpdate.BlockHeight.BlockHeight)
			}
			if blockUpdate.BlockTime != nil {
				fmt.Printf("   Block Time: %d\n", blockUpdate.BlockTime.Timestamp)
			}
			fmt.Printf("   Transactions Count: %d\n", len(blockUpdate.Transactions))
			if blockUpdate.Rewards != nil && blockUpdate.Rewards.Rewards != nil {
				fmt.Printf("   Rewards Count: %d\n", len(blockUpdate.Rewards.Rewards))
			}

			if len(blockUpdate.Transactions) > 0 {
				fmt.Printf("\n   First 3 transactions:\n")
				for i, tx := range blockUpdate.Transactions {
					if i >= 3 {
						break
					}
					if tx.Transaction != nil && len(tx.Transaction.Signatures) > 0 {
						sig := base64.StdEncoding.EncodeToString(tx.Transaction.Signatures[0])
						fmt.Printf("      %d. %s\n", i+1, sig[:32]+"...")
					}
				}
			}

			if blockUpdate.Rewards != nil && blockUpdate.Rewards.Rewards != nil && len(blockUpdate.Rewards.Rewards) > 0 {
				fmt.Printf("\n   Rewards:\n")
				for i, reward := range blockUpdate.Rewards.Rewards {
					if i >= 3 {
						fmt.Printf("      ... and %d more\n", len(blockUpdate.Rewards.Rewards)-3)
						break
					}
					fmt.Printf("      Pubkey: %s, Lamports: %d, Type: %s\n",
						reward.Pubkey[:16]+"...",
						reward.Lamports,
						reward.RewardType)
				}
			}

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
	fmt.Println("✅ Block subscription example completed")
}
