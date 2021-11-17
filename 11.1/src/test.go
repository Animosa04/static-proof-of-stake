/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 18-09-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed transaction system implemented as structured P2P flooding network.
**/

package main

import (
	"after_feedback/src/packages/peer"
	"fmt"
)

func main() {
	var from1 string
	var to1 string
	//var from2 string
	//var to2 string
	var p = peer.Peer{}
	p.StartTestPeer()
	fmt.Println("Please enter FROM1 port: ")
	fmt.Scanln(&from1)
	fmt.Println("Please enter TO1 port: ")
	fmt.Scanln(&to1)
	//fmt.Println("Please enter FROM2 port: ")
	//fmt.Scanln(&from2)
	//fmt.Println("Please enter TO2 port: ")
	//fmt.Scanln(&to2)
	go p.ConcurrencyTest(from1, to1)
	//go p.ConcurrencyTest(from2, to2)
	for true {
	}
}
