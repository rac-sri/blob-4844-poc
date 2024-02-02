package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var contractAddress = "0x"
var contractABI = `[CONTRACTABI]`

func CheckLatestEvents(client *ethclient.Client) {
	address := common.HexToAddress(contractAddress)

	// Load the contract ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		log.Fatalf("Failed to parse contract ABI: %v", err)
	}

	query := ethereum.FilterQuery{
		Addresses: []common.Address{address},
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Listening to contract events")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Println("log found")
			event, err := parsedABI.Unpack("[EVENT_NAME]", vLog.Data)
			if err != nil {
				log.Println("Error unpacking event:", err)
				continue
			}

			fmt.Println("Event:", event)

			// TODO: process the event data
		}
	}
}
