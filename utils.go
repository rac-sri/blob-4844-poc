package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

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

func createBlobTx(chainID *big.Int, nonce uint64, tip *big.Int, maxFeePerGas *uint256.Int, blob kzg4844.Blob, input []byte) (*types.Transaction, error) {
	blobCommit, _ := kzg4844.BlobToCommitment(blob)
	blobProof, _ := kzg4844.ComputeBlobProof(blob, blobCommit)
	sidecar := types.BlobTxSidecar{
		Blobs:       []kzg4844.Blob{blob},
		Commitments: []kzg4844.Commitment{blobCommit},
		Proofs:      []kzg4844.Proof{blobProof},
	}

	return types.NewTx(&types.BlobTx{
		ChainID:    uint256.MustFromBig(chainID),
		Nonce:      nonce,
		GasTipCap:  uint256.MustFromBig(tip),
		GasFeeCap:  maxFeePerGas,
		Gas:        250000,
		To:         common.HexToAddress(TO_ADDRESS),
		Value:      uint256.NewInt(0),
		Data:       input,
		BlobFeeCap: uint256.NewInt(1e10),
		BlobHashes: sidecar.BlobHashes(),
		Sidecar:    &sidecar,
	}), nil
}

func encodeBlob(data []byte) kzg4844.Blob {
	blob := kzg4844.Blob{}
	fieldIndex := 0

	for i := 0; i < len(data); i += 31 {
		if fieldIndex >= params.BlobTxFieldElementsPerBlob {
			panic("Data exceeds the capacity")
		}

		max := i + 31
		if max > len(data) {
			max = len(data)
		}

		copy(blob[fieldIndex*32:], data[i:max])
		fieldIndex++
	}
	return blob
}
