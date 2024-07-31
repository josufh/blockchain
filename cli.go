package main

import (
	"flag"
	"fmt"
	"os"
)

type CLI struct{}

func (cli *CLI) run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createbc", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddr := getBalanceCmd.String("addr", "", "")
	createBlockchainAdrr := createBlockchainCmd.String("addr", "", "")
	sendFrom := sendCmd.String("from", "", "")
	sendTo := sendCmd.String("to", "", "")
	sendAmount := sendCmd.Int("amount", 0, "")

	switch os.Args[1] {
	case "getbalance":
		getBalanceCmd.Parse(os.Args[2:])
	case "createblockchain":
		createBlockchainCmd.Parse(os.Args[2:])
	case "printchain":
		printChainCmd.Parse(os.Args[2:])
	case "send":
		sendCmd.Parse(os.Args[2:])
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		cli.getBalance(*getBalanceAddr)
	}
	if createBlockchainCmd.Parsed() {
		cli.createBlockchain(*createBlockchainAdrr)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
	if sendCmd.Parsed() {
		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// ToDo: improve message
func (cli *CLI) printUsage() {
	fmt.Println("Not correct usage.")
}

func (cli *CLI) getBalance(addr string) {
	bc := NewBlockchain(addr)
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.findUTXO(addr)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", addr, balance)
}

func (cli *CLI) createBlockchain(addr string) {
	bc := CreateBlockchain(addr)
	bc.db.Close()
	fmt.Println("Done.")
}

func (cli *CLI) printChain() {
	fmt.Println("Printing...")
}

func (cli *CLI) send(from, to string, amount int) {
	bc := NewBlockchain(from)
	defer bc.db.Close()

	tx := NewTransaction(from, to, amount, bc)
	bc.mineBlock([]*Transaction{tx})
	fmt.Println("Sent.")
}
