package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
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

	// SuggestGasTipCap returns a suggested tip cap.
	suggestedTip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error suggesting gas tip cap: %v", err)
	}

	// Fetch the latest block to get the base fee
	header, err := client.HeaderByNumber(context.Background(), nil) // nil for latest block
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error fetching latest block header: %v", err)
	}

	// Calculate maxFeePerGas as the sum of the base fee from the latest block plus a margin (e.g., the suggested tip).
	// This provides a buffer ensuring the transaction can cover the base fee plus provides a priority fee (miner tip).
	maxFeePerGas := new(big.Int).Add(header.BaseFee, suggestedTip)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return 0, nil, nil, nil, fmt.Errorf("error fetching chain id: %v", err)
	}

	return nonce, chainID, suggestedTip, uint256.MustFromBig(maxFeePerGas), nil
}

func createBlobTx(chainID *big.Int, nonce uint64, tip *big.Int, maxFeePerGas *uint256.Int, root []byte, input []byte) (*types.Transaction, error) {
	blobs, commits, proofs, versionedHashed, err := EncodeBlobs(root)

	if err != nil {
		log.Fatal(err)
	}
	sidecar := types.BlobTxSidecar{
		Blobs:       blobs,
		Commitments: commits,
		Proofs:      proofs,
	}
	fmt.Println("working")
	return types.NewTx(&types.BlobTx{
		ChainID:    uint256.MustFromBig(chainID),
		Nonce:      nonce,
		GasTipCap:  uint256.MustFromBig(tip),
		GasFeeCap:  maxFeePerGas,
		Gas:        2500000,
		To:         common.HexToAddress(TO_ADDRESS),
		Value:      uint256.NewInt(0),
		Data:       input,
		BlobFeeCap: uint256.NewInt(1e7 * 786432),
		BlobHashes: versionedHashed,
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

func EncodeBlobs(data []byte) ([]kzg4844.Blob, []kzg4844.Commitment, []kzg4844.Proof, []common.Hash, error) {
	var (
		blobs           = encodeBlobs(data)
		commits         []kzg4844.Commitment
		proofs          []kzg4844.Proof
		versionedHashes []common.Hash
	)

	for _, blob := range blobs {
		fmt.Println(len(blob))
		commit, err := kzg4844.BlobToCommitment(blob)
		fmt.Println("fsdklnfs")
		fmt.Println(err)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		commits = append(commits, commit)

		proof, err := kzg4844.ComputeBlobProof(blob, commit)

		if err != nil {
			return nil, nil, nil, nil, err
		}
		proofs = append(proofs, proof)

		versionedHashes = append(versionedHashes, kZGToVersionedHash(commit))

	}
	return blobs, commits, proofs, versionedHashes, nil
}

var blobCommitmentVersionKZG uint8 = 0x01

// kZGToVersionedHash implements kzg_to_versioned_hash from EIP-4844
func kZGToVersionedHash(kzg kzg4844.Commitment) common.Hash {
	h := sha256.Sum256(kzg[:])
	h[0] = blobCommitmentVersionKZG

	return h
}
