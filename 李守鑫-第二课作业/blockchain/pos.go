package blockchain

import (
	"time"
	"fmt"
	"encoding/json"
	"log"
	"bufio"
	"os"
	"strings"
	"strconv"
	"encoding/hex"
	"crypto/sha256"
	"math/rand"
)

// validators keeps track of open validators and balances
var Validators = make(map[string]int)
//tmpblock
var TempBlocks []Block

var announcements = make(chan []Block)

func pickWinner(rw *bufio.ReadWriter) {
	time.Sleep(10 * time.Second)
	mutex.Lock()
	temp := TempBlocks
	mutex.Unlock()
	fmt.Print("pickWinner timer1111111111111111111111111111111,temp = \n", temp)
	lotteryPool := []string{}
	if len(temp) > 0 {
		fmt.Print("tmp = 0\n")
		// slightly modified traditional proof of stake algorithm
		// from all validators who submitted a block, weight them by the number of staked tokens
		// in traditional proof of stake, validators can participate without submitting a block to be forged
	OUTER:
		for _, block := range temp {
			// if already in lottery pool, skip
			for _, node := range lotteryPool {
				if block.Validator == node {
					continue OUTER
				}
			}

			// lock list of validators to prevent data race
			mutex.Lock()
			setValidators := Validators
			mutex.Unlock()

			k, ok := setValidators[block.Validator]
			if ok {
				for i := 0; i < k; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}


		}
		fmt.Print("lotteryPool = \n", len(lotteryPool))
		// randomly pick winner from lottery pool
		s := rand.NewSource(time.Now().Unix())
		r := rand.New(s)
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

		// add block of winner to blockchain and let all the other nodes know
		for _, block := range temp {
			if block.Validator == lotteryWinner {

				tmpBlockChain := BlockchainInstance.Blocks
				mutex.Lock()
				tmpBlockChain = append(tmpBlockChain, block)
				mutex.Unlock()

				fmt.Printf("\x1b[32m pickwinner success---%v\x1b[0m\n", lotteryWinner)
				//announcements <- tmpBlockChain
				bytes, err := json.Marshal(tmpBlockChain)
				if err != nil {
					log.Println(err)
				}
				rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
				rw.Flush()
				break

			}
		}
		mutex.Lock()
		TempBlocks = []Block{}
		mutex.Unlock()
	}
}

//ReadDataByPos
func ReadDataByPos(rw *bufio.ReadWriter) {
	fmt.Printf("0000000000000000\n")

	go func() {
		for{
			pickWinner(rw)
		}

	}()

	go func() {
		msg := <-announcements
		fmt.Printf("update_msg,len=%d,  block= \n %v\n",len(msg),msg)
		bytes, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
		}
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
	}()

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

			inputBlock := Block{}
			block := make([]Block, 0)
			validator := make(map[string]int)

			if err := json.Unmarshal([]byte(str), &validator); err != nil {
				fmt.Printf("Unmarshal validator = bad\n")
				//log.Println(err)
			} else {

				if len(validator) > len(Validators) {
					mutex.Lock()
					Validators = validator
					fmt.Printf("new_validator =%v \n", validator)
					mutex.Unlock()
				}
				continue
			}

			if err := json.Unmarshal([]byte(str), &inputBlock); err != nil {
				fmt.Printf("Unmarshal inputBlock = bad\n")
				//log.Println(err)
			} else {

				mutex.Lock()
				TempBlocks = append(TempBlocks, inputBlock)
				mutex.Unlock()
				fmt.Printf("Unmarshal block,tmpBlock = %v\n", TempBlocks)
				continue
			}

			if err := json.Unmarshal([]byte(str), &block); err != nil {
				fmt.Printf("Unmarshal block = bad\n")
				//log.Println(err)
			} else {
				fmt.Printf("33333333333333333333,,new_block = %d block_len=%d\n",len(block),len(BlockchainInstance.Blocks))
				fmt.Printf("33333333333333333333,,new_block = \n%v \n",block)
				mutex.Lock()
				if len(block) > len(BlockchainInstance.Blocks) {
					BlockchainInstance.Blocks = block
					bytes, err := json.MarshalIndent(BlockchainInstance.Blocks, "", "  ")
					if err != nil {

						log.Fatal(err)
					}
					fmt.Printf("MarshalIndent =\n")
					//mutex.Lock()
					TempBlocks = []Block{}
					//mutex.Unlock()
					// Green console color: 	\x1b[32m
					// Reset console color: 	\x1b[0m
					fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
				} else {

					fmt.Printf("44444444444\n")
				}
				mutex.Unlock()
				continue
			}
		}
	}
}

func WriteDataByPos(rw *bufio.ReadWriter) {
	fmt.Printf("WriteData_000000000000000000000000\n")

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Printf("WriteData_1111111111111111111\n")
			//mutex.Lock()
			bytes, err := json.Marshal(BlockchainInstance.Blocks)
			if err != nil {
				log.Println(err)
			}
			//mutex.Unlock()
			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()
	//go func() {
	//	for {
	//		time.Sleep(5 * time.Second)
	//		mutex.Lock()
	//		bytes, err := json.Marshal(Validators)
	//		if err != nil {
	//			log.Println(err)
	//		}
	//		mutex.Unlock()
	//
	//		mutex.Lock()
	//		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	//		rw.Flush()
	//		mutex.Unlock()
	//
	//	}
	//}()
	stdReader := bufio.NewReader(os.Stdin)

	fmt.Print("\nEnter token balance: ")
	balance, err := stdReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	balance = strings.Replace(balance, "\n", "", -1)
	_result, err := strconv.Atoi(balance)
	if err != nil {
		log.Fatal(err)
	}

	mutex.Lock()
	t := time.Now()
	var address string
	address = CalculateHashByPos(t.String())
	Validators[address] = _result

	bytes, err := json.Marshal(Validators)
	if err != nil {
		fmt.Print("\nerrorororororoorororo ")
		log.Println(err)
	}

	rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	//rw.Flush()
	mutex.Unlock()

	for {
		//	tempBlock := make([]Block, 0)

		fmt.Print("\nEnter a new Result:")
		result, err := stdReader.ReadString('\n')
		if err != nil {
			delete(Validators, address)
		}
		result = strings.Replace(result, "\n", "", -1)
		iResult, err := strconv.Atoi(result)
		mutex.Lock()
		fmt.Print("\n             33333333333333-len = ", len(BlockchainInstance.Blocks))
		oldLastBlockIndex := BlockchainInstance.Blocks[len(BlockchainInstance.Blocks)-1]
		mutex.Unlock()

		newBlock := GenerateBlockByPos(oldLastBlockIndex, iResult, address)
		//tempBlock = append(tempBlock, newBlock)
		fmt.Print("\n    newBlock.result = \n", newBlock.Result)
		bytes, err := json.Marshal(newBlock)
		if err != nil {
			log.Println(err)
		}
		fmt.Print("\n             33333333333333-newBlock = ", newBlock)
		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}
}

// create a new block using previous block's hash
func GenerateBlockByPos(oldBlock Block, Result int, str string) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Result = Result
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = CalculateHash(newBlock)
	newBlock.Difficultly = difficultly
	newBlock.Validator = str
	return newBlock
}

func CalculateHashByPos(str string) string {
	record := str
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}
