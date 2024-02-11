// server.go represents a http+websocket server.
//
// The default behavior is to listen on localhost:8080 at the /ws subdirectory
// initiating registration requests for new connections.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	host string
	ws   websocket.Upgrader
)

// initialize the logger and flags
func init() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	host = *flag.String("host", "localhost:8080", "http server address and port")
	ws = websocket.Upgrader{}
}

// input validation for command line arguments
func checkFlags() bool {
	// Assuming valid input for assessment.
	return true
}

// The router
type router struct {
	messages   chan []byte        // incoming client messages
	register   chan *client       // new client connection registration requests
	unregister chan *client       // existing client connection unregistration requests
	clients    map[string]*client // map of all connected clients (not strictly needed)
	mutex      sync.Mutex         // synchronization primitive for clients object updates
}

// clients represent a websocket connection to the server and are referenced by
// name.
type client struct {
	connection *websocket.Conn
	name       string
}

// handleWebSocketConnectionRequest upgrades http connections to use the websocket protocol
// and initiates registration of client connections for receiving and sending messages to and
// from the client.
func handleWebSocketConnectionRequest(router *router, w http.ResponseWriter, r *http.Request) {
	// upgrade to websocket protocol
	connection, err := ws.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer connection.Close()

	// get client name
	_, name, err := connection.ReadMessage()
	if err != nil {
		log.Println("read:", err)
		return
	}

	// register client
	client := newClient(name, connection)
	router.register <- client
	defer func() { router.unregister <- client }()

	// listen for messages
	for {
		_, message, err := connection.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		router.messages <- message
	}
}

// returns a router which acts as an in-memory container for all connections
func newRouter() *router {
	return &router{
		messages:   make(chan []byte),
		register:   make(chan *client),
		unregister: make(chan *client),
		clients:    make(map[string]*client),
	}
}

// returns a client connection
func newClient(name []byte, connection *websocket.Conn) *client {
	return &client{
		name:       string(name),
		connection: connection,
	}
}

// connectionHandler adds new websocket connections to the router.
func (router *router) connectionHandler() {
	for connection := range router.register {
		log.Println("registering client:", connection.name)
		// Could use a lock-free algorithm here avoiding the need for a mutex.
		router.mutex.Lock()
		// Check would be appropriate to avoid overwriting an existing connection.
		// However, in production, I wouldn't do this at all :)
		router.clients[connection.name] = connection
		router.mutex.Unlock()
	}
	log.Println("channel closed")
}

// disconnectionHandler removes websocket connections from the router.
func (router *router) disconnectionHandler() {
	for connection := range router.unregister {
		log.Println("unregistering client:", connection.name)
		// Could use a lock-free algorithm here and avoid the need for a mutex.
		router.mutex.Lock()
		delete(router.clients, connection.name)
		router.mutex.Unlock()
	}
	log.Println("channel closed")
}

// messageBroadcastHandler broadcasts incoming messages to connected clients
func (router *router) messageBroadcastHandler() {
	// process incoming messages
	for message := range router.messages {
		// Could use a lock-free algorithm here and avoid the need for a mutex.
		router.mutex.Lock()
		// broadcast to clients
		for _, client := range router.clients {
			// go routine could work here as well to parallelize broadcasting
			err := client.connection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("write:", client.name, err)
				// disconnect on write error
				// configuring timeouts is more appropriate
				router.unregister <- client
			}
		}
		router.mutex.Unlock()
	}
	log.Println("channel closed")
}

func main() {
	flag.Parse()
	if !checkFlags() {
		log.Fatalln("Assuming valid input for assessment.")
	}

	router := newRouter()

	// handler for broadcasting incoming messages to clients
	go router.messageBroadcastHandler()

	// handler for connections
	go router.connectionHandler()

	// handler for disconnections
	go router.disconnectionHandler()

	// setup the http listener on /ws
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocketConnectionRequest(router, w, r)
	})

	// all threads running, safe to start listening
	log.Println("Server listening on", host)
	log.Fatal(http.ListenAndServe(host, nil))
}
