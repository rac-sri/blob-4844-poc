package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

const TO_ADDRESS = "0xF5106D4ef61cd0a04a345495f59f536bB7cd6074"

func main() {
	if err := loadEnv(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	client, err := ethclient.DialContext(context.Background(), os.Getenv("NODE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	privateKey, err := getECDSAPrivateKey(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		log.Fatalf("Error getting private key: %v", err)
	}

	nonce, chainID, tip, maxFeePerGas, err := prepareTransactionParams(client, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	requestId, err := CheckLatestRequestId(client)
	fmt.Print(requestId, err)
	if err != nil {
		return
	}
	requestId = big.NewInt(10)
	// matrices, err := GetMatrices(client, requestId)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Print(requestId)

	// matrixMul, singletonArray := MultiplyMatrices(matrices[0], matrices[1])

	matrixMul := [3][3]*big.Int{
		{big.NewInt(1), big.NewInt(2), big.NewInt(3)},
		{big.NewInt(4), big.NewInt(5), big.NewInt(6)},
		{big.NewInt(7), big.NewInt(8), big.NewInt(9)},
	}

	singletonArray := [9]*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5), big.NewInt(6), big.NewInt(7), big.NewInt(8), big.NewInt(9)}
	merkleRoot := MerkleTreeRoot(singletonArray[:])

	input := generateSubmitSolutionCalldata(merkleRoot, matrixMul, requestId)

	blobTx, err := createBlobTx(chainID, nonce, tip, maxFeePerGas, merkleRoot, input)
	if err != nil {
		log.Fatal("Failed to create blob transaction:", err)
	}

	sendTransaction(client, blobTx, privateKey)
}

func loadEnv() error {
	return godotenv.Load()
}

func getECDSAPrivateKey(privateKeyHex string) (*ecdsa.PrivateKey, error) {
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}
	return crypto.ToECDSA(privateKeyBytes)
}

func sendTransaction(client *ethclient.Client, tx *types.Transaction, privateKey *ecdsa.PrivateKey) {
	log.Printf("Transaction before signing: %+v", tx)
	fmt.Println("blob gas ", tx.BlobGas())
	fmt.Println("blob gas fee cap ", tx.BlobGasFeeCap())
	fmt.Println("blob hashes ", tx.BlobHashes())
	//fmt.Println("blob tx sidecar", tx.BlobTxSidecar())
	fmt.Println("cost", tx.Cost())
	fmt.Println("gas tip cap", tx.GasTipCap())
	fmt.Println("gas", tx.Gas())
	fmt.Println("gas price", tx.GasPrice())
	fmt.Println("type", tx.Type())

	_, _ = types.SignTx(tx, types.NewCancunSigner(tx.ChainId()), privateKey)
	// fmt.Println("signed tx", signedTx)
	// if err != nil {
	// 	log.Fatalf("Error signing transaction: %v", err)
	// }

	// err = client.SendTransaction(context.Background(), signedTx)
	// if err != nil {
	// 	log.Fatalf("Failed to send transaction: %v", err)
	// } else {
	// 	log.Printf("Successfully sent transaction. txhash= %s", signedTx.Hash().Hex())
	// }
}
