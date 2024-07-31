package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"
const blocksBucket = "blocks"
const originCoinbaseData = "Random stuff"

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BCIterator struct {
	currHash []byte
	db       *bolt.DB
}

func NewBlockchain(addr string) *Blockchain {
	if !dbExists() {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}
	var tip []byte
	db, _ := bolt.Open(dbFile, 0600, nil)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	return &Blockchain{tip, db}
}

func CreateBlockchain(addr string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}
	var tip []byte
	db, _ := bolt.Open(dbFile, 0600, nil)

	db.Update(func(tx *bolt.Tx) error {
		cbtx := NewCoinbase(addr, originCoinbaseData)
		origin := NewOriginBlock(cbtx)

		b, _ := tx.CreateBucket([]byte(blocksBucket))
		b.Put(origin.Hash, origin.serialize())
		b.Put([]byte("l"), origin.Hash)
		tip = origin.Hash

		return nil
	})
	return &Blockchain{tip, db}
}

func (bc *Blockchain) mineBlock(transactions []*Transaction) {
	var lastHash []byte

	bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	newBlock := NewBlock(transactions, lastHash)

	bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		b.Put(newBlock.Hash, newBlock.serialize())
		b.Put([]byte("l"), newBlock.Hash)
		bc.tip = newBlock.Hash

		return nil
	})
}

func (bc *Blockchain) iterator() *BCIterator {
	return &BCIterator{bc.tip, bc.db}
}

func (i *BCIterator) next() *Block {
	var block *Block

	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	i.currHash = block.PrevBlock
	return block
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (bc *Blockchain) findUnspentTransactions(addr string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.iterator()

	for {
		block := bci.next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIDx, out := range tx.OUTPUTS {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIDx {
							continue Outputs
						}
					}
				}
				if out.canBeUnlockedWith(addr) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if !tx.isCoinbase() {
				for _, in := range tx.INPUTS {
					if in.canUnlockOutputWith(addr) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.OUTPUT)
					}
				}
			}
		}
		if len(block.PrevBlock) == 0 {
			break
		}
	}
	return unspentTXs
}

func (bc *Blockchain) findUTXO(addr string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.findUnspentTransactions(addr)

	for _, tx := range unspentTransactions {
		for _, out := range tx.OUTPUTS {
			if out.canBeUnlockedWith(addr) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain) findSpendableOutputs(addr string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.findUnspentTransactions(addr)
	acc := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outID, out := range tx.OUTPUTS {
			if out.canBeUnlockedWith(addr) && acc < amount {
				acc += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outID)

				if acc >= amount {
					break Work
				}
			}
		}
	}
	return acc, unspentOutputs
}
