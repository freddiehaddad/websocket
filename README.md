# Websocket Pub/Sub Demonstration

Minimal pub/sub client/server program demonstrating websocket communication.

Clients connect to the server using the websocket scheme `wss://` to `/ws` path
to initiate a connection. The server request handler enables the two-way
websocket communication.

After connecting, clients can begin sending messages to the server. Each message
is then broadcast to all clients (including the sender).

Each client connection requires an initial message to be sent specifying a
unique client name.

## Usage

The program works by launching a server and any number of clients.

### Server

Start the server with the default values or by using the `-host` argument to
specify an override in the following format `localhost:8080`.

`go run ./pkg/server/server.go [-host localhost:8080]`

### Clients

Once the server is running, start client connections with unique names. If you
do not specify a name, then `anonymous` will be used. However, multiple clients
using the same name will overwrite each other and only the last connection will
work as expected. Be sure to specify unique names for each connection.

`go run ./pkg/client/client.go -name anonymous [-host localhost:8080]`
