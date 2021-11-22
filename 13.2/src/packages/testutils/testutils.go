package testutils

import (
	"math/rand"
	"static-proof-of-stake/src/packages/RSA"
	"static-proof-of-stake/src/packages/ledger"
	"strconv"
)

func MakeMockSignedTransaction(senderPrivateKey string, senderPublicKey string, receiverPublicKey string) ledger.SignedTransaction {
	var mockTransaction = ledger.Transaction{}
	mockTransaction.ID = strconv.Itoa(rand.Intn(100))
	mockTransaction.From = senderPublicKey
	mockTransaction.To = receiverPublicKey
	mockTransaction.Amount = 15

	var mockSignedTransaction = ledger.SignedTransaction{}
	mockSignedTransaction.Type = "signedTransaction"
	mockSignedTransaction.Transaction = mockTransaction
	mockSignedTransaction.Signature = RSA.GenerateSignature(mockTransaction, senderPrivateKey)

	return mockSignedTransaction
}
