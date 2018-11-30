package blockchain

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	//"github.com/davecgh/go-spew/spew"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	ma "github.com/multiformats/go-multiaddr"
)

var WalletSuffix string

var ConsensusType string

// Block represents each 'item' in the blockchain
type Block struct {
	Index        int               `json:"index"`
	Timestamp    string            `json:"timestamp"`
	Result       int               `json:"result"`
	Hash         string            `json:"hash"`
	PrevHash     string            `json:"prevhash"`
	Proof        uint64            `json:"proof"`
	Transactions []Transaction     `json:"transactions"`
	Accounts     map[string]uint64 `json:"accounts"`
	Difficultly  int               `json:"difficultly"`
	Nonce        string            `json:"nonce"`
	Validator    string            `json:"validator"`
}

type Transaction struct {
	Amount    uint64 `json:"amount"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Data      []byte `json:"data"`
}

type TxPool struct {
	AllTx []Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		AllTx: make([]Transaction, 0),
	}
}

func (p *TxPool) Clear() bool {
	if len(p.AllTx) == 0 {
		return true
	}
	p.AllTx = make([]Transaction, 0)
	return true
}

// Blockchain is a series of validated Blocks
type Blockchain struct {
	Blocks []Block
	TxPool *TxPool
}

func (t *Blockchain) NewTransaction(sender string, recipient string, amount uint64, data []byte) *Transaction {
	transaction := new(Transaction)
	transaction.Sender = sender
	transaction.Recipient = recipient
	transaction.Amount = amount
	transaction.Data = data

	return transaction
}

func (t *Blockchain) AddTxPool(tx *Transaction) int {
	t.TxPool.AllTx = append(t.TxPool.AllTx, *tx)
	return len(t.TxPool.AllTx)
}

func (t *Blockchain) LastBlock() Block {
	return t.Blocks[len(t.Blocks)-1]
}

func (t *Blockchain) GetBalance(address string) uint64 {
	accounts := t.LastBlock().Accounts
	if value, ok := accounts[address]; ok {
		return value
	}
	return 0
}

func (t *Blockchain) PackageTx(newBlock *Block) {
	(*newBlock).Transactions = t.TxPool.AllTx
	AccountsMap := t.LastBlock().Accounts
	for k1, v1 := range AccountsMap {
		fmt.Println(k1, "--", v1)
	}

	unusedTx := make([]Transaction, 0)

	for _, v := range t.TxPool.AllTx {
		if value, ok := AccountsMap[v.Sender]; ok {
			if value < v.Amount {
				unusedTx = append(unusedTx, v)
				continue
			}
			AccountsMap[v.Sender] = value - v.Amount
		}

		if value, ok := AccountsMap[v.Recipient]; ok {
			AccountsMap[v.Recipient] = value + v.Amount
		} else {
			AccountsMap[v.Recipient] = v.Amount
		}
	}

	t.TxPool.Clear()
	//余额不够的交易放回交易池
	if len(unusedTx) > 0 {
		for _, v := range unusedTx {
			t.AddTxPool(&v)
		}
	}

	(*newBlock).Accounts = AccountsMap
}

var BlockchainInstance Blockchain = Blockchain{
	TxPool: NewTxPool(),
}

var mutex = &sync.Mutex{}

func Lock() {
	mutex.Lock()
}

func UnLock() {
	mutex.Unlock()
}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func MakeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addr := basicHost.Addrs()[0]
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s, ConsensusType = %s\n", fullAddr, ConsensusType)
	if secio {
		log.Printf("Now run \"go run main.go -c chain -t %v -l %d -d %s -secio\" on a different terminal\n", ConsensusType, listenPort+2, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -c chain -t %v -l %d -d %s\" on a different terminal\n", ConsensusType, listenPort+2, fullAddr)
	}

	return basicHost, nil
}

func HandleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	if "pow" == ConsensusType {
		go ReadDataByPow(rw)
		go WriteDataByPow(rw)
	} else if "pos" == ConsensusType {
		go ReadDataByPos(rw)
		go WriteDataByPos(rw)
	}

	// stream 's' will stay open until you close it (or the other side closes it).
}

func ReadData(rw *bufio.ReadWriter) {
	fmt.Printf("0000000000000000\n")
	for {
		fmt.Printf("1111111111111111111\n")
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("22222222222222222222 = %v\n", str)
		if str == "" {
			fmt.Printf("str= null\n")
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("33333333333333333333\n")
			mutex.Lock()
			if len(chain) > len(BlockchainInstance.Blocks) {
				BlockchainInstance.Blocks = chain
				bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			} else {

				fmt.Printf("44444444444\n")
			}
			mutex.Unlock()
		} else {

			fmt.Printf("str= n\n")
		}
	}
}

func WriteData(rw *bufio.ReadWriter) {
	fmt.Printf("WriteData_000000000000000000000000\n")
	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Printf("WriteData_1111111111111111111\n")
			mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
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
	fmt.Printf("WriteData_222222222222222222222\n")
	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("WriteData_33333333333333333\n")
		sendData = strings.Replace(sendData, "\n", "", -1)
		_result, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := GenerateBlock(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		} else {
			newBlock.Accounts = BlockchainInstance.LastBlock().Accounts
		}

		if IsBlockValid(newBlock, BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]) {
			mutex.Lock()
			BlockchainInstance.Blocks = append(BlockchainInstance.Blocks, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(BlockchainInstance.Blocks)
		if err != nil {
			log.Println(err)
		}

		//spew.Dump(BlockchainInstance.Blocks)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()

	}

}

// make sure block is valid by checking index, and comparing the hash of the previous block
func IsBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if CalculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// create a new block using previous block's hash
func GenerateBlock(oldBlock Block, Result int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)
	newBlock.Difficultly = difficultly

	return newBlock
}

func isHashVaild(hash string, difficultly int) bool {
	prefix := strings.Repeat("0", difficultly)

	return strings.HasPrefix(hash, prefix)
}

// SHA256 hashing
func CalculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.Result) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

