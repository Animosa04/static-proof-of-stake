package blockchain

import (
	"math/big"
	"packages/ledger"
	"sync"
)

/* Lottery draw struct */
type Draw struct {
	Lottery string // "lottery"
	Seed    int    // Seed
	Slot    int    // Slot number
}

/* Block struct */
type Block struct {
	Type              string                     // block
	Vk                string                     // Verification key of the block (vk), signifies the creator of the block
	Slot              int                        // Slot number (block number)
	Draw              string                     // Draw that was used to win the lottery
	BlockData         []ledger.SignedTransaction // List of transactions contained in the block (U)
	Hash              string                     //	Hash of the block
	PreviousBlockHash string                     // Hash of the previous block (h)
	//NextBlocksHashes  []string                   // Hashes of the next blocks
	BlockLock sync.Mutex
}

/* Signed block struct */
type SignedBlock struct {
	Type      string // signedBlock
	Block     *Block // Block
	Signature string // Block signature (sigma)
	BlockLock sync.Mutex
}

/* Blockchain struct */
type Blockchain struct {
	BlocksMap         map[string]Block // List of blocks containing signed transactions
	GenesisBlock      Block            // Genesis block of the blockchain
	Seed              int
	Hardness          *big.Int
	SlotLengthSeconds int
	blockchainLock    sync.Mutex
}
