package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/lightningnetwork/lnd/lnrpc_client/lnrpc"

	"github.com/joho/godotenv"
	"github.com/lightningnetwork/lnd/macaroons"
	"gopkg.in/macaroon.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ErrorCode int

const ErrorWalletIsUnlocked ErrorCode = iota + 1 // https://github.com/golang/go/wiki/Iota

var errMessage = map[ErrorCode]string{
	ErrorWalletIsUnlocked: "rpc error: code = Unimplemented desc = unknown service lnrpc.WalletUnlocker",
}

var (
	cipherSeedMnemonic []string
	walletPassword     string
	tlsCertPath        string
	macaroonPath       string
)

func walletIsNotExist() bool {

	if _, err := os.Stat(tlsCertPath); os.IsNotExist(err) {
		return true
	}

	if _, err := os.Stat(macaroonPath); os.IsNotExist(err) {
		return true
	}

	return false
}

func main() {

	err := godotenv.Load()

	if err != nil {
		fmt.Println("Error loading .env vars")
	}

	ctx := context.Background()

	// Load env
	seedStr := os.Getenv("CIPHER_SEED_MNEMONIC")
	walletPassword = os.Getenv("WALLET_PASSWORD")
	tlsCertPath = os.Getenv("TLS_CERT_PATH")
	macaroonPath = os.Getenv("MACAROON_PATH")

	tlsCreds, err := credentials.NewClientTLSFromFile(tlsCertPath, "")
	if err != nil {
		fmt.Println("Cannot get node tls credentials", err)
		return
	}

	// Check wallet is existed or not to create
	if walletIsNotExist() {
		cipherSeedMnemonic = strings.Split(seedStr, " ")

		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(tlsCreds),
			grpc.WithBlock(),
		}

		conn, err := grpc.Dial("localhost:10009", opts...)
		if err != nil {
			fmt.Println("cannot dial to lnd", err)
			return
		}

		walletunlocker := lnrpc.NewWalletUnlockerClient(conn)

		_, err = walletunlocker.InitWallet(ctx, &lnrpc.InitWalletRequest{WalletPassword: []byte(walletPassword), CipherSeedMnemonic: cipherSeedMnemonic})
		if err != nil {
			fmt.Println("Cannot create wallet", err)
		} else {
			fmt.Println("Created wallet successfully!")
		}
		return
	}

	// Unlock wallet
	macaroonBytes, err := ioutil.ReadFile(macaroonPath)
	if err != nil {
		fmt.Println("Cannot read macaroon file", err)
		return
	}

	mac := &macaroon.Macaroon{}
	if err = mac.UnmarshalBinary(macaroonBytes); err != nil {
		fmt.Println("Cannot unmarshal macaroon", err)
		return
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithBlock(),
		grpc.WithPerRPCCredentials(macaroons.NewMacaroonCredential(mac)),
	}

	conn, err := grpc.Dial("localhost:10009", opts...)
	if err != nil {
		fmt.Println("cannot dial to lnd", err)
		return
	}

	walletunlocker := lnrpc.NewWalletUnlockerClient(conn)
	_, err = walletunlocker.UnlockWallet(ctx, &lnrpc.UnlockWalletRequest{WalletPassword: []byte(walletPassword)})
	if err != nil {
		if err.Error() == errMessage[ErrorWalletIsUnlocked] {
			fmt.Println("wallet is unlocked already")
			return
		}
	} else {
		fmt.Println("wallet is unlocked successfully!")
	}
}
