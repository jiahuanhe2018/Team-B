package blockchain

import (
	"strings"
	"bufio"
	"fmt"
	"encoding/json"
	"time"
	"log"
	"os"
	"strconv"
		)


var difficultly = 1
//ReadDataByPow
func ReadDataByPow(rw *bufio.ReadWriter) {
	fmt.Printf("0000000000000000\n")
	for {
		fmt.Printf("1111111111111111111\n")
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("22222222222222222222 = %v\n",str)
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
			}else{

				fmt.Printf("44444444444\n")
			}
			mutex.Unlock()
		}else {

			fmt.Printf("str= n\n")
		}
	}
}

func WriteDataByPow(rw *bufio.ReadWriter) {
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
		fmt.Print("\nEnter a new Result: ")
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
		newBlock := GenerateBlockByPow(BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1], _result)

		if len(BlockchainInstance.TxPool.AllTx) > 0 {
			BlockchainInstance.PackageTx(&newBlock)
		}else {
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


// create a new block using previous block's hash
func GenerateBlockByPow(oldBlock Block, Result int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)
	newBlock.Difficultly = difficultly

	for i:=0;;i++{

		hex := fmt.Sprintf("%x",i)

		newBlock.Nonce = hex

		if !isHashVaild(CalculateHash(newBlock), newBlock.Difficultly){
			fmt.Println(CalculateHash(newBlock),"do more work")
			time.Sleep(time.Second)
			continue
		}else{
			fmt.Println(CalculateHash(newBlock), "work done")

			newBlock.Hash = CalculateHash(newBlock)
			break
		}
	}

	return newBlock
}
