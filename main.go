package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
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

	blobTx, err := createBlobTx(chainID, nonce, tip, maxFeePerGas)
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

func prepareTransactionParams(client *ethclient.Client, privateKey *ecdsa.PrivateKey) (uint64, *big.Int, *big.Int, *uint256.Int, error) {
	publicKey := privateKey.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error getting nonce: %v", err)
	}

	suggestedTip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error suggesting gas tip cap: %v", err)
	}

	tip := new(big.Int).Add(suggestedTip, big.NewInt(10e9))

	val, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("Error getting suggested gas price: %v", err)
	}
	var nok bool
	maxFeePerGas, nok := uint256.FromBig(val)
	if nok {
		log.Fatalf("gas price is too high! got %v", val.String())
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error fetching chain id %v", err)
	}
	return nonce, chainID, tip, maxFeePerGas, nil
}

func createBlobTx(chainID *big.Int, nonce uint64, tip *big.Int, maxFeePerGas *uint256.Int) (*types.Transaction, error) {
	emptyBlob := kzg4844.Blob{}
	emptyBlobCommit, _ := kzg4844.BlobToCommitment(emptyBlob)
	emptyBlobProof, _ := kzg4844.ComputeBlobProof(emptyBlob, emptyBlobCommit)
	sidecar := types.BlobTxSidecar{
		Blobs:       []kzg4844.Blob{emptyBlob},
		Commitments: []kzg4844.Commitment{emptyBlobCommit},
		Proofs:      []kzg4844.Proof{emptyBlobProof},
	}

	return types.NewTx(&types.BlobTx{
		ChainID:    uint256.MustFromBig(chainID),
		Nonce:      nonce,
		GasTipCap:  uint256.MustFromBig(tip),
		GasFeeCap:  maxFeePerGas,
		Gas:        250000,
		To:         common.HexToAddress(TO_ADDRESS),
		Value:      uint256.NewInt(0),
		Data:       make([]byte, 128),
		BlobFeeCap: uint256.NewInt(1e10),
		BlobHashes: sidecar.BlobHashes(),
		Sidecar:    &sidecar,
	}), nil
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

	// signedTx, err := types.SignTx(tx, types.NewCancunSigner(tx.ChainId()), privateKey)
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

func encodeBlobs(data []byte) []kzg4844.Blob {
	blobs := []kzg4844.Blob{{}}
	blobIndex := 0
	fieldIndex := -1
	for i := 0; i < len(data); i += 31 {
		fieldIndex++
		if fieldIndex == params.BlobTxFieldElementsPerBlob {
			blobs = append(blobs, kzg4844.Blob{})
			blobIndex++
			fieldIndex = 0
		}
		max := i + 31
		if max > len(data) {
			max = len(data)
		}
		copy(blobs[blobIndex][fieldIndex*32:], data[i:max])
	}
	return blobs
}
