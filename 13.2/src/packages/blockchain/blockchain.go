package blockchain

import (
	"math/big"
	"packages/RSA"
	"packages/ledger"
	"time"
)

const SEED = 3
const SLOT_LENGTH_SECONDS = 3

// TODO: use draw as struct Draw instead of string
func MakeDraw(seed int, slot int, sk string) string {
	draw := new(Draw)
	draw.Lottery = "lottery"
	draw.Seed = seed
	draw.Slot = slot
	return RSA.GenerateSignature(draw, sk)
}

func IsWinner(draw string, tickets int, hardness *big.Int) bool {
	drawHash := RSA.ByteArrayToInt(RSA.ComputeHash(draw))
	ticketsBigInt := big.NewInt(int64(tickets)) //TODO: make tickets into big Int
	drawValue := big.NewInt(0).Mul(drawHash, ticketsBigInt)
	return drawValue.Cmp(hardness) == 0 || drawValue.Cmp(hardness) == 1
}

func VerifyWinner(drawToVerify string, tickets int, hardness *big.Int, vk string, seed int, slot int) bool {
	draw := new(Draw)
	draw.Lottery = "lottery"
	draw.Seed = seed
	draw.Slot = slot
	if !RSA.VerifySignature(draw, drawToVerify, vk) {
		return false
	}
	return IsWinner(drawToVerify, tickets, hardness)
}

func MakeSignedBlock(slot int, draw string, sk string, vk string, transactions []ledger.SignedTransaction) *SignedBlock {
	block := new(Block)
	block.Vk = vk
	block.Slot = slot
	block.Draw = draw
	block.BlockData = transactions
	block.Hash = RSA.ByteArrayToInt(RSA.ComputeHash(block)).String()
	//block.NextBlocksHashes = make([]string, 0, 1)
	signedBlock := new(SignedBlock)
	signedBlock.Type = "signedBlock"
	signedBlock.Block = block
	signedBlock.Signature = RSA.GenerateSignature(signedBlock.Block, sk)
	return signedBlock
}

func MakeGenesisBlock(slot int, draw string, sk string, vk string) *SignedBlock {
	block := new(Block)
	block.Vk = vk
	block.Slot = slot
	block.Draw = draw
	block.Hash = RSA.ByteArrayToInt(RSA.ComputeHash(block)).String()
	block.PreviousBlockHash = ""
	//block.NextBlocksHashes = make([]string, 0, 1)
	signedBlock := new(SignedBlock)
	signedBlock.Type = "signedBlock"
	signedBlock.Block = block
	signedBlock.Signature = RSA.GenerateSignature(signedBlock.Block, sk)
	return signedBlock
}

func MakeBlockchain() *Blockchain {
	blockchain := new(Blockchain)
	blockchain.BlocksMap = make(map[string]Block)
	blockchain.Seed = SEED
	blockchain.Hardness = new(big.Int)
	blockchain.Hardness, _ = blockchain.Hardness.SetString("98101277522421650198781678972208785932907589725093492146067428082680095847419000000", 10)
	blockchain.SlotLengthSeconds = SLOT_LENGTH_SECONDS
	return blockchain
}

/* Append block to blockchain */
func (blockchain *Blockchain) AppendBlock(block *Block) {
	blockchain.blockchainLock.Lock()
	defer blockchain.blockchainLock.Unlock()

	// if the received block does not have previous block hash, it means it is the genesis block
	if block.PreviousBlockHash == "" {
		// so set the genesis block for the blockchain
		blockchain.GenesisBlock = *block
	} else {
		// if it is a regular block, add block to the blockchain
		blockchain.BlocksMap[block.Hash] = *block
		// and update the previous block to point to the new block
		//previousBlock := blockchain.BlocksMap[block.PreviousBlockHash]
		//previousBlock.NextBlocksHashes = append(previousBlock.NextBlocksHashes, block.Hash)
	}
}

func (blockchain *Blockchain) GetSlotNumber() int {
	return int(time.Now().Unix() / int64(blockchain.SlotLengthSeconds))
}

/* func (blockchain *Blockchain) GetLongestChainLeaf() (int, string) {
	return blockchain.GenesisBlock.GetLongestChainLeaf()
} */
