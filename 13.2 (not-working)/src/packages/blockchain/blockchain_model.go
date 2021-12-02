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
	Vk                string                     // Verification key of the block (vk)
	Slot              int                        // Slot number (block number)
	Draw              string                     // Draw that was used to win the lottery
	BlockData         []ledger.SignedTransaction // List of transactions contained in the block (U)
	PreviousBlockHash string                     // Hash of the previous block (h)
	Signature         string                     // Block signature (sigma)
	BlockLock         sync.Mutex
}

/* Signed block struct */
type SignedBlock struct {
	Type      string // signed block
	Block     Block  // Block
	Signature string // Block signature
	BlockLock sync.Mutex
}

/** Genesis block struct **/
type GenesisBlock struct {
	Type                 string                     // block
	ID                   int                        // ID of the block
	Hash                 string                     // Hash of the genesis block
	PreviousBlock        string                     // Hash of the previous block
	Creator              string                     // Creator of the block
	BlockData            []ledger.SignedTransaction // List of transactions contained in the block
	Signature            string                     // Block signature
	BlockLock            sync.Mutex
	Seed                 int            // Seed of the blockchain
	InitialAccountStates map[string]int // Initial account states
}

/* Blockchain struct */
type Blockchain struct {
	BlockList         []Block // List of blocks containing signed transactions
	GenesisBlock      GenesisBlock
	Seed              int
	Hardness          *big.Int
	SlotLengthSeconds int
	blockchainLock    sync.Mutex
}
