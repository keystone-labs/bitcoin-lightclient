package main

import (
	"bytes"
	"crypto/sha256"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type Hash256Digest [32]byte

// We copy logic from bitcoin-spv. The main reason is bitcoin-spv is not maintain anymore.
// https://github.com/summa-tx/bitcoin-spv/
// Thank summa-tx for their awesome work

// Hash256MerkleStep concatenates and hashes two inputs for merkle proving
func Hash256MerkleStep(a, b []byte) Hash256Digest {
	c := []byte{}
	c = append(c, a...)
	c = append(c, b...)
	return DoubleHash(c)
}

func DoubleHash(in []byte) Hash256Digest {
	first := sha256.Sum256(in)
	second := sha256.Sum256(first[:])
	return Hash256Digest(second)
}

// follow logic on bitcoin-spv.
// This is check the tx belong to merkle tree hash in BTC header.
func VerifyHash256Merkle(proof []byte, index uint) bool {
	var current Hash256Digest
	idx := index
	proofLength := len(proof)

	if proofLength%32 != 0 {
		return false
	}
	if proofLength == 32 {
		return true
	}
	if proofLength == 64 {
		return false
	}

	root := proof[proofLength-32:]
	cur := proof[:32:32]
	copy(current[:], cur)
	numSteps := (proofLength / 32) - 1

	for i := 1; i < numSteps; i++ {
		start := i * 32
		end := i*32 + 32
		next := proof[start:end:end]
		if idx%2 == 1 {
			current = Hash256MerkleStep(next, current[:])
		} else {
			current = Hash256MerkleStep(current[:], next)
		}
		idx >>= 1
	}

	return bytes.Equal(current[:], root)
}

// verify UTXO (identified as tx and UTXO index) is in a blocks' Merkle root
func (lc *BTCLightClient) VerifyUTXO(tx *btcutil.Tx, utxoIdx uint, merkleRoot *chainhash.Hash, merklePath []byte) bool {
	txHash := tx.Hash()
	proof := []byte{}
	proof = append(proof, txHash[:]...)
	proof = append(proof, merklePath...)
	proof = append(proof, merkleRoot[:]...)

	return VerifyHash256Merkle(proof, utxoIdx)
}

func (lc *BTCLightClient) VerifyTXID(txHash *chainhash.Hash, utxoIdx uint, header *wire.BlockHeader, merklePath []byte) bool {
	proof := []byte{}
	proof = append(proof, txHash[:]...)
	proof = append(proof, merklePath...)
	merkleRoot := header.MerkleRoot
	proof = append(proof, merkleRoot[:]...)
	return VerifyHash256Merkle(proof, utxoIdx)
}
