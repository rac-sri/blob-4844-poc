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
	header, err := client.HeaderByNumber(context.Background(), nil) // nil for latest block
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error fetching latest block header: %v", err)
	}

	maxFeePerGas := new(big.Int).Add(header.BaseFee, suggestedTip)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error fetching chain id: %v", err)
	}

	return nonce, chainID, suggestedTip, uint256.MustFromBig(maxFeePerGas), nil
}

func createBlobTx(chainID *big.Int, nonce uint64, tip *big.Int, maxFeePerGas *uint256.Int, root []byte, input []byte) (*types.Transaction, error) {
	// blobs, commits, proofs, err := EncodeBlobs(root)

	// if err != nil {
	// 	log.Fatal(err)
	// }
	emptyBlob := kzg4844.Blob{}
	emptyBlobCommit, err := kzg4844.BlobToCommitment(emptyBlob)
	if err != nil {
		log.Fatal("Failed to create commitment", "err", err)
	}
	emptyBlobProof, err := kzg4844.ComputeBlobProof(emptyBlob, emptyBlobCommit)
	if err != nil {
		log.Fatal("Failed to create proof", "err", err)
	}

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
		Gas:        2500000,
		To:         common.HexToAddress(TO_ADDRESS),
		Value:      uint256.NewInt(0),
		Data:       input,
		BlobFeeCap: uint256.NewInt(3e10),
		BlobHashes: sidecar.BlobHashes(),
		Sidecar:    &sidecar,
	}), nil
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

func EncodeBlobs(data []byte) ([]kzg4844.Blob, []kzg4844.Commitment, []kzg4844.Proof, error) {
	var (
		blobs   []kzg4844.Blob
		commits []kzg4844.Commitment
		proofs  []kzg4844.Proof
	)

	for _, blob := range blobs {

		commit, err := kzg4844.BlobToCommitment(blob)

		if err != nil {
			return nil, nil, nil, err
		}

		commits = append(commits, commit)

		proof, err := kzg4844.ComputeBlobProof(blob, commit)

		if err != nil {
			return nil, nil, nil, err
		}
		proofs = append(proofs, proof)

	}
	return blobs, commits, proofs, nil
}
