package methods

import (
	"context"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

func SubscribeBlockMeta(endpoint string, token string) {
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
		BlocksMeta: map[string]*pb.SubscribeRequestFilterBlocksMeta{
			"block_meta_filter": {},
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

	fmt.Println("📊 Listening for block metadata updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_BlockMeta:
			blockMeta := update.GetBlockMeta()

			fmt.Printf("\n📊 Block Metadata Update:\n")
			fmt.Printf("   Slot: %d\n", blockMeta.Slot)
			fmt.Printf("   Blockhash: %s\n", blockMeta.Blockhash)
			fmt.Printf("   Parent Slot: %d\n", blockMeta.ParentSlot)
			fmt.Printf("   Parent Blockhash: %s\n", blockMeta.ParentBlockhash)
			if blockMeta.BlockHeight != nil {
				fmt.Printf("   Block Height: %d\n", blockMeta.BlockHeight.BlockHeight)
			}
			if blockMeta.BlockTime != nil {
				fmt.Printf("   Block Time: %d\n", blockMeta.BlockTime.Timestamp)
			}
			fmt.Printf("   Executed Transaction Count: %d\n", blockMeta.ExecutedTransactionCount)

			if blockMeta.Rewards != nil && blockMeta.Rewards.Rewards != nil && len(blockMeta.Rewards.Rewards) > 0 {
				fmt.Printf("   Rewards Count: %d\n", len(blockMeta.Rewards.Rewards))
				fmt.Printf("   First few rewards:\n")
				for i, reward := range blockMeta.Rewards.Rewards {
					if i >= 3 {
						fmt.Printf("      ... and %d more\n", len(blockMeta.Rewards.Rewards)-3)
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
	fmt.Println("✅ Block metadata subscription example completed")
}
