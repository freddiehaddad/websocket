// client.go represents connected clients.
//
// Clients send messages to the server and receive messages from the server.
// Each client has a unique name which is prefixed in the messages when sent
// or received.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	host string
	name string
)

// initialize the logger and flags
func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	host = *flag.String("host", "localhost:8080", "http server address and port")
	name = *flag.String("name", "anonymous", "client name")
}

// input validation for command line arguments
func checkFlags() bool {
	// Assuming valid input for assessment.
	return true
}

// incomingMessageHandler listens for messages from the server sent by this client or
// any other connected clients and logs them.
func incomingMessageHandler(wg *sync.WaitGroup, connection *websocket.Conn) {
	defer wg.Done()
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}
}

// outgoingMessageHandler generates messages every 5 seconds and sends them to the
// server to simulate a publish action.
func outgoingMessageHandler(wg *sync.WaitGroup, connection *websocket.Conn) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	defer wg.Done()

	tickCount := 1
	for range ticker.C {
		msg := fmt.Sprintf("%s: %d", name, tickCount)
		tickCount++
		err := connection.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	if !checkFlags() {
		log.Fatalln("Assuming valid input for assessment.")
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Attempt connection to server.
	// Could implement retries here incase server is temporarily inaccesable.
	u := url.URL{Scheme: "ws", Host: host, Path: "/ws"}
	log.Printf("client %s is connecting to %s", name, host)

	connection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer connection.Close()

	// Send client name
	err = connection.WriteMessage(websocket.TextMessage, []byte(name))
	if err != nil {
		log.Println("write:", err)
		return
	}

	// handler for incoming messages from server
	go incomingMessageHandler(wg, connection)

	// handler for outgoing messages from client
	go outgoingMessageHandler(wg, connection)

	// block until connection is closed
	wg.Wait()
	log.Println("client", name, "exiting")
}
