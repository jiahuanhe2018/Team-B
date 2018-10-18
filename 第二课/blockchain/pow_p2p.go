package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	log2 "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-net"
	"github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/multiformats/go-multiaddr"
	"github.com/whyrusleeping/go-logging"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var difficuty = 1


type BlockPow struct {
	Index     int
	Data      string
	Timestamp string
	Hash      string
	PrevHash  string
	Nonce     string
	Difficuty int
}

var BlockchainPow []BlockPow


func main() {
	now := time.Now()

	genesisBlock := BlockPow{}

	genesisBlock = BlockPow{0, "", now.String(), calculateHashPow(genesisBlock), "","",difficuty}

	BlockchainPow = append(BlockchainPow, genesisBlock)

	log2.SetAllLoggers(logging.INFO)

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	basicHost, err := makeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")

		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.
		basicHost.SetStreamHandler("/p2p/1.0.0", handleStreamPow)

		select {} // hang forever

		/**** This is where the listener code ends ****/

	} else {
		basicHost.SetStreamHandler("/p2p/1.0.0", handleStreamPow)
		// The following code extracts target's peer ID from the
		// given multiaddress

		ipfsaddr, err := multiaddr.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}
		pid, err := ipfsaddr.ValueForProtocol(multiaddr.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}
		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}
		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>

		targetPeerAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		basicHost.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)

		log.Println("opening stream")

		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := basicHost.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
		// Create a thread to read and write data.
		go writeDataPow(rw)
		go readDataPow(rw)

		select {} // hang forever
	}
}

func handleStreamPow(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readDataPow(rw)
	go writeDataPow(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}



// SHA256 hashing
func calculateHashPow(block BlockPow) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Data + block.PrevHash+block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func readDataPow(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]BlockPow, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(BlockchainPow) {
				BlockchainPow = chain
				bytes, err := json.MarshalIndent(BlockchainPow, "", "  ")
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

func writeDataPow(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainPow)
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
		newBlock := generateBlockPow(BlockchainPow[len(BlockchainPow)-1], sendData)

		if IsBlockValidPow(newBlock, BlockchainPow[len(BlockchainPow)-1]) {
			mutex.Lock()
			BlockchainPow = append(BlockchainPow, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(BlockchainPow)
		if err != nil {
			log.Println(err)
		}

		spew.Dump(BlockchainPow)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}

}


func IsBlockValidPow(newBlock, oldBlock BlockPow) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHashPow(newBlock) != newBlock.Hash {
		return false
	}

	return true
}


// create a new block using previous block's hash
func generateBlockPow(oldBlock BlockPow, Result string) BlockPow {

	var newBlock BlockPow

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Data = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHashPow(newBlock)
	newBlock.Difficuty = difficuty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHashPow(newBlock), newBlock.Difficuty) {
			fmt.Println(calculateHashPow(newBlock), "继续工作！")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHashPow(newBlock), "完成！")
			newBlock.Hash = calculateHashPow(newBlock)
			break
		}
	}

	return newBlock
}


// 校验 hash 是否正确
func isHashValid(hash string, difficuty int) bool {
	prefix := strings.Repeat("0", difficuty)
	return strings.HasPrefix(hash, prefix)
}

