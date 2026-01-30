package methods

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"github.com/gagliardetto/solana-go"
)

func SubscribeTransactions(endpoint string, token string) {
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
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"tx_filter": {
				AccountInclude: []string{
					"FLUXubRmkEi2q6K3Y9kBPg9248ggaZVsoSFhtJHSrm1X",
				},
			},
		},
	}

	ctx := context.Background()
	stream, err := grpcClient.SubscribeWithRequest(ctx, req)
	if err != nil {
		log.Fatalf("Error subscribing to geyser: %v", err)
	}

	fmt.Println("💸 Listening for transaction updates...")

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Transaction:
			txUpdate := update.GetTransaction()
			tx := txUpdate.Transaction

			if len(tx.Meta.GetPostTokenBalances()) == 0 {
				return nil
			}

			fmt.Printf("\n💸 Transaction Update:\n")
			fmt.Printf("   Signature: %s\n", solana.SignatureFromBytes(tx.Signature).String())
			fmt.Printf("   Slot: %d\n", txUpdate.Slot)
			fmt.Printf("   Is Vote: %v\n", tx.IsVote)

			if tx.Meta != nil {
				fmt.Printf("   Fee: %d lamports\n", tx.Meta.Fee)
				if tx.Meta.Err != nil {
					fmt.Printf("   Error: %v\n", tx.Meta.Err)
				} else {
					fmt.Printf("   Status: Success\n")
				}
				fmt.Printf("   Pre Balances: %v\n", tx.Meta.GetPreTokenBalances())
				fmt.Printf("   Post Balances: %v\n", tx.Meta.GetPostTokenBalances())

				fmt.Println("    POST Ballances: ", tx.Meta.GetPostBalances())
				fmt.Printf("   Compute Units Consumed: %d\n", tx.Meta.ComputeUnitsConsumed)
			}

			if tx.Transaction != nil {
				if len(tx.Transaction.Signatures) > 0 {
					fmt.Printf("   Signatures Count: %d\n", len(tx.Transaction.Signatures))
				}
				if tx.Transaction.Message != nil {
					fmt.Printf("   Account Keys Count: %d\n", len(tx.Transaction.Message.AccountKeys))
					fmt.Printf("   Instructions Count: %d\n", len(tx.Transaction.Message.Instructions))
				}
			}

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
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<- sigChan
	grpcClient.Close()
	fmt.Println("✅ Transaction subscription example completed")
}
