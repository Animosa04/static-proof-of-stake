/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 16-10-2021
COURSE: Distributed Systems and Security
DESCRIPTION: Distributed transaction system implemented as structured P2P flooding network.
**/

package peer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"packages/RSA"
	"packages/blockchain"
	"packages/ledger"
	"strconv"
	"sync"
	"time"
)

const MAX_CON = 10
const e = 3

/* Message struct containing list of peers */
type PeersMapMsg struct {
	Type     string
	PeersMap map[string]string // address -> public key map
}

/* Message struct containing address of new peer */
type NewPeerMsg struct {
	Type      string
	Address   string
	PublicKey string
}

/* Peer struct */
type Peer struct {
	outIP            string
	outPort          string
	inIP             string
	inPort           string
	address          string
	broadcast        chan []byte
	ln               net.Listener
	transactionsSeen map[string]bool
	connections      map[string]net.Conn
	ledger           *ledger.Ledger
	lock             sync.Mutex
	peers            PeersMapMsg
	privateKey       string
	publicKey        string

	blockchain           *blockchain.Blockchain
	pendingTransactions  []ledger.SignedTransaction
	transactionsExecuted map[string]bool
}

/* Initialize peer method */
func (peer *Peer) StartPeer() {
	/* User input */
	fmt.Println("Please enter IP to connect to:")
	fmt.Scanln(&peer.outIP)
	fmt.Println("Please enter port to connect to:")
	fmt.Scanln(&peer.outPort)

	/* Initialize variables */
	ln, _ := net.Listen("tcp", "127.0.0.1:")
	ip, port, _ := net.SplitHostPort(ln.Addr().String())
	peer.ln = ln
	peer.inIP = ip
	peer.inPort = port
	peer.address = ip + ":" + port
	peer.broadcast = make(chan []byte)
	peer.transactionsSeen = make(map[string]bool)
	peer.connections = make(map[string]net.Conn, 0)
	peer.ledger = ledger.MakeLedger()

	peer.peers.Type = "peersMap"
	peer.peers.PeersMap = make(map[string]string)

	k := RSA.GenerateRandomK()
	publicKey, privateKey := RSA.KeyGen(k, e)
	peer.privateKey = privateKey.ToString()
	peer.publicKey = publicKey.ToString()

	/* Print address for connectivity */
	peer.printDetails()

	/* Initialize connection and routines */
	peer.connect(peer.outIP + ":" + peer.outPort)
	go peer.write()
	go peer.broadcastMsg()
	go peer.acceptConnect()

	peer.blockchain = blockchain.MakeBlockchain()
	peer.ledger.Accounts[peer.publicKey] = 1000
	peer.transactionsExecuted = make(map[string]bool)
	go peer.playLottery()

}

/* Accept connection method */
func (peer *Peer) connect(address string) {
	/* Check if the peers are already connected */
	for addresses, _ := range peer.connections {
		if addresses == address {
			return
		}
	}
	/* Otherwise, dial the connection */
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Error at peer destination. Connecting to own network...")
		defer peer.connect(peer.inIP + ":" + peer.inPort)
		return
	}
	/* Store the connection for broadcasting */
	peer.connections[address] = conn

	/* Initialize reading routine associated with the conenction */
	go peer.read(conn)
}

/* Accept connect method */
func (peer *Peer) acceptConnect() {
	for {
		/* Accept connection that dials */
		conn, _ := peer.ln.Accept()
		peer.connections[conn.RemoteAddr().String()] = conn
		fmt.Println(peer.address + " got a connection from " + conn.RemoteAddr().String())
		defer peer.ln.Close()

		/* Forward local list of peers */
		jsonString, _ := json.Marshal(peer.peers)
		conn.Write(jsonString)

		/* Start reading input from the connection */
		go peer.read(conn)
	}
}

/* Accept disconnect */
func (peer *Peer) acceptDisconnect(conn net.Conn) {
	/* Locate address and remove it */
	for address, conn := range peer.connections {
		if conn == conn {
			delete(peer.connections, address)
			return
		}
	}
	fmt.Println("Connection not found...")
	return
}

/* Read method of server */
func (peer *Peer) read(conn net.Conn) {
	defer conn.Close()
	/* Decode every message into a string-interface map */
	var temp map[string]interface{}
	decoder := json.NewDecoder(conn)
	for {
		err := decoder.Decode(&temp)
		/* In case of empty string, disconnect the peer */
		if err == io.EOF {
			peer.acceptDisconnect(conn)
			return
		}
		/* In case of an error, crash the peer */
		if err != nil {
			log.Println(err.Error())
			return
		}
		/* Forward the map to the handleRead method */
		peer.handleRead(temp)
	}
}

