package blockchain

import (
	"after_feedback/src/packages/RSA"
	"after_feedback/src/packages/ledger"
	"sync"
)

/* Block struct */
type SignedBlock struct {
	Type      string // Signed Block
	Signature string // Block signature
	Block     Block  // Block
	BlockLock sync.Mutex
}

/* Block struct */
type Block struct {
	ID              int                        // ID of the block
	TransactionList []ledger.SignedTransaction // List of transactions contained in the block
}

/* Block adder method */
func (signedBlock *SignedBlock) AddTransaction(transaction ledger.SignedTransaction) {
	signedBlock.BlockLock.Lock()
	signedBlock.Block.TransactionList = append(signedBlock.Block.TransactionList, transaction)
	defer signedBlock.BlockLock.Unlock()
}

/* Sign block method */
func (signedBlock *SignedBlock) SignBlock(privateKey string) {
	signedBlock.BlockLock.Lock()
	signedBlock.Signature = RSA.GenerateSignature(signedBlock.Block, privateKey)
	defer signedBlock.BlockLock.Unlock()
}

func (signedBlock *SignedBlock) GetLength() int {
	signedBlock.BlockLock.Lock()
	tempLen := len(signedBlock.Block.TransactionList)
	defer signedBlock.BlockLock.Unlock()
	return tempLen
}

/* Next block method */
func (signedBlock *SignedBlock) NextBlock() {
	signedBlock.BlockLock.Lock()
	signedBlock.Block.TransactionList = nil
	signedBlock.Block.ID += 1
	defer signedBlock.BlockLock.Unlock()
}

/* Blockchain struct */
type Blockchain struct {
	blockList []SignedBlock // List of blocks containing signed transactions
}

/* Blockchain adder */
func (blockchain *Blockchain) AddBlock(signedBlock SignedBlock) {
	var localLock sync.Mutex
	localLock.Lock()
	blockchain.blockList = append(blockchain.blockList, signedBlock)
	localLock.Unlock()
}

/* Blockchain get head */
func (blockchain *Blockchain) GetHeadID() int {
	var localLock sync.Mutex
	localLock.Lock()
	tempLen := len(blockchain.blockList)
	localLock.Unlock()
	return tempLen
}
