package ledger

import (
	"fmt"
	"strconv"
	"sync"
)

const TRANSACTION_FEE = 1

/* Transaction struct */
type Transaction struct {
	ID     string // ID of the transaction
	From   string // Sender of the transaction (public key)
	To     string // Receiver of the transaction (public key)
	Amount int    // Amount to transfer
}

/* Signed transaction struct */
type SignedTransaction struct {
	Type        string      // Signed transaction
	Transaction Transaction //Transaction object of a signed transaction
	Signature   string      // Signature of the transaction
}

/* Ledger struct */
type Ledger struct {
	Type       string
	Accounts   map[string]int
	LedgerLock sync.Mutex
}

/* Ledger constructor */
func MakeLedger() *Ledger {
	ledger := new(Ledger)
	ledger.Accounts = make(map[string]int)
	return ledger
}

/* Transaction method */
func (ledger *Ledger) Transaction(signedTransaction SignedTransaction) {
	ledger.LedgerLock.Lock()
	ledger.Accounts[signedTransaction.Transaction.From] -= signedTransaction.Transaction.Amount
	ledger.Accounts[signedTransaction.Transaction.To] += (signedTransaction.Transaction.Amount - TRANSACTION_FEE)
	defer ledger.LedgerLock.Unlock()
}

/* Print ledger method */
func (ledger *Ledger) PrintLedger() {
	ledger.LedgerLock.Lock()
	for account, amount := range ledger.Accounts {
		fmt.Println("Account name: " + account + " amount: " + strconv.Itoa(amount) + " AU")
	}
	defer ledger.LedgerLock.Unlock()
}
