package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

const subsidy = 20

type Transaction struct {
	ID      []byte
	INPUTS  []TXInput
	OUTPUTS []TXOutput
}

type TXOutput struct {
	Value     int
	ScriptPub string
}

type TXInput struct {
	ID        []byte
	OUTPUT    int
	ScriptSig string
}

func (tx Transaction) isCoinbase() bool {
	return len(tx.INPUTS) == 1 && len(tx.INPUTS[0].ID) == 0 && tx.INPUTS[0].OUTPUT == -1
}

func (tx *Transaction) setID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	enc.Encode(tx)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

func (in *TXInput) canUnlockOutputWith(data string) bool {
	return in.ScriptSig == data
}

func (out *TXOutput) canBeUnlockedWith(data string) bool {
	return out.ScriptPub == data
}

func NewTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	acc, validOutputs := bc.findSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("err NOT ENOUGH FUNDS")
		os.Exit(1)
	}

	for txid, outs := range validOutputs {
		txID, _ := hex.DecodeString(txid)
		for _, out := range outs {
			input := TXInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.setID()

	return &tx
}

func NewCoinbase(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	input := TXInput{[]byte{}, -1, data}
	output := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{input}, []TXOutput{output}}
	tx.setID()

	return &tx
}
