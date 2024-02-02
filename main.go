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

	chainId := uint256.MustFromBig(big.NewInt(5))

	blobtx := types.NewTx(&types.BlobTx{
		ChainID:    chainId,
		Nonce:      5,
		GasTipCap:  uint256.NewInt(2),
		GasFeeCap:  uint256.NewInt(2),
		Gas:        25000,
		To:         common.HexToAddress("0x06fd9d0Ae9052A85989D0A30c60fB11753537f9A"),
		Value:      uint256.NewInt(99),
		Data:       make([]byte, 50),
		BlobFeeCap: uint256.NewInt(15),
		BlobHashes: sidecar.BlobHashes(),
	})

	privateKey, err := getECDSAPrivateKey()

	signedTx, _ := types.SignTx(blobtx, types.NewCancunSigner(big.NewInt(5)), privateKey)

	if err != nil {
		fmt.Printf("error doing RLP encoding")
	}

	rlpData, _ := signedTx.MarshalBinary()

	ctx := context.Background()
	client, err := ethclient.DialContext(ctx, os.Getenv("NODE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	err = client.Client().CallContext(context.Background(), nil, "eth_sendRawTransaction", hexutil.Encode(rlpData))

	if err != nil {
		log.Fatalf("failed to send transaction: %v", err)
	} else {
		log.Printf("successfully sent transaction. txhash=%v", signedTx.Hash())
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

// func generateAndAppendSignatureValues(tx *types.BlobTx, encodedTx *[]byte, privateKey *ecdsa.PrivateKey) {
// 	signature, err := crypto.Sign(crypto.Keccak256Hash(*encodedTx).Bytes(), privateKey)
// 	if err != nil {
// 		fmt.Println("error singing the encoded tx object")
// 		return
// 	}

// 	r, overflow := uint256.FromBig(new(big.Int).SetBytes(signature[:32]))
// 	if overflow {
// 		return
// 	}

// 	s, overflow := uint256.FromBig(new(big.Int).SetBytes(signature[32:64]))
// 	if overflow {
// 		return
// 	}

// 	vByte := signature[64]
// 	if vByte == 0 || vByte == 1 {
// 		vByte += 27 // Adjust according to Ethereum's v value conventions
// 	}
// 	v := uint256.NewInt(uint64(vByte))

// 	tx.Value = v
// 	tx.R = r
// 	tx.S = s
// }

// func generateRandomHashes(n int) [][]byte {
// 	hashes := make([][]byte, n)
// 	for i := 0; i < n; i++ {
// 		hash := make([]byte, 32)
// 		_, err := rand.Read(hash)
// 		if err != nil {
// 			panic("error generating random hash")
// 		}
// 		hashes[i] = hash
// 	}
// 	return hashes
// }

// func createAndSendTransaction(txHex string) {
// 	payload := map[string]interface{}{
// 		"jsonrpc": "2.0",
// 		"method":  "eth_sendRawTransaction",
// 		"params":  []interface{}{txHex},
// 		"id":      5,
// 	}
// 	requestBody, err := json.Marshal(payload)
// 	if err != nil {
// 		fmt.Println("Error marshaling request:", err)
// 		return
// 	}

// 	req, err := http.NewRequest("POST", os.Getenv("NODE_URL"), bytes.NewBuffer(requestBody))
// 	if err != nil {
// 		fmt.Println("Error creating request:", err)
// 		return
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println("Error sending request:", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	responseBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println("Error reading response body:", err)
// 		return
// 	}

// 	fmt.Printf("Response: %s\n", responseBody)
// }
