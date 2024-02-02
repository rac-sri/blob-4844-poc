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

type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

// specifications -> https://eips.ethereum.org/EIPS/eip-4844

var (
	emptyBlob          = kzg4844.Blob{}
	emptyBlobCommit, _ = kzg4844.BlobToCommitment(emptyBlob)
	emptyBlobProof, _  = kzg4844.ComputeBlobProof(emptyBlob, emptyBlobCommit)
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// https://github.com/ethereum/go-ethereum/blob/master/core/types/tx_blob.go
	sidecar := types.BlobTxSidecar{
		Blobs:       []kzg4844.Blob{emptyBlob},
		Commitments: []kzg4844.Commitment{emptyBlobCommit},
		Proofs:      []kzg4844.Proof{emptyBlobProof},
	}

	ctx := context.Background()
	client, _ := ethclient.DialContext(ctx, os.Getenv("NODE_URL"))

	privateKey, _ := getECDSAPrivateKey()
	publicKey := privateKey.PublicKey

	nonce, err := client.PendingNonceAt(ctx, crypto.PubkeyToAddress(publicKey))
	if err != nil {
		log.Fatalf("Error getting nonce: %v", err)
	}

	head, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	baseFee := head.BaseFee
	suggestedTip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	tip := new(big.Int).Add(suggestedTip, big.NewInt(10e9))
	maxFeePerGas := new(big.Int).Add(baseFee, tip)
	maxFeePerGas.Add(maxFeePerGas, big.NewInt(10e9))

	chainId := uint256.MustFromBig(big.NewInt(5))

	blobtx := types.NewTx(&types.BlobTx{
		ChainID:    chainId,
		Nonce:      nonce,
		GasTipCap:  uint256.MustFromBig(tip),
		GasFeeCap:  uint256.MustFromBig(maxFeePerGas),
		Gas:        2500000,
		To:         common.HexToAddress("0xF5106D4ef61cd0a04a345495f59f536bB7cd6074"),
		Value:      uint256.NewInt(0),
		Data:       make([]byte, 50),
		BlobFeeCap: uint256.NewInt(15),
		BlobHashes: sidecar.BlobHashes(),
		Sidecar:    &sidecar,
	})

	signedTx, _ := types.SignTx(blobtx, types.NewCancunSigner(big.NewInt(5)), privateKey)

	if err != nil {
		fmt.Printf("error doing RLP encoding")
	}

	rlpData, _ := signedTx.MarshalBinary()

	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	err = client.Client().CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData))

	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	} else {
		log.Printf("successfully sent transaction. txhash= %v", signedTx.Hash())
	}

}

func getECDSAPrivateKey() (*ecdsa.PrivateKey, error) {
	privateKeyHex := os.Getenv("PRIVATE_KEY")

	privatKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex string: %v", err)
	}

	privateKey, err := crypto.ToECDSA(privatKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA private key: %v", err)
	}

	return privateKey, nil
}
