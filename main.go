package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
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
		To:                   common.HexToAddress("0xReceiverAddress"),
		Value:                big.NewInt(10000000000000000), // 0.01 eth
		Data:                 []byte{},
		AccessList:           types.AccessList{},
		MaxFeePerBlobGas:     big.NewInt(1000000),
		BlobVersionedHashes:  generateRandomHashes(3),
		SignatureValues:      [3][]byte{},
	}

	encodedTx, err := rlp.EncodeToBytes(&tx)

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

	fmt.Printf("Encoded Blob transactions: %x \n", encodedTx)
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