/* Handle read method */
func (peer *Peer) handleRead(temp map[string]interface{}) {
	/* Reads the type of the object received and activates appropriate switch-statement */
	jsonString, _ := json.Marshal(temp)
	objectType, _ := temp["Type"]
	switch objectType {
	case "peersMap":
		peers := &PeersMapMsg{}
		json.Unmarshal(jsonString, &peers)
		peer.handlePeersMap(*peers)
		return
	case "signedTransaction":
		transaction := &ledger.SignedTransaction{}
		json.Unmarshal(jsonString, &transaction)
		peer.handleSignedTransaction(*transaction)
		return
	case "newPeer":
		newPeer := &NewPeerMsg{}
		json.Unmarshal(jsonString, &newPeer)
		peer.handleNewPeer(*newPeer)
	case "signedBlock":
		signedBlock := &blockchain.SignedBlock{}
		json.Unmarshal(jsonString, &signedBlock)
		peer.handleBlock(*signedBlock)
	default:
		fmt.Println("Error... Type conversion could not be performed...")
		return
	}
}

/* Handle peer map method */
func (peer *Peer) handlePeersMap(peersMap PeersMapMsg) {
	/* If peer already has a map, return */
	if len(peer.peers.PeersMap) != 0 {
		return
	}

	/* Otherwise store the received map */
	peer.peers = peersMap
	for _, publicKey := range peer.peers.PeersMap {
		peer.ledger.Accounts[publicKey] = 1000
	}

	if peer.peers.PeersMap == nil {
		peer.peers.PeersMap = make(map[string]string, 0)
	}

	/* If there are more than 10 peers on list,
	connect to the 10 peers before itself */
	if MAX_CON < len(peer.peers.PeersMap) {
		diff := len(peer.peers.PeersMap) - MAX_CON
		i := 1
		for address, _ := range peer.peers.PeersMap {
			if i >= diff {
				peer.connect(address)
			}
			i++
		}

		/* Otherwise connect to all peers on the map */
	} else {
		for address, _ := range peer.peers.PeersMap {
			peer.connect(address)
		}
	}

	/* Then append itself */
	ownAddress := peer.inIP + ":" + peer.inPort
	peer.peers.PeersMap[ownAddress] = peer.publicKey

	/* As the peer only handles a list of peers, it is new on the network,
	it broadcasts its presence after having connected to the previous 10 peers */
	newPeer := &NewPeerMsg{Type: "newPeer"}
	newPeer.Address = peer.inIP + ":" + peer.inPort
	newPeer.PublicKey = peer.publicKey
	jsonString, _ := json.Marshal(newPeer)
	peer.broadcast <- jsonString
}

/* Handle new peer method */
func (peer *Peer) handleNewPeer(newPeer NewPeerMsg) {
	/* If the peer is not in the local map of peers yet, add it to the map of peers  */
	if _, is_found := peer.peers.PeersMap[newPeer.Address]; !is_found {
		peer.peers.PeersMap[newPeer.Address] = newPeer.PublicKey
		peer.ledger.Accounts[newPeer.PublicKey] = 1000
	}
}

/* Received when a transaction is made */
func (peer *Peer) handleSignedTransaction(signedTransaction ledger.SignedTransaction) {
	valid := RSA.VerifySignature(signedTransaction.Transaction, signedTransaction.Signature, signedTransaction.Transaction.From)

	/* If the transaction signature is valid */
	if valid {
		if signedTransaction.Transaction.Amount < 1 {
			fmt.Println("Invalid transaction. Transaction must send at least 1 AU to be valid.")
			return
		} else if signedTransaction.Transaction.Amount > peer.ledger.Accounts[signedTransaction.Transaction.From] {
			fmt.Println("Invalid transaction. Insufficient funds in the sender's account.")
		}
		/* and if the transaction has not been processed, then */
		if !peer.transactionSeen(signedTransaction) {
			/* add it to the list of transactionsMade and broadcast it */
			peer.markTransactionAsSeen(signedTransaction)
			peer.pendingTransactions = append(peer.pendingTransactions, signedTransaction)
			fmt.Println("Transaction received. Awaing processing...")
			jsonString, _ := json.Marshal(signedTransaction)
			peer.broadcast <- jsonString
		}
		/* If the transaction has been processed, do nothing */
		return
	} else {
		fmt.Println("Signature invalid.")
	}
}

func (peer *Peer) handleBlock(signedBlock blockchain.SignedBlock) {
	signedBlock.BlockLock.Lock()
	defer signedBlock.BlockLock.Unlock()

	// verify block
	senderPublicKey := signedBlock.Block.Vk
	ticketsOfWinner := peer.ledger.Accounts[senderPublicKey]
	valid := blockchain.VerifyWinner(signedBlock.Block.Draw, ticketsOfWinner, peer.blockchain.Hardness, senderPublicKey, peer.blockchain.Seed, signedBlock.Block.Slot)
	if valid {
		// if valid, append block to the blockchain
		fmt.Println("Block was successfully verified.")
		// execute the transactions in the block
		peer.executeTransactions(signedBlock.Block.BlockData)
		// reward the creator of the block
	} else {
		fmt.Println("Block verification failed. Penalizing validator...")
	}
}

