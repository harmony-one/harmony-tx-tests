package testcases

import (
	"fmt"

	"github.com/SebastianJ/harmony-tx-tests/accounts"
	"github.com/SebastianJ/harmony-tx-tests/balances"
	"github.com/SebastianJ/harmony-tx-tests/testing"
	"github.com/SebastianJ/harmony-tx-tests/transactions"
	"github.com/SebastianJ/harmony-tx-tests/utils"
)

// Common test parameters are defined here - e.g. the test case name, expected result of the test and the required parameters to run the test case
var sbs5TestCase testing.TestCase = testing.TestCase{
	Scenario: "Same Beacon Shard",
	Name:     "SBS5",
	Goal:     "Insufficient amount",
	Priority: 0,
	Expected: true,
	Verbose:  true,
	Parameters: testing.TestCaseParameters{
		FromShardID:          0,
		ToShardID:            0,
		Data:                 "",
		Amount:               1.00E+20,
		GasPrice:             0,
		Count:                1,
		ConfirmationWaitTime: 16,
	},
}

// Sbs5TestCase - Same Beacon Shard single account transfer A -> B, Shard 0 -> 0, Amount 1.00E+20, Tx Data nil, expects: unsuccessful token transfer from A to B within 2 blocks time 16s
func Sbs5TestCase(accs map[string]string, passphrase string, node string) testing.TestCase {
	keyName, fromAddress := utils.RandomItemFromMap(accs)

	testing.Title(sbs5TestCase.Name, "header", sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Using source/sender key: %s and address: %s", keyName, fromAddress), sbs5TestCase.Verbose)

	sinkAccountName := fmt.Sprintf("%s_sink", keyName)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Generating a new receiver/sink account: %s", sinkAccountName), sbs5TestCase.Verbose)
	toAddress, err := accounts.GenerateAccountAndReturnAddress(sinkAccountName, passphrase)

	senderStartingBalance, _ := balances.GetShardBalance(fromAddress, sbs5TestCase.Parameters.FromShardID, node)
	receiverStartingBalance, _ := balances.GetShardBalance(toAddress, sbs5TestCase.Parameters.ToShardID, node)

	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Generated a new receiver/sink account: %s, address: %s", sinkAccountName, toAddress), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Using source account %s (address: %s) and sink account %s (address : %s)", keyName, fromAddress, sinkAccountName, toAddress), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Source account %s (address: %s) has a starting balance of %f in shard %d before the test", keyName, fromAddress, senderStartingBalance, sbs5TestCase.Parameters.FromShardID), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Sink account %s (address: %s) has a starting balance of %f in shard %d before the test", sinkAccountName, toAddress, receiverStartingBalance, sbs5TestCase.Parameters.ToShardID), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Will let the transaction wait up to %d seconds to try to get finalized within 2 blocks", sbs5TestCase.Parameters.ConfirmationWaitTime), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, "Sending transaction...", sbs5TestCase.Verbose)

	rawTx, err := transactions.SendTransaction(fromAddress, sbs5TestCase.Parameters.FromShardID, toAddress, sbs5TestCase.Parameters.ToShardID, sbs5TestCase.Parameters.Amount, sbs5TestCase.Parameters.GasPrice, sbs5TestCase.Parameters.Data, passphrase, node, sbs5TestCase.Parameters.ConfirmationWaitTime)
	testCaseTx := testing.ConvertToTestCaseTransaction(fromAddress, toAddress, rawTx, sbs5TestCase.Parameters, err)
	sbs5TestCase.Transactions = append(sbs5TestCase.Transactions, testCaseTx)

	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Sent %f token(s) from %s to %s - transaction hash: %s, tx successful: %t", sbs5TestCase.Parameters.Amount, fromAddress, toAddress, testCaseTx.TransactionHash, testCaseTx.Success), sbs5TestCase.Verbose)

	senderEndingBalance, _ := balances.GetShardBalance(fromAddress, sbs5TestCase.Parameters.FromShardID, node)
	receiverEndingBalance, _ := balances.GetShardBalance(toAddress, sbs5TestCase.Parameters.ToShardID, node)

	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Source account %s (address: %s) has an ending balance of %f in shard %d after the test", keyName, fromAddress, senderEndingBalance, sbs5TestCase.Parameters.FromShardID), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, fmt.Sprintf("Sink account %s (address: %s) has an ending balance of %f in shard %d after the test", sinkAccountName, toAddress, receiverEndingBalance, sbs5TestCase.Parameters.ToShardID), sbs5TestCase.Verbose)
	testing.Log(sbs5TestCase.Name, "Performing test teardown (returning funds and removing sink account)", sbs5TestCase.Verbose)
	testing.Title(sbs5TestCase.Name, "footer", sbs5TestCase.Verbose)

	testing.Teardown(sinkAccountName, toAddress, sbs5TestCase.Parameters.FromShardID, fromAddress, sbs5TestCase.Parameters.ToShardID, sbs5TestCase.Parameters.Amount, sbs5TestCase.Parameters.GasPrice, passphrase, node, 0)

	sbs5TestCase.Result = (!testCaseTx.Success && receiverEndingBalance == 0.0)

	return sbs5TestCase
}