package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp    int64
	Transactions []*Transaction
	PrevBlock    []byte
	Hash         []byte
	Filler       int
}

func NewBlock(transactions []*Transaction, prevBlock []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlock, []byte{}, 0}

	pow := NewProofOfWork(block)
	filler, hash := pow.run()

	block.Hash = hash[:]
	block.Filler = filler
	return block
}

func NewOriginBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}

func (b *Block) hashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}
