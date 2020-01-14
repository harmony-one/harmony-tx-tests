package testing

import (
	"fmt"
	"sync"

	"github.com/SebastianJ/harmony-tx-tests/accounts"
	"github.com/SebastianJ/harmony-tx-tests/balances"
	"github.com/SebastianJ/harmony-tx-tests/config"
	"github.com/SebastianJ/harmony-tx-tests/funding"
	"github.com/SebastianJ/harmony-tx-tests/transactions"
)

// RunMultipleSenderTestCase - runs a tests where multiple sender wallets are used to send to one respective new wallet
func RunMultipleSenderTestCase(testCase TestCase) {
	Title(testCase.Name, "header", testCase.Verbose)

	nameTemplate := fmt.Sprintf("TestCase_%s_Sender_", testCase.Name)
	senderAccounts, _ := funding.GenerateAndFundAccounts(testCase.Parameters.SenderCount, nameTemplate, testCase.Parameters.FromShardID, testCase.Parameters.ToShardID, testCase.Parameters.Amount)
	receiverAccount := accounts.GenerateTypedAccount(fmt.Sprintf("TestCase_%s_Receiver", testCase.Name))

	responses := make(chan bool, testCase.Parameters.SenderCount)
	var waitGroup sync.WaitGroup

	for _, senderAccount := range senderAccounts {
		waitGroup.Add(1)
		go performSingleAccountTest(testCase, senderAccount, receiverAccount, responses, &waitGroup)
	}

	waitGroup.Wait()
	close(responses)

	successfulCount := 0
	for response := range responses {
		if response {
			successfulCount++
		}
	}

	receiverEndingBalance, _ := balances.GetShardBalance(receiverAccount.Address, testCase.Parameters.ToShardID)
	Log(testCase.Name, fmt.Sprintf("Receiver account %s (address: %s) has an ending balance of %f in shard %d after the test", receiverAccount.Name, receiverAccount.Address, receiverEndingBalance, testCase.Parameters.ToShardID), testCase.Verbose)

	// && ((receiverStartingBalance)+testCase.Parameters.Amount == receiverEndingBalance))
	txsSuccessful := (successfulCount == testCase.Parameters.SenderCount)
	testCase.Result = (txsSuccessful && receiverEndingBalance == (float64(testCase.Parameters.SenderCount)*testCase.Parameters.Amount))

	Teardown(receiverAccount.Name, receiverAccount.Address, testCase.Parameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.Parameters.ToShardID, testCase.Parameters.Amount, testCase.Parameters.GasPrice, 0)

	Title(testCase.Name, "footer", testCase.Verbose)

	Results = append(Results, testCase)
}

func performSingleAccountTest(testCase TestCase, senderAccount accounts.Account, receiverAccount accounts.Account, responses chan<- bool, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	senderStartingBalance, _ := balances.GetShardBalance(senderAccount.Address, testCase.Parameters.FromShardID)

	Log(testCase.Name, fmt.Sprintf("Generated a new receiver account: %s, address: %s", receiverAccount.Name, receiverAccount.Address), testCase.Verbose)
	Log(testCase.Name, fmt.Sprintf("Using sender account %s (address: %s) and receiver account %s (address : %s)", senderAccount.Name, senderAccount.Address, receiverAccount.Name, receiverAccount.Address), testCase.Verbose)
	Log(testCase.Name, fmt.Sprintf("Sender account %s (address: %s) has a starting balance of %f in shard %d before the test", senderAccount.Name, senderAccount.Address, senderStartingBalance, testCase.Parameters.FromShardID), testCase.Verbose)
	Log(testCase.Name, fmt.Sprintf("Will let the transaction wait up to %d seconds to try to get finalized within 2 blocks", testCase.Parameters.ConfirmationWaitTime), testCase.Verbose)
	Log(testCase.Name, "Sending transaction...", testCase.Verbose)

	rawTx, err := transactions.SendTransaction(senderAccount.Address, testCase.Parameters.FromShardID, receiverAccount.Address, testCase.Parameters.ToShardID, testCase.Parameters.Amount, -1, testCase.Parameters.GasPrice, testCase.Parameters.Data, testCase.Parameters.ConfirmationWaitTime)
	testCaseTx := ConvertToTestCaseTransaction(senderAccount.Address, receiverAccount.Address, rawTx, testCase.Parameters, err)
	testCase.Transactions = append(testCase.Transactions, testCaseTx)

	Log(testCase.Name, fmt.Sprintf("Sent %f token(s) from %s to %s - transaction hash: %s, tx successful: %t", testCase.Parameters.Amount, config.Configuration.Funding.Account.Address, receiverAccount.Address, testCaseTx.TransactionHash, testCaseTx.Success), testCase.Verbose)

	senderEndingBalance, _ := balances.GetShardBalance(senderAccount.Address, testCase.Parameters.FromShardID)

	Log(testCase.Name, fmt.Sprintf("Sender account %s (address: %s) has an ending balance of %f in shard %d after the test", senderAccount.Name, senderAccount.Address, senderEndingBalance, testCase.Parameters.FromShardID), testCase.Verbose)
	Log(testCase.Name, "Performing test teardown (returning funds and removing sender account)", testCase.Verbose)

	Teardown(senderAccount.Name, senderAccount.Address, testCase.Parameters.FromShardID, config.Configuration.Funding.Account.Address, testCase.Parameters.ToShardID, testCase.Parameters.Amount, testCase.Parameters.GasPrice, 0)

	responses <- testCaseTx.Success
}