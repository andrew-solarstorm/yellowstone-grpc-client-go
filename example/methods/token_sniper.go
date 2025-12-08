package methods

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	yellowstone "github.com/andrew-solarstorm/yellowstone-grpc-client-go"
	pb "github.com/andrew-solarstorm/yellowstone-grpc-client-go/proto"
	"github.com/gagliardetto/solana-go"
)

const (
	TOKEN_PROGRAM       = "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"
	TOKEN_2022_PROGRAM  = "TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb"
	RAYDIUM_AMM_V4      = "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8"
	PUMP_FUN_PROGRAM    = "6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P"
	METAPLEX_TOKEN_META = "metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s"
)

func TokenSniper(endpoint string, token string) {
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
			"token_accounts": {
				Owner: []string{TOKEN_PROGRAM, TOKEN_2022_PROGRAM},
			},
		},
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"token_txs": {
				AccountInclude: []string{
					RAYDIUM_AMM_V4,
					PUMP_FUN_PROGRAM,
					METAPLEX_TOKEN_META,
				},
			},
		},
	}

	stream, err := grpcClient.SubscribeWithRequest(context.Background(), req)
	if err != nil {
		log.Fatalf("Error subscribing: %v", err)
	}

	fmt.Println("🎯 Token Sniper Active - Monitoring for new token creations...")
	fmt.Println("📡 Watching: Raydium, Pump.fun, Token Programs")
	fmt.Println()

	tokenCreations := make(map[string]bool)

	go grpcClient.Start(stream, func(update *pb.SubscribeUpdate) error {
		switch update.GetUpdateOneof().(type) {
		case *pb.SubscribeUpdate_Account:
			accountUpdate := update.GetAccount()
			account := accountUpdate.Account

			if len(account.Data) >= 82 && !accountUpdate.IsStartup {
				mintPubkey := solana.PublicKeyFromBytes(account.Pubkey).String()

				if !tokenCreations[mintPubkey] {
					tokenCreations[mintPubkey] = true

					fmt.Printf("🆕 NEW TOKEN DETECTED!\n")
					fmt.Printf("   Mint: %s\n", mintPubkey)
					fmt.Printf("   Owner: %s\n", solana.PublicKeyFromBytes(account.Owner).String())
					fmt.Printf("   Slot: %d\n", accountUpdate.Slot)
					fmt.Printf("   Data Size: %d bytes\n", len(account.Data))
					fmt.Printf("   Timestamp: %s\n", time.Now().Format(time.RFC3339))
					fmt.Println()
				}
			}

		case *pb.SubscribeUpdate_Transaction:
			txUpdate := update.GetTransaction()
			tx := txUpdate.Transaction

			if tx.Meta != nil && tx.Meta.Err == nil {
				sig := base64.StdEncoding.EncodeToString(tx.Signature)

				isPumpFun := false
				isRaydium := false
				hasTokenProgram := false

				if tx.Transaction != nil && tx.Transaction.Message != nil {
					for _, key := range tx.Transaction.Message.AccountKeys {
						keyStr := solana.PublicKeyFromBytes(key).String()
						if keyStr == PUMP_FUN_PROGRAM {
							isPumpFun = true
						}
						if keyStr == RAYDIUM_AMM_V4 {
							isRaydium = true
						}
						if keyStr == TOKEN_PROGRAM || keyStr == TOKEN_2022_PROGRAM {
							hasTokenProgram = true
						}
					}
				}

				if (isPumpFun || isRaydium) && hasTokenProgram {
					fmt.Printf("🚀 TOKEN LAUNCH TRANSACTION\n")
					fmt.Printf("   Signature: %s\n", sig[:32]+"...")
					fmt.Printf("   Slot: %d\n", txUpdate.Slot)

					if isPumpFun {
						fmt.Printf("   Platform: Pump.fun\n")
					} else if isRaydium {
						fmt.Printf("   Platform: Raydium\n")
					}

					if tx.Transaction != nil && tx.Transaction.Message != nil {
						fmt.Printf("   Accounts: %d\n", len(tx.Transaction.Message.AccountKeys))
						fmt.Printf("   Instructions: %d\n", len(tx.Transaction.Message.Instructions))
					}

					fmt.Printf("   Fee: %d lamports\n", tx.Meta.Fee)
					fmt.Printf("   Compute Units: %d\n", tx.Meta.ComputeUnitsConsumed)
					fmt.Printf("   Timestamp: %s\n", time.Now().Format(time.RFC3339))
					fmt.Println()
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("Error starting client: %v", err)
	}

	time.Sleep(300 * time.Second)
	fmt.Printf("\n📊 Session Summary:\n")
	fmt.Printf("   Total unique tokens detected: %d\n", len(tokenCreations))
	fmt.Println("✅ Completed")
}

func readUint64LE(b []byte) uint64 {
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}
