package main

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

func hashData(data []byte) []byte {
	return crypto.Keccak256(data)
}

func merkleTreeRoot(values []*big.Int) []byte {
	var nodes [][]byte

	for _, value := range values {
		hash := hashData(value.Bytes())
		nodes = append(nodes, hash)
	}

	for len(nodes) > 1 {
		var tempNodes [][]byte
		for i := 0; i < len(nodes); i += 2 {
			if i+1 < len(nodes) {
				combinedHash := hashData(append(nodes[i], nodes[i+1]...))
				tempNodes = append(tempNodes, combinedHash)
			} else {

				tempNodes = append(tempNodes, nodes[i])
			}
		}
		nodes = tempNodes
	}

	return nodes[0]
}
