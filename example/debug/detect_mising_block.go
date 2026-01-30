package debug

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

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

	// Open file for writing missing blocks
	file, err := os.OpenFile("missing_blocks.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening missing_blocks.txt: %v", err)
	}
	defer file.Close()

	// Write header with timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(file, "\n=== Missing Block Detection Started at %s ===\n", timestamp)
	fmt.Println("📝 Writing missing blocks to missing_blocks.txt")

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
					timestamp := time.Now().Format("2006-01-02 15:04:05")
					
					// Print to console
					fmt.Printf("\n❌ MISSING BLOCKS DETECTED!\n")
					fmt.Printf("   Time: %s\n", timestamp)
					fmt.Printf("   Expected: %d\n", expectedSlot)
					fmt.Printf("   Received: %d\n", currentSlot)
					fmt.Printf("   Missing count: %d\n", missingCount)
					fmt.Printf("   Missing slots: ")

					// Write to file and collect missing slots
					fmt.Fprintf(file, "\n[%s] MISSING BLOCKS DETECTED\n", timestamp)
					fmt.Fprintf(file, "Expected: %d, Received: %d, Missing count: %d\n", expectedSlot, currentSlot, missingCount)
					fmt.Fprintf(file, "Missing slots: ")
					
					// Print and write all missing slot numbers
					for slot := expectedSlot; slot < currentSlot; slot++ {
						if slot > expectedSlot {
							fmt.Printf(", ")
							fmt.Fprintf(file, ", ")
						}
						fmt.Printf("%d", slot)
						fmt.Fprintf(file, "%d", slot)
					}
					fmt.Printf("\n\n")
					fmt.Fprintf(file, "\n")
					
					// Ensure data is written to disk
					file.Sync()
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
	
	// Write closing message to file
	timestamp = time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(file, "\n=== Missing Block Detection Stopped at %s ===\n", timestamp)
	file.Sync()
	
	grpcClient.Close()
	fmt.Println("\n✅ Missing block detection completed")
	fmt.Println("📝 Results saved to missing_blocks.txt")
}
