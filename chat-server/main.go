package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

// TCP Server listens and accepts incoming connections

// handleClientConnections
// goroutine per connection, loop to collect buffer message
// append message to channel

// broadcastMessages
// goroutine with loop over range of message channel and sends to all
// track connections

// In order to manage and keep track of connections I need some way to store that list

type Message struct {
	from string
	text string
}

var HOST = "127.0.0.1"
var PORT = "8080"

func main() {
	addr := net.JoinHostPort(HOST, PORT)
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		log.Fatal("Error starting server: ", err)
		return
	}

	defer listener.Close()

	log.Printf("Listening on %s", addr)

	messages := make(chan Message)

	go func() {
		for message := range messages {
			log.Printf("New message from %s: %s", message.from, message.text)
		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		go handleConnection(conn, messages)
	}

}

func handleConnection(conn net.Conn, messages chan Message) {

	defer func() {
		log.Printf("Closing connection with %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()

	id := strings.Split(conn.RemoteAddr().String(), ":")[1]
	log.Printf("Client connection: %s\n", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		message := Message{from: id, text: scanner.Text()}
		messages <- message
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error from %s: %v ", conn.RemoteAddr().String(), err)
		return
	}

}
