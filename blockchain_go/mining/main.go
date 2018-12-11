package main 

import(
	"crypto/sha256"
  "encoding/hex"
  "encoding/json"
  "fmt"
  "io"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"
  "sync"
  "time"

  "github.com/davecgh/go-spew/spew"
  "github.com/gorilla/mux"
  "github.com/joho/godotenv"
)

// add difficulty as a nuber of 0s leading the hash
const difficulty = 1


// add difficulty and nonce to block struct
type Block struct {
	Index				int
	Timestamp		string
	BPM					int
	Hash				string 
	PrevHash		string
	Difficulty 	int
	Nonce 			string
}

// blockchain array
var Blockchain []Block

// add new BPM as a message
type Message struct {
	BPM int
}

// we will use it to add 1 block at a time and no confilcts
var mutex = &sync.Mutex{}

// create and run a web server
func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on port: ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr: 					":" + httpAddr,
		Handler: 				mux,
		ReadTimeout: 		10 * time.Second,
		WriteTimeout: 	10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, 
																						payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	response, err := json.MarshalIndent(payload, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var m Message

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}

	defer r.Body.Close()

	// ensure only 1 block added at a time
	mutex.Lock()
	newBlock := generateBlock(Blockchain[len(Blockchain) - 1], m.BPM)
	mutex.Unlock()

	if isBlockValid(newBlock, Blockchain[len(Blockchain) - 1]) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

// basic blockchain functions
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index + 1 != newBlock.Index {
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

func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + 
						strconv.Itoa(block.BPM) + block.PrevHash + block.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// proof of work stuff
func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func generateBlock(oldBlock Block, BPM int) Block {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = difficulty

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		// store calculate hash and store it in a temp var
		tempHash := calculateHash(newBlock)
		if !isHashValid(tempHash, newBlock.Difficulty) {
			fmt.Println(tempHash, " do more work!")
			time.Sleep(time.Second)
			continue
		} else {
			fmt.Println(tempHash, " work done!")
			newBlock.Hash = tempHash
			break
		}
	}

	return newBlock
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		t := time.Now()
		genesisBlock := Block{}
		genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), "", 
																														difficulty, ""}
		spew.Dump(genesisBlock)

		mutex.Lock()
		Blockchain = append(Blockchain, genesisBlock)
		mutex.Unlock()
	}()
	log.Fatal(run())
}













