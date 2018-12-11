package main 

import(
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

// Block item
type Block struct {
	Index 			int
	Timestamp		string
	BPM					int
	Hash				string
	PrevHash		string
	Validator		string
}

// Blockchain as an array of validated Blocks
var Blockchain []Block
var tempBlocks	 []Block

// candidateBlocks handles incoming blocks for validation
var candidateBlocks = make(chan Block)

// announcements broadcasts winning validator to all nodes
var announcements = make(chan string)

var mutex = &sync.Mutex{}

// validators keep track of open validators and balances
var validators = make(map[string]int)

// standard blockchain functions -> hashing and calculate block hash
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	return calculateHash(record)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int, address string) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address

	return newBlock, nil
}

// check if block is valid
func isBlockValid(newBlock, oldBlock Block) bool {
	if newBlock.Index != oldBlock.Index + 1 {
		return false
	}

	if newBlock.PrevHash != oldBlock.Hash {
		return false 
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}
	return true
}

// validator functions
func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			msg := <-announcements
			io.WriteString(conn, msg)
		}
	}()

	// validator address
	var address string

	// allow user to allocate tokens for staking
	// more tokens -> higher chance of forgin a new block
	io.WriteString(conn, "Enter token balance:")
	scanBalance := bufio.NewScanner(conn)
	for scanBalance.Scan() {
		balance, err := strconv.Atoi(scanBalance.Text())
		if err != nil {
			log.Printf("%v not a number: %v", scanBalance.Text(), err)
			return
		}
		t := time.Now()
		address = calculateHash(t.String())
		validators[address] = balance
		fmt.Println(validators)
		break
	}

	io.WriteString(conn, "\nEnter a new BPM:")
	scanBPM := bufio.NewScanner(conn)

	go func() {
		for {
			// take in BPM from input and add it to blockchain
			for scanBPM.Scan() {
				bpm, err := strconv.Atoi(scanBPM.Text())
				// if someone enters invalid input remove him as a validator!
				if err != nil {
					log.Printf("%v not a number: %v", scanBPM.Text(), err)
					delete(validators, address)
					conn.Close()
				}

				mutex.Lock()
				oldLastIndex := Blockchain[len(Blockchain) - 1]
				mutex.Unlock()

				// create a new block considering to be forged
				newBlock, err := generateBlock(oldLastIndex, bpm, address)
				if err != nil {
					log.Println(err)
					continue
				}
				if isBlockValid(newBlock, oldLastIndex) {
					candidateBlocks <- newBlock
				}
				io.WriteString(conn, "\nEnter a new BPM:")

			}
		}
	}()

	// simulate receiving broadcast
	for {
		time.Sleep(time.Minute)
		mutex.Lock()
		output, err := json.Marshal(Blockchain)
		mutex.Unlock()
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(conn, string(output) + "\n")
	}
}

// pick a winner implementation, random selection with more weight to more staked tokens
func pickWinner() {
	time.Sleep(30 * time.Second)
	mutex.Lock()
	temp := tempBlocks
	mutex.Unlock()

	lotteryPool := []string{}
	if len(temp) > 0 {
		// slightly modified traditional PoS algorythm
		// from all validators who submitted a block, weight them by staked tokens amount
		// in traditional PoS validators can participate without submitting a block for forge
		OUTER:
			for _, block := range temp {
				// if already in lotter pool, skip
				for _,  node := range lotteryPool {
					if block.Validator == node {
						continue OUTER
					}
				}

				// lock list of validators to prevent data race
				mutex.Lock()
				setValidators := validators
				mutex.Unlock()

				// check if address is a validator and enter him as many times as tokens staked
				k, ok := setValidators[block.Validator]
				if ok {
					for i := 0; i < k; i++ {
						lotteryPool = append(lotteryPool, block.Validator)
					}
				}

			}

			// randomly pick a winner from lottery pool
			s := rand.NewSource(time.Now().Unix())
			r := rand.New(s)
			lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

			// add block of winner and let other nodes know
			for _, block := range temp {
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

// main func
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	//create genesis block
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateBlockHash(genesisBlock), "", ""}
	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	// start TCP server
	server, err := net.Listen("tcp", ":" + os.Getenv("ADDR"))
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	go func() {
		for candidate := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidate)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner()
		}
	}()

	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConn(conn)
	}

}





























