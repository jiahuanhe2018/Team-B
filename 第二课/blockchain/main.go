package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	"github.com/multiformats/go-multiaddr"
	"io"
	"log"
	rand2 "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Block struct {
	Index     int
	Data      string
	Timestamp string
	Hash      string
	PrevHash  string
}

var Blockchain []Block

var mutex = &sync.Mutex{}

func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {
	var r io.Reader

	if randseed == 0 {
		r = rand.Reader
	} else {
		r = rand2.New(rand2.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.

	privKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}
	opt := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(privKey),
	}
	basicHost, err := libp2p.New(context.Background(), opt...)
	if err != nil {
		return nil, err
	}

	newMultiaddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(newMultiaddr)
	log.Printf("I am %s\n", fullAddr)

	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}

func handleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], sendData)

		if IsBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
			mutex.Lock()
			Blockchain = append(Blockchain, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(Blockchain)
		if err != nil {
			log.Println(err)
		}

		spew.Dump(Blockchain)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}

}

//func main() {
//	now := time.Now()
//
//	genesisBlock := Block{}
//
//	//genesisBlock = Block{0, "", now.String(), calculateHash(genesisBlock), ""}
//
//	Blockchain = append(Blockchain, genesisBlock)
//
//	log2.SetAllLoggers(logging.INFO)
//
//	// Parse options from the command line
//	listenF := flag.Int("l", 0, "wait for incoming connections")
//	target := flag.String("d", "", "target peer to dial")
//	secio := flag.Bool("secio", false, "enable secio")
//	seed := flag.Int64("seed", 0, "set random seed for id generation")
//	flag.Parse()
//
//	if *listenF == 0 {
//		log.Fatal("Please provide a port to bind on with -l")
//	}
//
//	basicHost, err := makeBasicHost(*listenF, *secio, *seed)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if *target == "" {
//		log.Println("listening for connections")
//
//		// Set a stream handler on host A. /p2p/1.0.0 is
//		// a user-defined protocol name.
//		basicHost.SetStreamHandler("/p2p/1.0.0", handleStream)
//
//		select {} // hang forever
//
//		/**** This is where the listener code ends ****/
//
//	} else {
//		basicHost.SetStreamHandler("/p2p/1.0.0", handleStream)
//		// The following code extracts target's peer ID from the
//		// given multiaddress
//
//		ipfsaddr, err := multiaddr.NewMultiaddr(*target)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		pid, err := ipfsaddr.ValueForProtocol(multiaddr.P_IPFS)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		peerid, err := peer.IDB58Decode(pid)
//		if err != nil {
//			log.Fatalln(err)
//		}
//		// Decapsulate the /ipfs/<peerID> part from the target
//		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
//
//		targetPeerAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
//		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
//		// We have a peer ID and a targetAddr so we add it to the peerstore
//		// so LibP2P knows how to contact it
//		basicHost.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
//
//		log.Println("opening stream")
//
//		// make a new stream from host B to host A
//		// it should be handled on host A by the handler we set above because
//		// we use the same /p2p/1.0.0 protocol
//		s, err := basicHost.NewStream(context.Background(), peerid, "/p2p/1.0.0")
//		if err != nil {
//			log.Fatalln(err)
//		}
//		// Create a buffered stream so that read and writes are non blocking.
//		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
//		// Create a thread to read and write data.
//		go writeData(rw)
//		go readData(rw)
//
//		select {} // hang forever
//	}
//}

// make sure block is valid by checking index, and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Data + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, Result string) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Data = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock
}
