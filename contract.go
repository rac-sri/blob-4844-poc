package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

var contractAddress = "0x"

func CheckLatestRequestId(client *ethclient.Client) (*big.Int, error) {

	parsedABI, address, err := ParseABI()

	if err != nil {
		return nil, err
	}

	query := ethereum.FilterQuery{
		FromBlock: nil, // Use nil for latest or specify a block number
		ToBlock:   nil,
		Addresses: []common.Address{*address},
	}

	logs, err := client.FilterLogs(context.Background(), query)
	fmt.Println("logs", err, logs)
	if err != nil {
		return nil, fmt.Errorf("error fetching logs: %v", err)
	}

	var requestId *big.Int

	for _, vLog := range logs {
		_, err := parsedABI.Unpack("NewReceipt", vLog.Data)
		if err != nil {
			log.Printf("Failed to unpack log data: %v", err)
			continue
		}

		requestId = big.NewInt(0).SetBytes(vLog.Topics[1].Bytes())
	}

	return requestId, nil
}
func GetMatrices(client *ethclient.Client, requestId *big.Int) ([2][3][3]*big.Int, error) {
	parsedABI, address, err := ParseABI()
	if err != nil {
		return [2][3][3]*big.Int{}, fmt.Errorf("failed to parse ABI: %v", err)
	}

	callData, err := parsedABI.Pack("getMatrices", requestId)
	if err != nil {
		return [2][3][3]*big.Int{}, fmt.Errorf("failed to pack call data for getMatrices: %v", err)
	}

	msg := ethereum.CallMsg{To: address, Data: callData}
	res, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return [2][3][3]*big.Int{}, fmt.Errorf("failed to call getMatrices: %v", err)
	}

	var matrices [2][3][3]*big.Int
	err = parsedABI.UnpackIntoInterface(&matrices, "getMatrices", res)
	if err != nil {
		return [2][3][3]*big.Int{}, fmt.Errorf("failed to unpack response from getMatrices: %v", err)
	}

	return matrices, nil
}

func generateSubmitSolutionCalldata(root []byte, matrixMul [3][3]*big.Int, requestId *big.Int) []byte {

	parsedABI, _, err := ParseABI()

	if err != nil {
		log.Fatal("cannot parse abi")
	}

	input, err := parsedABI.Pack("submitResult", root, matrixMul, requestId)

	return input
}

func ParseABI() (*abi.ABI, *common.Address, error) {
	abiPath := "./abi.json"
	abiBytes, err := os.ReadFile(abiPath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading ABI file: %v", err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse contract ABI: %v", err)
	}

	address := common.HexToAddress(contractAddress)
	return &parsedABI, &address, nil
}
