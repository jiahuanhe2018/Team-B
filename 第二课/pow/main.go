package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 难度值
const difficuty = 1

// 定义区块
type Block struct {
	Index     int
	Timestemp string
	Data    string
	Hash      string
	PrevHash  string
	Difficuty int
	Nonce     string
}

// 区块链定义
var Blockchain []Block

type Message struct {
	Data string
}

var mutex = sync.Mutex{}

func main() {
	//err := godotenv.Load()
	//if err!=nil{
	//	log.Fatal(err)
	//}
	go func() {
		now := time.Now()
		genesisBlock := Block{}
		genesisBlock = Block{0, now.String(), "", calculateHash(genesisBlock), "", difficuty, ""}
		spew.Dump(genesisBlock)

		mutex.Lock()
		Blockchain = append(Blockchain,genesisBlock)
		mutex.Unlock()
	}()
	log.Fatal(run())
}

//web 服务
func run() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")
	httpPort = "8010"
	log.Println("HTTP Server Listening on port :", httpPort)
	server := &http.Server{
		Addr:              ":" + httpPort,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

// web 路由设置
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handelGetBlockChain).Methods("GET")
	muxRouter.HandleFunc("/", handlerWirteBlock).Methods("POST")
	return muxRouter
}

// 获取已经存在的区块链
func handelGetBlockChain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// 创建一个新的块
func handlerWirteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJson(w, r, http.StatusBadRequest, r.Body)
		return
	}
	defer r.Body.Close()
	// 加锁，创建区块
	mutex.Lock()
	newBlock := generateBlock(Blockchain[len(Blockchain)-1], m.Data)
	mutex.Unlock()

	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}
	respondWithJson(w, r, http.StatusCreated, newBlock)

}

// 处理 web  返回信息
func respondWithJson(w http.ResponseWriter, r *http.Request, code int, paylod interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(paylod, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("错误！"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

// 校验新区块的有效性，确保 新 区块的  index、hash 的正确性。
func isBlockValid(newBlock Block, oldBlock Block) bool {
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

// 制造区块
func generateBlock(oldBlock Block, Data string) Block {
	var newBlock Block
	t := time.Now()
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestemp = t.String()
	newBlock.Data = Data
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficuty = difficuty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		if !isHashValid(calculateHash(newBlock), newBlock.Difficuty) {
			fmt.Println(calculateHash(newBlock), "继续工作！")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(calculateHash(newBlock), "完成！")
			newBlock.Hash = calculateHash(newBlock)
			break
		}
	}
	return newBlock
}

// 计算区块hash 值
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestemp + block.Data + block.PrevHash + block.Nonce
	hash := sha256.New()
	hash.Write([]byte(record))
	hashed := hash.Sum(nil)
	return hex.EncodeToString(hashed)
}

// 校验 hash 是否正确
func isHashValid(hash string, difficuty int) bool {
	prefix := strings.Repeat("0", difficuty)
	return strings.HasPrefix(hash, prefix)
}
