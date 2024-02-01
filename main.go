package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/joho/godotenv"
)

// specifications -> https://eips.ethereum.org/EIPS/eip-4844
type BlobTx struct {
	ChainID              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
	GasLimit             uint64
	To                   common.Address
	Value                *big.Int
	Data                 []byte
	AccessList           types.AccessList
	MaxFeePerBlobGas     *big.Int
	BlobVersionedHashes  [][]byte
	SignatureValues      [3][]byte
}

type Payload struct {
	ChainID              *big.Int
	Nonce                uint64
	MaxPriorityFeePerGas *big.Int
	MaxFeePerGas         *big.Int
	GasLimit             uint64
	To                   common.Address
	Value                *big.Int
	Data                 []byte
	AccessList           types.AccessList
	MaxFeePerBlobGas     *big.Int
	BlobVersionedHashes  [][]byte
}

type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tx := BlobTx{
		ChainID:              big.NewInt(5), // goerli
		Nonce:                0,
		MaxPriorityFeePerGas: big.NewInt(1000000000),
		MaxFeePerGas:         big.NewInt(1000000000),
		GasLimit:             21000,
		To:                   common.HexToAddress("0x06fd9d0Ae9052A85989D0A30c60fB11753537f9A"),
		Value:                big.NewInt(1000000000000000), // 0.001 eth
		Data:                 []byte{},
		AccessList:           types.AccessList{},
		MaxFeePerBlobGas:     big.NewInt(1000000),
		BlobVersionedHashes:  generateRandomHashes(3),
		SignatureValues:      [3][]byte{},
	}

	txEncode := Payload{
		ChainID:              big.NewInt(5), // goerli
		Nonce:                0,
		MaxPriorityFeePerGas: big.NewInt(1000000000),
		MaxFeePerGas:         big.NewInt(1000000000),
		GasLimit:             21000,
		To:                   common.HexToAddress("0x06fd9d0Ae9052A85989D0A30c60fB11753537f9A"),
		Value:                big.NewInt(1000000000000000), // 0.001 eth
		Data:                 []byte{},
		AccessList:           types.AccessList{},
		MaxFeePerBlobGas:     big.NewInt(1000000),
		BlobVersionedHashes:  generateRandomHashes(3),
	}

	encodedTx, err := rlp.EncodeToBytes(&txEncode)

	hexEncodedTx2 := hex.EncodeToString(encodedTx)

	fmt.Println(hexEncodedTx2)

	if err != nil {
		fmt.Println("error encoding transactions", err)
		return
	}

	privateKey, err := getECDSAPrivateKey()

	if err != nil {
		fmt.Println("error parsing private key", err)
		return
	}

	generateAndAppendSignatureValues(&tx, &encodedTx, privateKey)

	encodedTxWithSignature, err := rlp.EncodeToBytes(&tx)
	hexEncodedTx := hex.EncodeToString(encodedTxWithSignature)

	fmt.Println(hexEncodedTx)

	if err != nil {
		fmt.Println("Error re-encoding transaction with signature:", err)
		return
	}

	createAndSendTransaction(encodedTxWithSignature)

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

func generateAndAppendSignatureValues(tx *BlobTx, encodedTx *[]byte, privateKey *ecdsa.PrivateKey) {
	signature, err := crypto.Sign(crypto.Keccak256Hash(*encodedTx).Bytes(), privateKey)
	if err != nil {
		fmt.Println("error singing the encoded tx object")
		return
	}

	r := big.NewInt(0).SetBytes(signature[:32])
	s := big.NewInt(0).SetBytes(signature[32:64])
	v := big.NewInt(0).SetBytes([]byte{signature[64] + 27})

	tx.SignatureValues[0] = v.Bytes()
	tx.SignatureValues[1] = r.Bytes()
	tx.SignatureValues[2] = s.Bytes()
}

func generateRandomHashes(n int) [][]byte {
	hashes := make([][]byte, n)
	for i := 0; i < n; i++ {
		hash := make([]byte, 32)
		_, err := rand.Read(hash)
		if err != nil {
			panic("error generating random hash")
		}
		hashes[i] = hash
	}
	return hashes
}

func createAndSendTransaction(encodedTxWithSignature []byte) {
	requestPayload := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_sendRawTransaction",
		Params:  []interface{}{fmt.Sprintf("0x%x", encodedTxWithSignature)},
		ID:      1,
	}

	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		fmt.Println("Error marshaling request:", err)
		return
	}

	req, err := http.NewRequest("POST", os.Getenv("NODE_URL"), bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Printf("Response: %s\n", responseBody)
}
