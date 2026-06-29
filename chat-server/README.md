# Chat Server

This is a chat server where clients can join to send messages to each other. Every client sends 
messages to all other clients. It's meant for local use only as the client Id's are based on ports.

It uses Go's concurrency features to easily spin up concurrent "routines" that handle 
different operations.

There are 3 main things going on:

- We handle incoming connections and set up a goroutine to listen for new messages arriving 
from the client.
- We spin up another goroutine per connection that handles writing messages to the client.
- We have an event-broker goroutine that processes the incoming messages, passes them to 
each of the clients' outgoing messages channels.

The flow is: 
  1. client writes message
  2. connection reads message and passes it to broker
  3. broker puts message in each of the clients "inbox" `client.outgoing`
  4. write connection sends message to client (goroutine for each client, messages sent concurrently)

## Graceful shutdown

On server shutdown we wait for all messages to be delivered before termination.

The flow is:
  1. Close listener to avoid new connections
  2. Close each of the clients outgoing channel
  3. Wait for all channels to drain before closing the connections
  4. Close connections
  5. Delete the client pool and exit

## Running the program

To run the server:

```sh
go run chat-server/server/main.go
```

And you can open a couple of clients in your terminal using netcat:

```sh
nc localhost 8080
```

If you want to test many connections, you can use the minimal `chat-server/client/main.go` script. 
You can change the amount of clients you want to create by modifying the first for loop inside main.

To run the client:

```sh
go run chat-server/client/main.go
```

