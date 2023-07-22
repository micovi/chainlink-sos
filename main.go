package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"strings"

	"github.com/micovi/chainlink-sos/contracts/Feed"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var pk string
var npk string
var rpc string
var chainid int64
var file string
var typee string
var output string
var address1 common.Address
var address2 common.Address
var privateKey *ecdsa.PrivateKey
var privateKey2 *ecdsa.PrivateKey
var client *ethclient.Client
var feedContract *Feed.Feed
var txs []string

func init() {
	flag.StringVar(&pk, "private-key", "", "Private key")
	flag.StringVar(&rpc, "rpc", "", "rpc")
	flag.Int64Var(&chainid, "chain-id", 0, "chainid")
	flag.StringVar(&npk, "new-private-key", "", "New Private key")
	flag.StringVar(&file, "file", "", "File")
	flag.StringVar(&typee, "type", "", "Type")
	flag.StringVar(&output, "output", "screen", "Output")
	flag.Parse()

	if pk == "" {
		log.Fatal("private-key is required")
	}

	if rpc == "" {
		log.Fatal("rpc is required")
	}

	if chainid == 0 {
		log.Fatal("chainid is required")
	}

	if npk == "" {
		log.Fatal("new-private-key is required")
	}

	if file == "" {
		log.Fatal("file is required")
	}

	if typee == "" || typee != "transfer" && typee != "accept" {
		log.Fatal("type should be transfer or accept")
	}
}

func runTransferPayeeship(feed string, transmitter string, privateKey *ecdsa.PrivateKey, address2 common.Address) {
	log.Println("Running TransferPayeeship ...")
	nonce, err := client.PendingNonceAt(context.Background(), address1)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Nonce: ", nonce)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Print("GasPrice: ", gasPrice)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainid))
	if err != nil {
		log.Fatal(err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(150000) // in units
	auth.GasPrice = gasPrice

	parsedABI, err := abi.JSON(strings.NewReader(Feed.FeedMetaData.ABI))
	if err != nil {
		log.Fatal(err)
	}
	inputData, err := parsedABI.Pack("transferPayeeship", common.HexToAddress(transmitter), address2)

	if err != nil {
		log.Fatal(err)
	}

	var add = common.HexToAddress(feed)

	// Estimate gas
	msg := ethereum.CallMsg{
		From: auth.From,
		To:   &add,
		Data: inputData,
	}

	estimatedGas, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Estimated gas: %v\n", estimatedGas)
	//auth.GasLimit = estimatedGas
	log.Printf("Gas limit: %v\n", auth.GasLimit)

	feedContract, err := Feed.NewFeed(common.HexToAddress(feed), client)
	if err != nil {
		log.Fatalf("Failed to instantiate Feed contract: %v", err)
	}
	log.Println("Loaded Feed contract", feed)

	tx, err := feedContract.TransferPayeeship(auth,
		common.HexToAddress(transmitter),
		address2,
	)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Tx pending: ", tx.Hash().Hex())

	txs = append(txs, tx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Transaction mined: ", receipt.Status, receipt.GasUsed)
}

func runAcceptPayeeship(feed string, transmitter string, privateKey2 *ecdsa.PrivateKey, address2 common.Address) {
	log.Println("Running AcceptPayeeship ...")
	nonce, err := client.PendingNonceAt(context.Background(), address2)
	if err != nil {
		log.Fatal(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	auth2, err := bind.NewKeyedTransactorWithChainID(privateKey2, big.NewInt(chainid))
	if err != nil {
		log.Fatal(err)
	}

	auth2.Nonce = big.NewInt(int64(nonce))
	auth2.Value = big.NewInt(0)     // in wei
	auth2.GasLimit = uint64(150000) // in units
	auth2.GasPrice = gasPrice

	parsedABI, err := abi.JSON(strings.NewReader(Feed.FeedMetaData.ABI))
	if err != nil {
		log.Fatal(err)
	}

	inputData, err := parsedABI.Pack("acceptPayeeship", common.HexToAddress(transmitter))

	if err != nil {
		log.Fatal(err)
	}

	var add = common.HexToAddress(feed)

	// Estimate gas
	msg := ethereum.CallMsg{
		From: auth2.From,
		To:   &add,
		Data: inputData,
	}

	estimatedGas, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Estimated gas: %v\n", estimatedGas)
	//auth2.GasLimit = uint64(estimatedGas)
	log.Printf("Gas limit: %v\n", auth2.GasLimit)

	feedContract, err := Feed.NewFeed(common.HexToAddress(feed), client)
	if err != nil {
		log.Fatalf("Failed to instantiate Feed contract: %v", err)
	}
	log.Println("Loaded Feed contract", feed)

	tx, err := feedContract.AcceptPayeeship(auth2, common.HexToAddress(transmitter))

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Tx pending: ", tx.Hash().Hex())

	txs = append(txs, tx.Hash().Hex())

	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Transaction mined: ", receipt.Status, receipt.GasUsed)
}

func main() {
	log.Println("Starting service ... ")

	var err error

	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		log.Fatal(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	address1 = crypto.PubkeyToAddress(*publicKeyECDSA)
	log.Print("Address1: ", address1.Hex())

	privateKey2, err := crypto.HexToECDSA(npk)
	if err != nil {
		log.Fatal(err)
	}

	publicKey2 := privateKey2.Public()
	publicKeyECDSA2, ok := publicKey2.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}

	address2 = crypto.PubkeyToAddress(*publicKeyECDSA2)
	log.Print("Address2: ", address2.Hex())

	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln("Error reading file", err)
	}

	type FeedData struct {
		Feed        string `json:"feed"`
		Transmitter string `json:"transmitter"`
	}

	var fileContents struct {
		Feeds []FeedData `json:"feeds"`
	}

	err = json.Unmarshal(content, &fileContents)
	if err != nil {
		log.Fatalln("Error reading file", err)
	}

	// loop through content feeds
	for _, feedData := range fileContents.Feeds {
		log.Println("Feed: ", feedData.Feed)
		log.Println("Transmitter: ", feedData.Transmitter)

		var ts = feedData.Transmitter
		var fs = feedData.Feed

		log.Println("Chain ID: ", chainid)

		client, err = ethclient.Dial(rpc)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Connected to RPC server")

		if typee == "transfer" {
			runTransferPayeeship(fs, ts, privateKey, address2)

			if output == "screen" {
				fmt.Println("Tx Hashes: ", txs)
			} else {
				var fc struct {
					Txs []string `json:"txs"`
				}

				fc.Txs = txs

				content, err := json.Marshal(fc)
				if err != nil {
					fmt.Println(err)
				}

				err = ioutil.WriteFile(output, content, 0644)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		if typee == "accept" {
			runAcceptPayeeship(fs, ts, privateKey2, address2)

			if output == "screen" {
				fmt.Println("Tx Hashes: ", txs)
			} else {
				var fc struct {
					Txs []string `json:"txs"`
				}

				fc.Txs = txs

				content, err := json.Marshal(fc)
				if err != nil {
					fmt.Println(err)
				}

				err = ioutil.WriteFile(output, content, 0644)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
