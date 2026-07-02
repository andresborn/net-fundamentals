# Chat Server

This is a demo chat server intended for local use only. It's purpose is to apprehend concurrency 
with goroutines, connection lifecycle management, graceful drain and signal handling.

It follows a **Server Push** pattern where: client subscribes/connects once then server pushes data 
thereafter without further client requests. Architecturally, seems to me, we're also applying 
multiplexing/demultiplexing and publish-subscribe patterns, at least conceptually. Our event broker 
is multiplexing multiple events from wide set of sources, streamlining them into a single operation 
and then, for messages at least, demultiplexing/fanning out and sending the messages to all 
connections. The publish-subscribe is being applied by default as clients connect to the "general" 
chat room on connection, there's just one "topic" in this case. Clients don't need to know of the 
existence of others, yet they receive their messages.

The main components of the application are:

- Three channels, this is how we communicate events between goroutines: `subscribe`, `unsubscribe`, 
`broadcast`.
- A map of the clients equipped with a mutex for safe access.
- An event broker; processes channels sequencially.
- The connection handler with independent goroutines for reading/writing from/to clients.

Message flow is: 
  1. Client sends message.
  2. Read goroutine reads message.
  3. Read goroutine passes message to `broadcast` channel.
  4. Event broker selects `broadcast` channel's incoming message and adds the message to each of 
  the clients "inbox" channel, `client.outgoing`.
  5. Every write goroutine sees new message in its `outgoing` channel and writes the message to 
  the connection.
  6. Clients receives message.

Subscribe flow is:
  1. TCP server is listening on `localhost:8080`, accepting incoming connections.
  2. Listening server accepts incoming connections, establishing a read goroutine and a write 
  goroutine, and sending the client to the `subscribe` channel.
  3. Event broker selects `subscribe` channel, receives client, locks the client map, and adds the 
  new one to it.
  4. Client is now subscribed and will receive messages from other clients.

Unsubscribe flow is:
  1. There's a client disconnection that returns an error, either from write or read (but most 
  likely from read).
  2. Goroutine exits, program flow continues, triggering unsubscribe event, sending the client to 
  the unsubscribe channel, and finally closing the connection.
  3. Event broker selects `unsubscribe` channel, receives client, locks the client map, removing the 
  client.
  4. We are no longer sending messages to this client.


## Graceful shutdown

I implemented graceful shutdown on the previous to last iteration but opted to remove it for now. 
I encountered a race condition when error handling the listener on shutdown.

Ideally, the shutdown would:
  1. Close listener to stop new connections.
  2. Close each of the clients' outgoing channels.
  3. Wait for all channels to drain/finish writing messages before closing client connections.
  4. Close client connections.
  5. Delete the client pool and exit.

We would introduce a wait group in order to wait for the write goroutines to finish. 

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

## Journal

I took a lot of turns with this project. Initially, I was on the right track, processing events 
sequentially through the event broker, but at some point, I think it was the complexity that the 
graceful shutdown introduced, I started trying some nonsensical things (you can see the commit 
histroy). After a while of being lost, I resorted to [this article](https://www.freecodecamp.org/news/how-to-build-a-production-grade-distributed-chatroom-in-go-full-handbook/) for guidance. It was 
very useful as it explained [*why*](https://www.freecodecamp.org/news/how-to-build-a-production-grade-distributed-chatroom-in-go-full-handbook/#:~:text=Why%20Use%20a%20Single%20Event%20Loop) we are building this this way and it cleared some of the issues 
I had introduced. Sequential vs parallel processing, race conditions, state management in a distributed 
environment and writing code that human minds can reason about were some of the things this guide 
made me reflect on.

## Upcoming

This project will be continued in [this repository](https://github.com/andresborn/chat-server-go). 

The plan is to add specific pub-sub topics and private channels, implement gRPC, learn how to 
breack the backlog, build a load balancer and deploy multiple replicas of the server.