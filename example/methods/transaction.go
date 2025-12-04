package methods

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"google.golang.org/grpc/credentials"
)

func SubscribeTransactions(endpoint string, token string) {
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
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"tx_filter": {
				AccountInclude: []string{},
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

			fmt.Printf("\n💸 Transaction Update:\n")
			fmt.Printf("   Signature: %s\n", base64.StdEncoding.EncodeToString(tx.Signature))
			fmt.Printf("   Slot: %d\n", txUpdate.Slot)
			fmt.Printf("   Is Vote: %v\n", tx.IsVote)

			if tx.Meta != nil {
				fmt.Printf("   Fee: %d lamports\n", tx.Meta.Fee)
				if tx.Meta.Err != nil {
					fmt.Printf("   Error: %v\n", tx.Meta.Err)
				} else {
					fmt.Printf("   Status: Success\n")
				}
				fmt.Printf("   Pre Balances: %v\n", tx.Meta.PreBalances)
				fmt.Printf("   Post Balances: %v\n", tx.Meta.PostBalances)
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

	time.Sleep(30 * time.Second)
	grpcClient.Close()
	fmt.Println("✅ Transaction subscription example completed")
}
