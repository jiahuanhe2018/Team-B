package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type Block struct {
	Index     int
	Timestamp string
	Data      string
	Hash      string
	PrevHash  string
	Validator string
}

var Blockchain []Block
var tempBlocks []Block

// 创建 block 的一个通道
var candidateBlocks = make(chan Block)

// 创建 字符串 通道
var announcements = make(chan string)

var mutex = &sync.Mutex{}

var validators = make(map[string]int)

func main() {
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), "创世", calculateBlockHash(genesisBlock), "", ""}
	spew.Dump(genesisBlock)

	Blockchain = append(Blockchain, genesisBlock)
	httpPort := "8081"
	server, err := net.Listen("tcp", ":"+httpPort)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("HTTP Server Listening on port :", httpPort)
	defer server.Close()

	go func() {
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks,candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for{

			pickWinner()
		}
	}()
	for  {
		conn, err := server.Accept()
		if   err!=nil{
			log.Fatal(err)
		}
		go handleConn(conn)
	}

}

func pickWinner() {
	time.Sleep(15 * time.Second)
	mutex.Lock()
	blocks := tempBlocks
	mutex.Unlock()

	lotteryPool := []string{}
	if len(blocks) > 0 {
	OUTER:
		for _, block := range blocks {
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}
			mutex.Lock()
			setValidators := validators
			mutex.Unlock()
			k, ok := setValidators[block.Validator]
			if ok {
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}
		}
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)

		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]
		for _, block := range blocks {
			if block.Validator == lotteryWinner {
				mutex.Lock()
				Blockchain = append(Blockchain, block)
				mutex.Unlock()
				for _ = range validators {
					announcements <- "\nwinning validator: " + lotteryWinner + "\n"
				}
				break
			}
		}
	}
	mutex.Lock()
	tempBlocks = []Block{}
	mutex.Unlock()
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()

	var address string
	io.WriteString(conn, "Enter token balance:")
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		i, err := strconv.Atoi(scanner.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanner.Text(), err)
			return
		}
		address = calculateHash(time.Now().String())
		validators[address] = i
		fmt.Println(validators)
		break
	}

	io.WriteString(conn, "\nEnter a new Result:")
	newScanner := bufio.NewScanner(conn)
	go func() {
		for newScanner.Scan() {
			text := newScanner.Text()

			mutex.Lock()
			oldBlock := Blockchain[len(Blockchain)-1]
			mutex.Unlock()

			newBlock, err := generateBlock(oldBlock, text, address)
			if err != nil {
				log.Println(err)
				continue
			}
			if isBlockValid(newBlock, oldBlock) {
				candidateBlocks <- newBlock
			}
			io.WriteString(conn, "\nEnter a new Result:")
		}
	}()
	for {
		time.Sleep(time.Minute)
		mutex.Lock()
		bytes, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(bytes)+"\n")
	}
}

// 校验新区块的合法性
func isBlockValid(newBlock Block, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

// 创建 区块
func generateBlock(oldblock Block, data string, address string) (Block, error) {
	var newBlock Block
	newBlock.Index = oldblock.Index + 1
	newBlock.Data = data
	newBlock.Timestamp = time.Now().String()
	newBlock.PrevHash = oldblock.Hash
	// 计算hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address
	return newBlock, nil

}

// 计算区块 hash
func calculateBlockHash(block Block) string {
	datastr := string(block.Index) + block.Timestamp + block.Data + block.PrevHash
	return calculateHash(datastr)
}

// 计算hash
func calculateHash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	sum := hash.Sum(nil)
	return hex.EncodeToString(sum)
}
