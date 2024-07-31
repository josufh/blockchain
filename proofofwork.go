package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const zeroBits = 20

var maxFiller = math.MaxInt64

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-zeroBits))

	return &ProofOfWork{b, target}
}

func (pow *ProofOfWork) prepareData(filler int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlock, pow.block.hashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(zeroBits)),
			IntToHex(int64(filler)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	filler := 0

	fmt.Println("Mining new block")

	for filler < maxFiller {
		data := pow.prepareData(filler)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			filler++
		}
	}
	fmt.Printf("\n\n")
	return filler, hash[:]
}

func (pow *ProofOfWork) validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Filler)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
