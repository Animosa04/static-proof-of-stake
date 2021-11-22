package peer

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"static-proof-of-stake/src/packages/RSA"
	"static-proof-of-stake/src/packages/blockchain"
	"static-proof-of-stake/src/packages/ledger"
	"strconv"
	"sync"
)

const MAX_CON = 10
const e = 3
const SEED = 42
const slot = 0
const LOTTERY = "lottery"

/* Message struct containing list of peers */
// TODO: mutex here?
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
// TODO:should the peer have a mutex
type Peer struct {
	outIP            string              // Outbound IP address
	outPort          string              // Outbound port
	inIP             string              // Inbound IP address
	InPort           string              // Inbound port
	address          string              // Address (IP:port)
	broadcast        chan []byte         // Channel for broadcasting messages
	ln               net.Listener        // Listener for incoming messages
	transactionsSeen map[string]bool     // Transactions seen
	transactionsLock sync.Mutex          //Lock for list of seen transactions
	connections      map[string]net.Conn // Connections to other peers
	ledger           *ledger.Ledger      // Peer ledger
	peers            PeersMapMsg         // Map of peers' addresses and public keys
	PrivateKey       string              // Private key of the peer
	PublicKey        string              // Public key of the peer
	blockID          int
	currBlock        blockchain.SignedBlock
	blockchain       blockchain.Blockchain
}

/* Initialize peer method */
func (peer *Peer) StartPeer(params ...string) { // params = {type: ["manual", "test"], port (optional)}
	if len(params) == 1 {
		if params[0] == "manual" {
			/* User input */
			/* fmt.Println("Please enter IP to connect to:")
			fmt.Scanln(&peer.outIP) */
			peer.outIP = "127.0.0.1"
			fmt.Println("Please enter port to connect to:")
			fmt.Scanln(&peer.outPort)
		}
		if params[0] == "test" {
			peer.outPort = "9007"
		}
	} else if len(params) == 2 {
		peer.outPort = params[1]
	}

	/* Initialize variables */
	ln, _ := net.Listen("tcp", "127.0.0.1:")
	ip, port, _ := net.SplitHostPort(ln.Addr().String())
	peer.ln = ln
	peer.inIP = ip
	peer.InPort = port
	peer.address = ip + ":" + port
	peer.broadcast = make(chan []byte, 10)
	peer.transactionsSeen = make(map[string]bool)
	peer.connections = make(map[string]net.Conn, 0)
	peer.ledger = ledger.MakeLedger()

	peer.peers.Type = "peersMap"
	peer.peers.PeersMap = make(map[string]string)
	/*
		peer.currBlock.Type = "signedBlock"
		peer.currBlock.Block.ID = 0
		peer.blockID = 0 */

	k := RSA.GenerateRandomK()
	publicKey, privateKey := RSA.KeyGen(k, e)
	peer.PrivateKey = privateKey.ToString()
	peer.PublicKey = publicKey.ToString()

	/* Print address for connectivity */
	peer.printDetails()
	fmt.Println("[" + peer.address + "], publicKey=" + peer.PublicKey)

	/* Initialize connection and routines */
	peer.connect(peer.outIP + ":" + peer.outPort)
	go peer.write()
	go peer.broadcastMsg()
	go peer.acceptConnect()
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
		defer peer.connect(peer.inIP + ":" + peer.InPort)
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
	// TODO: what?
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
		//go peer.handleRead(temp)
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
		peer.HandleSignedTransaction(*transaction)
		return
	case "newPeer":
		newPeer := &NewPeerMsg{}
		json.Unmarshal(jsonString, &newPeer)
		peer.handleNewPeer(*newPeer)
		return
	case "signedBlock":
		signedBlock := &blockchain.SignedBlock{}
		json.Unmarshal(jsonString, &signedBlock)
		peer.handleSignedBlock(*signedBlock)
		return
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
	ownAddress := peer.inIP + ":" + peer.InPort
	peer.peers.PeersMap[ownAddress] = peer.PublicKey

	/* As the peer only handles a list of peers, it is new on the network,
	it broadcasts its presence after having connected to the previous 10 peers */
	newPeer := &NewPeerMsg{Type: "newPeer"}
	newPeer.Address = peer.inIP + ":" + peer.InPort
	newPeer.PublicKey = peer.PublicKey
	jsonString, _ := json.Marshal(newPeer)
	peer.broadcast <- jsonString
}

/* Handle new peer method */
func (peer *Peer) handleNewPeer(newPeer NewPeerMsg) {
	/* If the peer is not in the local map of peers yet, add it to the map of peers  */
	if _, is_found := peer.peers.PeersMap[newPeer.Address]; !is_found {
		peer.peers.PeersMap[newPeer.Address] = newPeer.PublicKey
	}
}

