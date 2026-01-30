package debug

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
)

func DetectMissingBlocks(endpoint string, token string) {
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
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("🔍 Starting missing block detection...")

	var (
		lastSlot   uint64
		mu         sync.Mutex
		firstBlock = true
	)

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Block:
			blockUpdate := update.GetBlock()
			currentSlot := blockUpdate.Slot

			mu.Lock()
			defer mu.Unlock()

			if firstBlock {
				// Initialize with the first block received
				lastSlot = currentSlot
				firstBlock = false
				fmt.Printf("✅ Initial block: Slot %d\n", currentSlot)
			} else {
				// Check if there are missing blocks
				expectedSlot := lastSlot + 1

				if currentSlot > expectedSlot {
					// Missing blocks detected
					missingCount := currentSlot - expectedSlot
					fmt.Printf("\n❌ MISSING BLOCKS DETECTED!\n")
					fmt.Printf("   Expected: %d\n", expectedSlot)
					fmt.Printf("   Received: %d\n", currentSlot)
					fmt.Printf("   Missing count: %d\n", missingCount)
					fmt.Printf("   Missing slots: ")

					// Print all missing slot numbers
					for slot := expectedSlot; slot < currentSlot; slot++ {
						if slot > expectedSlot {
							fmt.Printf(", ")
						}
						fmt.Printf("%d", slot)
					}
					fmt.Printf("\n\n")
				} else if currentSlot == expectedSlot {
					// Normal sequential block
					fmt.Printf("✓ Slot %d (sequential)\n", currentSlot)
				} else if currentSlot < lastSlot {
					// Out of order block (should be rare)
					fmt.Printf("⚠️  Out of order block: Slot %d (last was %d)\n", currentSlot, lastSlot)
					return nil // Don't update lastSlot for out-of-order blocks
				}

				lastSlot = currentSlot
			}

		case *pb.SubscribeUpdate_Ping:
			return nil

		case *pb.SubscribeUpdate_Pong:
			return nil

		default:
			// Ignore other update types
		}
		return nil
	})

	// Run for a longer time to detect missing blocks
	fmt.Println("📊 Monitoring for missing blocks (Press Ctrl+C to stop)...")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	grpcClient.Close()
	fmt.Println("\n✅ Missing block detection completed")
}
