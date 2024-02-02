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
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/holiman/uint256"
	"github.com/joho/godotenv"
)

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

func prepareTransactionParams(client *ethclient.Client, privateKey *ecdsa.PrivateKey) (uint64, *big.Int, *big.Int, *big.Int, error) {
	publicKey := privateKey.PublicKey
	fromAddress := crypto.PubkeyToAddress(publicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error getting nonce: %v", err)
	}

	head, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error getting header: %v", err)
	}

	baseFee := head.BaseFee
	suggestedTip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error suggesting gas tip cap: %v", err)
	}

	tip := new(big.Int).Add(suggestedTip, big.NewInt(10e9))
	maxFeePerGas := new(big.Int).Add(baseFee, tip)
	maxFeePerGas.Add(maxFeePerGas, big.NewInt(10e9))

	chainId := big.NewInt(5)

	return nonce, chainId, tip, maxFeePerGas, nil
}

func createBlobTx(chainID *big.Int, nonce uint64, tip *big.Int, maxFeePerGas *big.Int) (*types.Transaction, error) {
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
		GasFeeCap:  uint256.MustFromBig(maxFeePerGas),
		Gas:        25000000,
		To:         common.HexToAddress("0xF5106D4ef61cd0a04a345495f59f536bB7cd6074"),
		Value:      uint256.NewInt(0),
		Data:       make([]byte, 50),
		BlobFeeCap: uint256.NewInt(15),
		BlobHashes: sidecar.BlobHashes(),
		Sidecar:    &sidecar,
	}), nil
}

func sendTransaction(client *ethclient.Client, tx *types.Transaction, privateKey *ecdsa.PrivateKey) {
	signedTx, err := types.SignTx(tx, types.NewCancunSigner(tx.ChainId()), privateKey)
	if err != nil {
		log.Fatalf("Error signing transaction: %v", err)
	}

	rlpData, err := signedTx.MarshalBinary()
	if err != nil {
		log.Fatalf("Error marshaling signed transaction: %v", err)
	}

	err = client.Client().CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData))
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	} else {
		log.Printf("Successfully sent transaction. txhash= %s", signedTx.Hash().Hex())
	}
}
