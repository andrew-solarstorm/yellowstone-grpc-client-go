package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andrew-solarstorm/yellowstone-grpc-client-go/example/methods"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	token := os.Getenv("TOKEN")
	endpoint := os.Getenv("ENDPOINT")
	if token == "" || endpoint == "" {
		log.Fatalf("TOKEN and ENDPOINT must be set")
	}

	example := "slot"
	if len(os.Args) > 1 {
		example = os.Args[1]
	}

	fmt.Printf("🚀 Running Example: %s\n", example)
	fmt.Printf("📡 Endpoint: %s\n\n", endpoint)

	switch example {
	case "slot", "slots":
		methods.SubscribeSlot(endpoint, token)
	case "account", "accounts":
		methods.SubscribeAccounts(endpoint, token)
	case "transaction", "transactions", "tx":
		methods.SubscribeTransactions(endpoint, token)
	case "block", "blocks":
		methods.SubscribeBlocks(endpoint, token)
	case "block_meta", "blockmeta", "meta":
		methods.SubscribeBlockMeta(endpoint, token)
	case "token_sniper", "sniper", "token":
		methods.TokenSniper(endpoint, token)
	default:
		fmt.Println("Available examples:")
		fmt.Println("  slot")
		fmt.Println("  account")
		fmt.Println("  transaction")
		fmt.Println("  block")
		fmt.Println("  block_meta")
		fmt.Println("  token_sniper")
		fmt.Println("\nUsage: go run . [example]")
		os.Exit(1)
	}
}
