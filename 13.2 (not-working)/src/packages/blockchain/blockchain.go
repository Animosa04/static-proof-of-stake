package blockchain

import (
	"math/big"
	"packages/RSA"
	"time"
)

const SEED = 3
const SLOT_LENGTH_SECONDS = 2

func MakeDraw(seed int, slot int, sk string) string {
	draw := new(Draw)
	draw.Lottery = "lottery"
	draw.Seed = seed
	draw.Slot = slot
	return RSA.GenerateSignature(draw, sk)
}

func VerifyDraw(drawToVerify string, vk string, seed int, slot int) bool {
	draw := new(Draw)
	draw.Lottery = "lottery"
	draw.Seed = seed
	draw.Slot = slot
	return RSA.VerifySignature(draw, drawToVerify, vk)
}

func MakeGenesisBlock(vk string, draw string) *Block {
	genesisBlock := new(Block)
	genesisBlock.Type = "block"
	genesisBlock.Vk = vk
	genesisBlock.BlockData = nil
	genesisBlock.PreviousBlockHash = ""
	genesisBlock.Draw = draw
	//genesisBlock.Signature = genesisBlock.MakeSignature()
	return genesisBlock
}

func (block *Block) GetHash() string {
	/* Hash object with SHA-256 and get integer representation of hash, */
	blockHash := RSA.ByteArrayToInt(RSA.ComputeHash(block))
	return blockHash.String()
}

func MakeBlockchain() *Blockchain {
	blockchain := new(Blockchain)
	blockchain.Seed = SEED
	blockchain.Hardness = big.NewInt(1)
	blockchain.SlotLengthSeconds = SLOT_LENGTH_SECONDS
	return blockchain
}

func (blockchain *Blockchain) GetSlotNumber() int {
	return int(time.Now().Unix() / int64(blockchain.SlotLengthSeconds))
}
