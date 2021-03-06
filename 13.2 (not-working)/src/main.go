package main

import (
	"fmt"
	"packages/peer"
)

func main() {
	testMakeGenesisBlock()
}

func manual() {
	var p = peer.Peer{}
	p.StartPeer("manual")
	for {
	}
}

func test() {
	var peer1 = peer.Peer{}
	peer1.StartPeer("test")
	peer1_port := peer1.InPort
	fmt.Println(peer1_port)
	var peer2 = peer.Peer{}
	peer2.StartPeer("test", peer1_port)
	//mockSignedTransaction := testutils.MakeMockSignedTransaction(peer1.PrivateKey, peer1.PublicKey, peer2.PublicKey)
	//peer2.HandleSignedTransaction(mockSignedTransaction)
}

func testMakeGenesisBlock() {
	var peer1 = peer.Peer{}
	peer1.StartPeer("test")
	peer1.PlayLottery()
}