/* Received when a transaction is made */
func (peer *Peer) HandleSignedTransaction(signedTransaction ledger.SignedTransaction) {
	valid := RSA.VerifySignature(signedTransaction.Transaction, signedTransaction.Signature, signedTransaction.Transaction.From)
	/* If the transaction signature is valid */
	if valid {
		if signedTransaction.Transaction.Amount < 0 {
			fmt.Println("Amount cannot be negative...")
			return
		}
		/* if peer.isSequencer {
			if peer.locateTransaction(signedTransaction) == false {
				fmt.Println("Adding a trans. to currBlock...")
				peer.currBlock.AddTransaction(signedTransaction)
			}
		} */
		/* and if the transaction has not been seen, then */
		if !peer.locateTransaction(signedTransaction) {
			/* add it to the list of transactionsMade and broadcast it */
			fmt.Println("New transaction received...")
			peer.addTransaction(signedTransaction)
			peer.ledger.Transaction(signedTransaction)
			defer peer.ledger.PrintLedger()
			jsonString, _ := json.Marshal(signedTransaction)
			peer.broadcast <- jsonString
		}
		/* If the transaction has been processed, do nothing */
		return
	} else {
		fmt.Println("Signature invalid.")
	}
}

// TODO: validate transaction here
func (peer *Peer) handleSignedBlock(signedBlock blockchain.SignedBlock) {
	valid := RSA.VerifySignature(signedBlock.Block, signedBlock.Signature, peer.PublicKey)
	if valid {
		if signedBlock.Block.ID == peer.blockID {
			for _, currTransaction := range signedBlock.Block.TransactionList {
				peer.ledger.Transaction(currTransaction)
				//TODO: NEGATIVE TRANSACTIONS WILL BE IGNORED HERE
				//tempVal := peer.ledger.Accounts[currTransaction.Transaction.From] - currTransaction.Transaction.Amount
				//if tempVal <= 0 {
				//	peer.ledger.Transaction(currTransaction)
				//} else {
				//	fmt.Println("Sender account has insufficient funds...")
				//}
			}
			peer.blockID += 1
			peer.ledger.PrintLedger()
		}
	} else {
		fmt.Println("Invalid signature.")
	}
}

/* Write method for client */
func (peer *Peer) write() {
	var i int
	var amount int
	var senderAddress string
	var senderAccountBalance int
	var receiverAddress string
	for {
		/* Read transaction from user */
		fmt.Println("Amount to send: ")
		fmt.Scanln(&amount)
		fmt.Println("Sender's address: ")
		fmt.Scanln(&senderAddress)
		fmt.Println("Receiver's address: ")
		fmt.Scanln(&receiverAddress)

		senderAccountBalance, _ = strconv.Atoi(peer.peers.PeersMap[senderAddress])

		if amount < 1 {
			fmt.Println("Invalid transaction. Transaction must send at least 1 AU to be valid. Please try again.")
		} else if senderAccountBalance-amount < 0 {
			fmt.Println("Invalid transaction. Insufficient AUs. Available AUs: " + strconv.Itoa(senderAccountBalance) + ". Please try again.")
		} else {
			/* Make transaction object from the details, */
			signedTransaction := &ledger.SignedTransaction{Type: "signedTransaction"}
			signedTransaction.Transaction.ID = senderAddress + strconv.Itoa(i) + strconv.Itoa(rand.Intn(100))
			signedTransaction.Transaction.From = peer.PublicKey
			signedTransaction.Transaction.To = peer.peers.PeersMap[receiverAddress]
			signedTransaction.Transaction.Amount = amount

			/* Generate RSA signature for the transaction using the private key of the sender, */
			signedTransaction.Signature = RSA.GenerateSignature(signedTransaction.Transaction, peer.PrivateKey)

			/* and broadcast it */
			jsonString, _ := json.Marshal(signedTransaction)
			peer.broadcast <- jsonString
			i++
		}
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
}

func (peer *Peer) printPeersMap() {
	fmt.Println("Peer map:")
	for k, v := range peer.peers.PeersMap {
		fmt.Println("Public key of [" + k + "]:" + v)
	}
}

//TODO: why is transaction executed twice
/* Locate transaction method */
func (peer *Peer) locateTransaction(signedTransaction ledger.SignedTransaction) bool {
	peer.transactionsLock.Lock()
	_, found := peer.transactionsSeen[signedTransaction.Transaction.ID]
	peer.transactionsLock.Unlock()
	return found
}

/* Add transaction method */
func (peer *Peer) addTransaction(signedTransaction ledger.SignedTransaction) {
	peer.transactionsLock.Lock()
	peer.transactionsSeen[signedTransaction.Transaction.ID] = true
	peer.transactionsLock.Unlock()
}
