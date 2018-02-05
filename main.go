package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// A server struct to wrap the http server, and channels, waitgroups, etc
type Serve struct {
	server       *http.Server
	gracefulStop chan int
	wg           *sync.WaitGroup
	isStopping   bool
	passMap      map[int]string
}

// Return the ID of the given hashed password
func (server Serve) returnId(w http.ResponseWriter, pass string) {
	mapLen := len(server.passMap)
	server.passMap[mapLen+1] = pass
	s := strconv.Itoa(mapLen + 1)
	w.Write([]byte(s + "\n"))

}

// Send a signal to the waiting goroutine to shutdown
func (server Serve) shutdown(w http.ResponseWriter, req *http.Request) {
	server.gracefulStop <- 1
}

// The handler for /hash. Performs differenly for POST/GET
func (server Serve) hash(w http.ResponseWriter, req *http.Request) {
	// if the shutdown signal was sent, no more requests should be made
	if server.isStopping {
		w.Write([]byte("Server is shutting down, no more requests"))
		return
	}

	// the waitgroup for handling if their are still requests in process
	server.wg.Add(1)
	defer server.wg.Done()

	switch req.Method {
	case "POST":

		err := req.ParseForm()
		if err != nil {
			panic(err)
		}
		// Perform the hashing
		input := req.Form.Get("password")
		hmac512 := hmac.New(sha512.New, []byte("secret"))
		hmac512.Write([]byte(input))
		hashed_code := base64.StdEncoding.EncodeToString(hmac512.Sum(nil))

		server.returnId(w, hashed_code)
		// Keeping the socket open for 5 seconds. This may be wrong, but I do not know enough
		// about sockets and 'keep alve' to know for sure.
		time.Sleep(5 * time.Second)

		log.Println(hashed_code)
		w.Write([]byte(hashed_code))

	case "GET":
		id_str := req.URL.Path[len("/hash/"):]
		id, err := strconv.Atoi(id_str)
		if err != nil {
			log.Fatal("Could not get ID")
		}
		hash := server.passMap[id]
		w.Write([]byte(hash))

	default:
		log.Println("You didnt Post/GEt")
	}
}

func main() {
	s := Serve{
		server:       &http.Server{},
		gracefulStop: make(chan int),
		wg:           &sync.WaitGroup{},
		isStopping:   false,
		passMap:      make(map[int]string),
	}
	// Add the handler for /hash endpoint
	http.HandleFunc("/hash/", s.hash)
	http.HandleFunc("/shutdown", s.shutdown)

	// Create a HTTP server -- timeout set at 10 seconds for read/write
	s.server = &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// The go routine to handle shutdowns, if it reads from the channel, it waits for all
	// remaining requests to finish (s.wg.Wait()) and calls Go's shutdown (1.8 specific)
	go func() {
		<-s.gracefulStop
		log.Println("Gracefully Shutting down the Server")
		s.wg.Wait()
		s.server.Shutdown(nil)
	}()

	// Start Listening!
	log.Fatal(s.server.ListenAndServe())
}