/* Write method for client */
func (peer *Peer) write() {
	var i int
	var amount string
	var senderAddress string
	var receiverAddress string
	for {
		/* Read transaction from user */
		fmt.Println("Amount to send: ")
		fmt.Scanln(&amount)
		fmt.Println("Sender's address: ")
		fmt.Scanln(&senderAddress)
		fmt.Println("Receiver's address: ")
		fmt.Scanln(&receiverAddress)

		/* Make transaction object from the details, */
		signedTransaction := &ledger.SignedTransaction{Type: "signedTransaction"}
		signedTransaction.Transaction.ID = senderAddress + strconv.Itoa(i) + strconv.Itoa(rand.Intn(100))
		signedTransaction.Transaction.From = peer.publicKey
		signedTransaction.Transaction.To = peer.peers.PeersMap[receiverAddress]
		signedTransaction.Transaction.Amount, _ = strconv.Atoi(amount)

		/* Generate RSA signature for the transaction using the private key of the sender, */
		signedTransaction.Signature = RSA.GenerateSignature(signedTransaction.Transaction, peer.privateKey)

		/* and broadcast it */
		jsonString, _ := json.Marshal(signedTransaction)
		peer.broadcast <- jsonString
		i++
	}
}

/* Broadcast method */
func (peer *Peer) broadcastMsg() {
	for {
		jsonString := <-peer.broadcast
		for _, con := range peer.connections {
			con.Write(jsonString)
		}
	}
}

/* Print details method */
func (peer *Peer) printDetails() {
	ip, port, _ := net.SplitHostPort(peer.ln.Addr().String())
	fmt.Println("Listening on address " + ip + ":" + port)
	fmt.Println("[" + peer.address + "], publicKey=" + peer.publicKey)
}

func (peer *Peer) printPeersMap() {
	fmt.Println("Peer map:")
	for k, v := range peer.peers.PeersMap {
		fmt.Println("Public key of [" + k + "]:" + v)
	}
}

/* Check if transaction has been seen */
func (peer *Peer) transactionSeen(signedTransaction ledger.SignedTransaction) bool {
	peer.lock.Lock()
	_, seen := peer.transactionsSeen[signedTransaction.Transaction.ID]
	peer.lock.Unlock()
	return seen
}

/* Add transaction method */
func (peer *Peer) markTransactionAsSeen(signedTransaction ledger.SignedTransaction) {
	peer.lock.Lock()
	peer.transactionsSeen[signedTransaction.Transaction.ID] = true
	peer.lock.Unlock()
}

/* Check if transaction has been executed already */
func (peer *Peer) transactionExecuted(signedTransaction ledger.SignedTransaction) bool {
	peer.lock.Lock()
	_, executed := peer.transactionsExecuted[signedTransaction.Transaction.ID]
	peer.lock.Unlock()
	return executed
}

/* Add transaction method */
func (peer *Peer) markTransactionAsExecuted(signedTransaction ledger.SignedTransaction) {
	peer.lock.Lock()
	peer.transactionsExecuted[signedTransaction.Transaction.ID] = true
	peer.lock.Unlock()
}

func (peer *Peer) executeTransactions(transactions []ledger.SignedTransaction) {
	for _, transaction := range transactions {
		if !peer.transactionExecuted(transaction) {
			peer.ledger.ExecuteTransaction(transaction)
			peer.markTransactionAsExecuted(transaction)
		}
	}
	if len(transactions) > 0 {
		fmt.Println("Processed " + strconv.Itoa(len(transactions)) + " transactions")
		defer peer.ledger.PrintLedger()
	}
}

func (peer *Peer) playLottery() {
	for {
		slot := peer.blockchain.GetSlotNumber()
		draw := blockchain.MakeDraw(peer.blockchain.Seed, slot, peer.privateKey)
		tickets := peer.ledger.Accounts[peer.publicKey]
		drawIsWinner := blockchain.IsWinner(draw, tickets, peer.blockchain.Hardness)
		if drawIsWinner {
			// if the peer wins the slot
			fmt.Println("Draw is winner. Making new block...")
			// make a new block with unprocessed transactions
			fmt.Println(strconv.Itoa(len(peer.pendingTransactions)) + " unprocessed transactions found.")
			signedBlock := blockchain.MakeSignedBlock(slot, draw, peer.privateKey, peer.publicKey, peer.pendingTransactions)
			peer.pendingTransactions = nil
			// transmit the new block
			fmt.Println("Transmitting new block ...")
			jsonString, _ := json.Marshal(signedBlock)
			peer.broadcast <- jsonString
		}
		time.Sleep(time.Duration(int64(peer.blockchain.SlotLengthSeconds) * int64(time.Second)))

	}
}
