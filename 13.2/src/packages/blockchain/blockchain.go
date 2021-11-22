package blockchain

import (
	"math/big"
	"static-proof-of-stake/src/packages/RSA"
	"static-proof-of-stake/src/packages/ledger"
	"sync"
)

/* Block struct */
//TODO: do we really need all these fields?
type Block struct {
	ID               int                        // ID of the block
	Hash             string                     // Hash of the block
	Epoch            int                        // Epoch of the block
	Creator          string                     // Creator of the block
	PreviousBlock    string                     // Hash of the previous block
	TransactionList  []ledger.SignedTransaction // List of transactions contained in the block
	NextBlocksHashes []string                   // Next blocks hashes
}

/* Signed block struct */
type SignedBlock struct {
	Type      string // Signed Block
	Block     Block  // Block
	Signature string // Block signature
	BlockLock sync.Mutex
}

/** Draw for the lottery **/
type Draw struct {
	Slot      int    // SlotLengthSeconds
	Signature string // Signature of the draw
}

/** Genesis block struct **/
type GenesisBlock struct {
	Hash                 string         // Hash of the genesis block
	Seed                 int            //
	InitialAccountStates map[string]int // Initial account states
	NextBlocksHashes     []string       // Next blocks hashes

}

/* Blockchain struct */
type Blockchain struct {
	BlockList []SignedBlock // List of blocks containing signed transactions
	//Blocks       map[string]Block
	GenesisBlock      GenesisBlock
	Hardness          *big.Int
	Seed              int
	SlotLengthSeconds int
	lock              sync.Mutex
}

/** Block operations **/

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

/** Blockchain operations **/
/* Blockchain adder */
func (blockchain *Blockchain) AddBlock(signedBlock SignedBlock) {
	var localLock sync.Mutex
	localLock.Lock()
	blockchain.BlockList = append(blockchain.BlockList, signedBlock)
	localLock.Unlock()
}

/* Blockchain get head */
func (blockchain *Blockchain) GetHeadID() int {
	var localLock sync.Mutex
	localLock.Lock()
	tempLen := len(blockchain.BlockList)
	localLock.Unlock()
	return tempLen
}
